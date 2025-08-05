package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"./services"
)

// App represents the main application
type App struct {
	router          *gin.Engine
	ethClient       *ethclient.Client
	logger          *logrus.Logger
	analyticsEngine *services.AnalyticsEngine
	dataCollector   *services.DataCollector
	chatEngine      *services.ChatEngine
}

// Config holds application configuration
type Config struct {
	Port        string
	EthNodeURL  string
	Environment string
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
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

	// Initialize services
	analyticsEngine, err := services.NewAnalyticsEngine(ethClient)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize analytics engine")
	}
	defer analyticsEngine.Close()

	dataCollector := services.NewDataCollector(ethClient)
	chatEngine := services.NewChatEngine(ethClient, analyticsEngine, dataCollector)

	// Initialize application
	app := &App{
		router:          gin.New(),
		ethClient:       ethClient,
		logger:          logger,
		analyticsEngine: analyticsEngine,
		dataCollector:   dataCollector,
		chatEngine:      chatEngine,
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
		v1.GET("/contract/:address/info", a.getContractInfo)
		
		// Analytics endpoints
		v1.POST("/analytics/yield", a.getYieldOpportunities)
		v1.POST("/analytics/trading-suggestions", a.getTradingSuggestions)
		v1.POST("/analytics/portfolio", a.getPortfolioAnalysis)
		v1.POST("/analytics/governance", a.getGovernanceSentiment)
		v1.POST("/analytics/risk-assessment", a.getRiskAssessment)
		
		// Data collection endpoints
		v1.GET("/data/market", a.getMarketData)
		v1.GET("/data/protocols", a.getProtocolData)
		v1.GET("/data/gas", a.getGasData)
		v1.GET("/data/blockchain", a.getBlockchainData)
		v1.GET("/data/historical/:start/:end", a.getHistoricalData)
		
		// Chat endpoints
		v1.POST("/chat/message", a.processChatMessage)
		v1.GET("/chat/ws", a.handleWebSocket)
		v1.GET("/chat/metrics", a.getChatMetrics)
		
		// Service metrics
		v1.GET("/metrics/analytics", a.getAnalyticsMetrics)
		v1.GET("/metrics/data", a.getDataMetrics)
	}

	// WebSocket endpoint
	a.router.GET("/ws", a.handleWebSocket)
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

