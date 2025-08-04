package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"kaia-analytics-ai/internal/analytics"
	"kaia-analytics-ai/internal/chat"
	"kaia-analytics-ai/internal/collector"
	"kaia-analytics-ai/internal/config"
	"kaia-analytics-ai/internal/contracts"
	"kaia-analytics-ai/internal/middleware"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.Load()

	// Set up logging
	logrus.SetLevel(cfg.LogLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	// Initialize blockchain client
	blockchainClient, err := contracts.NewBlockchainClient(cfg)
	if err != nil {
		logrus.Fatalf("Failed to initialize blockchain client: %v", err)
	}

	// Initialize services
	analyticsEngine := analytics.NewEngine(cfg, blockchainClient)
	dataCollector := collector.NewCollector(cfg, blockchainClient)
	chatEngine := chat.NewEngine(cfg, blockchainClient)

	// Set up Gin router
	router := gin.Default()

	// Add middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())

	// API routes
	api := router.Group("/api/v1")
	{
		// Analytics routes
		analyticsGroup := api.Group("/analytics")
		{
			analyticsGroup.GET("/yield", analyticsEngine.GetYieldOpportunities)
			analyticsGroup.GET("/governance", analyticsEngine.GetGovernanceSentiment)
			analyticsGroup.GET("/trading", analyticsEngine.GetTradingSuggestions)
			analyticsGroup.GET("/volume", analyticsEngine.GetTransactionVolume)
			analyticsGroup.GET("/gas", analyticsEngine.GetGasTrends)
		}

		// Data collection routes
		collectorGroup := api.Group("/collector")
		{
			collectorGroup.GET("/blockchain", dataCollector.GetBlockchainData)
			collectorGroup.GET("/market", dataCollector.GetMarketData)
			collectorGroup.GET("/historical", dataCollector.GetHistoricalData)
		}

		// Chat routes
		chatGroup := api.Group("/chat")
		{
			chatGroup.POST("/query", chatEngine.HandleQuery)
			chatGroup.GET("/ws", chatEngine.HandleWebSocket)
		}

		// Subscription routes
		subscriptionGroup := api.Group("/subscription")
		{
			subscriptionGroup.GET("/plans", blockchainClient.GetSubscriptionPlans)
			subscriptionGroup.POST("/purchase", blockchainClient.PurchaseSubscription)
			subscriptionGroup.GET("/status/:address", blockchainClient.GetSubscriptionStatus)
		}

		// Health check
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":    "healthy",
				"timestamp": time.Now().Unix(),
				"version":   "1.0.0",
			})
		})
	}

	// Start services
	go analyticsEngine.Start()
	go dataCollector.Start()
	go chatEngine.Start()

	// Create HTTP server
	server := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logrus.Infof("Starting server on %s", cfg.ServerAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop services
	analyticsEngine.Stop()
	dataCollector.Stop()
	chatEngine.Stop()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		logrus.Fatalf("Server forced to shutdown: %v", err)
	}

	logrus.Info("Server exited")
}