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

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// App represents the main application
type App struct {
	router    *gin.Engine
	ethClient *ethclient.Client
	logger    *logrus.Logger
}

// Config holds application configuration
type Config struct {
	Port        string
	EthNodeURL  string
	Environment string
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default configuration")
	}

	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	
	// Set log level based on environment
	if os.Getenv("ENVIRONMENT") == "development" {
		logger.SetLevel(logrus.DebugLevel)
		gin.SetMode(gin.DebugMode)
	} else {
		logger.SetLevel(logrus.InfoLevel)
		gin.SetMode(gin.ReleaseMode)
	}

	// Load configuration
	config := &Config{
		Port:        getEnvOrDefault("PORT", "8080"),
		EthNodeURL:  getEnvOrDefault("ETH_NODE_URL", "https://mainnet.infura.io/v3/your-project-id"),
		Environment: getEnvOrDefault("ENVIRONMENT", "development"),
	}

	// Initialize Ethereum client
	ethClient, err := ethclient.Dial(config.EthNodeURL)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to Ethereum client")
	}
	defer ethClient.Close()

	// Initialize application
	app := &App{
		router:    gin.New(),
		ethClient: ethClient,
		logger:    logger,
	}

	// Setup middleware
	app.setupMiddleware()

	// Setup routes
	app.setupRoutes()

	// Start server
	app.start(config.Port)
}

func (a *App) setupMiddleware() {
	// Add gin logger middleware
	a.router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			return fmt.Sprintf("[%s] %s %s %d %s %s\n",
				param.TimeStamp.Format("2006-01-02 15:04:05"),
				param.Method,
				param.Path,
				param.StatusCode,
				param.Latency,
				param.ClientIP,
			)
		},
	}))

	// Recovery middleware
	a.router.Use(gin.Recovery())

	// CORS middleware
	a.router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})
}

func (a *App) setupRoutes() {
	// Health check endpoint
	a.router.GET("/health", a.healthCheck)

	// API v1 routes
	v1 := a.router.Group("/api/v1")
	{
		// Blockchain analytics endpoints
		v1.GET("/block/:number", a.getBlockByNumber)
		v1.GET("/transaction/:hash", a.getTransactionByHash)
		v1.GET("/address/:address/balance", a.getAddressBalance)
		v1.GET("/network/stats", a.getNetworkStats)
		
		// Contract analytics endpoints
		v1.GET("/contract/:address/info", a.getContractInfo)
	}
}

func (a *App) start(port string) {
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: a.router,
	}

	// Start server in a goroutine
	go func() {
		a.logger.WithField("port", port).Info("Starting server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	a.logger.Info("Shutting down server...")

	// Give outstanding requests 5 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		a.logger.WithError(err).Error("Server forced to shutdown")
	}

	a.logger.Info("Server exited")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}