// Analytics endpoints
func (a *App) getYieldOpportunities(c *gin.Context) {
	var request struct {
		UserAddress string                 `json:"user_address"`
		Parameters  map[string]interface{} `json:"parameters"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := a.analyticsEngine.ProcessAnalyticsTask(c.Request.Context(), "yield_analysis", request.Parameters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (a *App) getTradingSuggestions(c *gin.Context) {
	var request struct {
		UserAddress string                 `json:"user_address"`
		Parameters  map[string]interface{} `json:"parameters"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := a.analyticsEngine.ProcessAnalyticsTask(c.Request.Context(), "trading_suggestions", request.Parameters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (a *App) getPortfolioAnalysis(c *gin.Context) {
	var request struct {
		UserAddress string                 `json:"user_address"`
		Parameters  map[string]interface{} `json:"parameters"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := a.analyticsEngine.ProcessAnalyticsTask(c.Request.Context(), "portfolio_optimization", request.Parameters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (a *App) getGovernanceSentiment(c *gin.Context) {
	var request struct {
		UserAddress string                 `json:"user_address"`
		Parameters  map[string]interface{} `json:"parameters"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := a.analyticsEngine.ProcessAnalyticsTask(c.Request.Context(), "governance_sentiment", request.Parameters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (a *App) getRiskAssessment(c *gin.Context) {
	var request struct {
		UserAddress string                 `json:"user_address"`
		Parameters  map[string]interface{} `json:"parameters"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := a.analyticsEngine.ProcessAnalyticsTask(c.Request.Context(), "risk_assessment", request.Parameters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Data collection endpoints
func (a *App) getMarketData(c *gin.Context) {
	symbols := c.QueryArray("symbols")
	if len(symbols) == 0 {
		symbols = []string{"ETH", "USDC", "DAI"}
	}

	data, err := a.dataCollector.CollectMarketData(c.Request.Context(), symbols)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

func (a *App) getProtocolData(c *gin.Context) {
	data, err := a.dataCollector.CollectProtocolData(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

func (a *App) getGasData(c *gin.Context) {
	data, err := a.dataCollector.CollectGasData(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

func (a *App) getBlockchainData(c *gin.Context) {
	data, err := a.dataCollector.CollectBlockchainData(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

func (a *App) getHistoricalData(c *gin.Context) {
	startBlock := c.Param("start")
	endBlock := c.Param("end")
	
	// Parse block numbers (simplified)
	start := uint64(0)
	end := uint64(100)
	
	data, err := a.dataCollector.CollectHistoricalData(c.Request.Context(), start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

// Chat endpoints
func (a *App) processChatMessage(c *gin.Context) {
	var message services.ChatMessage
	if err := c.ShouldBindJSON(&message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := a.chatEngine.ProcessMessage(c.Request.Context(), &message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (a *App) handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		a.logger.WithError(err).Error("Failed to upgrade connection to WebSocket")
		return
	}
	defer conn.Close()

	// Register connection
	userID := c.Query("user_id")
	if userID == "" {
		userID = "anonymous"
	}
	a.chatEngine.RegisterConnection(userID, conn)
	defer a.chatEngine.UnregisterConnection(userID)

	a.logger.WithField("user_id", userID).Info("WebSocket connection established")

	for {
		// Read message
		var message services.ChatMessage
		err := conn.ReadJSON(&message)
		if err != nil {
			a.logger.WithError(err).Info("WebSocket connection closed")
			break
		}

		// Process message
		response, err := a.chatEngine.ProcessMessage(c.Request.Context(), &message)
		if err != nil {
			a.logger.WithError(err).Error("Failed to process chat message")
			continue
		}

		// Send response
		err = conn.WriteJSON(response)
		if err != nil {
			a.logger.WithError(err).Error("Failed to send WebSocket response")
			break
		}
	}
}

func (a *App) getChatMetrics(c *gin.Context) {
	metrics := a.chatEngine.GetChatMetrics()
	c.JSON(http.StatusOK, metrics)
}

// Metrics endpoints
func (a *App) getAnalyticsMetrics(c *gin.Context) {
	metrics := a.analyticsEngine.GetAnalyticsMetrics()
	c.JSON(http.StatusOK, metrics)
}

func (a *App) getDataMetrics(c *gin.Context) {
	metrics := a.dataCollector.GetDataMetrics()
	c.JSON(http.StatusOK, metrics)
}

// Existing endpoints (keeping for backward compatibility)
func (a *App) healthCheck(c *gin.Context) {
	// Check Ethereum connection
	_, err := a.ethClient.BlockNumber(c.Request.Context())
	ethStatus := "connected"
	if err != nil {
		ethStatus = "disconnected"
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"timestamp": time.Now().Unix(),
		"ethereum": ethStatus,
		"services": map[string]string{
			"analytics_engine": "running",
			"data_collector":   "running",
			"chat_engine":      "running",
		},
	})
}

func (a *App) getBlockByNumber(c *gin.Context) {
	blockNumber := c.Param("number")
	
	var blockNum *big.Int
	if blockNumber == "latest" {
		blockNum = nil
	} else {
		blockNum = new(big.Int)
		blockNum.SetString(blockNumber, 10)
	}

	block, err := a.ethClient.BlockByNumber(c.Request.Context(), blockNum)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"number": block.NumberU64(),
		"hash": block.Hash().Hex(),
		"timestamp": block.Time(),
		"transactions": len(block.Transactions()),
		"gas_used": block.GasUsed(),
		"gas_limit": block.GasLimit(),
	})
}

func (a *App) getTransactionByHash(c *gin.Context) {
	txHash := c.Param("hash")
	
	tx, isPending, err := a.ethClient.TransactionByHash(c.Request.Context(), common.HexToHash(txHash))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	receipt, err := a.ethClient.TransactionReceipt(c.Request.Context(), common.HexToHash(txHash))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hash": tx.Hash().Hex(),
		"from": receipt.From.Hex(),
		"to": receipt.To.Hex(),
		"value": tx.Value().String(),
		"gas_used": receipt.GasUsed,
		"status": receipt.Status,
		"is_pending": isPending,
	})
}

func (a *App) getAddressBalance(c *gin.Context) {
	address := c.Param("address")
	
	balance, err := a.ethClient.BalanceAt(c.Request.Context(), common.HexToAddress(address), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"address": address,
		"balance": balance.String(),
		"balance_eth": new(big.Float).Quo(new(big.Float).SetInt(balance), new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))),
	})
}

func (a *App) getNetworkStats(c *gin.Context) {
	// Get latest block
	header, err := a.ethClient.HeaderByNumber(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get gas price
	gasPrice, err := a.ethClient.SuggestGasPrice(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"latest_block": header.Number.Uint64(),
		"gas_price": gasPrice.String(),
		"difficulty": header.Difficulty.String(),
		"timestamp": time.Now().Unix(),
	})
}

func (a *App) getContractInfo(c *gin.Context) {
	address := c.Param("address")
	
	// Get contract code
	code, err := a.ethClient.CodeAt(c.Request.Context(), common.HexToAddress(address), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	isContract := len(code) > 0

	c.JSON(http.StatusOK, gin.H{
		"address": address,
		"is_contract": isContract,
		"code_size": len(code),
	})
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}