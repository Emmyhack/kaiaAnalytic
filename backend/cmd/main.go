package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kaia-analytics-ai/internal/analytics"
	"kaia-analytics-ai/internal/chat"
	"kaia-analytics-ai/internal/collector"
	"kaia-analytics-ai/internal/contracts"
	"kaia-analytics-ai/pkg/config"
	"kaia-analytics-ai/pkg/database"
	"kaia-analytics-ai/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize logger
	logger := logger.NewLogger()
	logger.Info("Starting KaiaAnalyticsAI Backend Services")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.WithError(err).Fatal("Failed to load configuration")
	}

	// Initialize database connections
	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Initialize Redis connection
	redisClient, err := database.NewRedisConnection(cfg.RedisURL)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to Redis")
	}
	defer redisClient.Close()

	// Initialize blockchain contracts
	contractManager, err := contracts.NewManager(cfg.KaiaRPCURL, cfg.ContractAddresses)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize contract manager")
	}

	// Initialize services
	dataCollector := collector.NewService(cfg, db, redisClient, contractManager, logger)
	analyticsEngine := analytics.NewService(cfg, db, redisClient, contractManager, logger)
	chatEngine := chat.NewService(cfg, db, redisClient, contractManager, logger)

	// Start background services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start data collector
	go func() {
		if err := dataCollector.Start(ctx); err != nil {
			logger.WithError(err).Error("Data collector failed")
		}
	}()

	// Start analytics engine
	go func() {
		if err := analyticsEngine.Start(ctx); err != nil {
			logger.WithError(err).Error("Analytics engine failed")
		}
	}()

	// Start chat engine
	go func() {
		if err := chatEngine.Start(ctx); err != nil {
			logger.WithError(err).Error("Chat engine failed")
		}
	}()

	// Initialize HTTP server
	router := setupRouter(cfg, dataCollector, analyticsEngine, chatEngine, logger)
	
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	// Start HTTP server
	go func() {
		logger.WithField("port", cfg.Port).Info("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Cancel background services
	cancel()

	// Shutdown HTTP server
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	}

	logger.Info("Server exited")
}

func setupRouter(
	cfg *config.Config,
	dataCollector *collector.Service,
	analyticsEngine *analytics.Service,
	chatEngine *chat.Service,
	logger *logrus.Logger,
) *gin.Engine {
	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middleware
	router.Use(gin.LoggerWithWriter(logger.Writer()))
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"version":   "1.0.0",
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Analytics routes
		analytics := api.Group("/analytics")
		{
			analytics.GET("/yield", analyticsEngine.GetYieldOpportunities)
			analytics.GET("/trading-suggestions", analyticsEngine.GetTradingSuggestions)
			analytics.GET("/governance", analyticsEngine.GetGovernanceData)
			analytics.GET("/market-trends", analyticsEngine.GetMarketTrends)
			analytics.POST("/custom-query", analyticsEngine.HandleCustomQuery)
		}

		// Data routes
		data := api.Group("/data")
		{
			data.GET("/transactions", dataCollector.GetTransactionData)
			data.GET("/blocks", dataCollector.GetBlockData)
			data.GET("/tokens", dataCollector.GetTokenData)
			data.GET("/protocols", dataCollector.GetProtocolData)
		}

		// Chat routes
		chat := api.Group("/chat")
		{
			chat.POST("/query", chatEngine.HandleQuery)
			chat.POST("/action", chatEngine.HandleAction)
			chat.GET("/history", chatEngine.GetChatHistory)
			chat.GET("/ws", chatEngine.HandleWebSocket)
		}

		// User routes
		user := api.Group("/user")
		{
			user.GET("/subscription", getUserSubscription)
			user.GET("/usage", getUserUsage)
			user.GET("/analytics-history", getUserAnalyticsHistory)
		}
	}

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Placeholder handlers for user routes
func getUserSubscription(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get user subscription - to be implemented"})
}

func getUserUsage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get user usage - to be implemented"})
}

func getUserAnalyticsHistory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get user analytics history - to be implemented"})
}