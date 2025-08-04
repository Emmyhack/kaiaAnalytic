package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"kaia-analytics-ai/internal/config"
	"kaia-analytics-ai/internal/contracts"
)

// Engine handles chat functionality and query processing
type Engine struct {
	config           *config.Config
	blockchainClient *contracts.BlockchainClient
	upgrader         websocket.Upgrader
	connections      map[*websocket.Conn]bool
	connectionsMutex sync.RWMutex
	stopChan         chan struct{}
}

// ChatMessage represents a chat message
type ChatMessage struct {
	Type      string      `json:"type"`      // "query", "response", "action", "error"
	Content   string      `json:"content"`
	UserID    string      `json:"userId,omitempty"`
	Timestamp int64       `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"`
}

// QueryResponse represents a response to a user query
type QueryResponse struct {
	Answer     string                 `json:"answer"`
	Confidence float64                `json:"confidence"`
	Actions    []SuggestedAction      `json:"actions,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Timestamp  int64                  `json:"timestamp"`
}

// SuggestedAction represents a suggested on-chain action
type SuggestedAction struct {
	Type        string                 `json:"type"`        // "stake", "vote", "swap", "yield_farm"
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Confidence  float64                `json:"confidence"`
}

// NewEngine creates a new chat engine
func NewEngine(cfg *config.Config, bc *contracts.BlockchainClient) *Engine {
	engine := &Engine{
		config:           cfg,
		blockchainClient: bc,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // In production, implement proper origin checking
			},
		},
		connections: make(map[*websocket.Conn]bool),
		stopChan:    make(chan struct{}),
	}

	return engine
}

// Start starts the chat engine
func (e *Engine) Start() {
	logrus.Info("Starting chat engine")
	
	// Start connection cleanup
	go e.cleanupConnections()
}

// Stop stops the chat engine
func (e *Engine) Stop() {
	logrus.Info("Stopping chat engine")
	close(e.stopChan)
	
	// Close all connections
	e.connectionsMutex.Lock()
	for conn := range e.connections {
		conn.Close()
	}
	e.connectionsMutex.Unlock()
}

// cleanupConnections periodically cleans up dead connections
func (e *Engine) cleanupConnections() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			e.connectionsMutex.Lock()
			for conn := range e.connections {
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					delete(e.connections, conn)
					conn.Close()
				}
			}
			e.connectionsMutex.Unlock()
		case <-e.stopChan:
			return
		}
	}
}

// HTTP Handlers

// HandleQuery handles HTTP query requests
func (e *Engine) HandleQuery(c *gin.Context) {
	var req struct {
		Query  string `json:"query" binding:"required"`
		UserID string `json:"userId,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// Process query
	response, err := e.processQuery(req.Query, req.UserID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to process query"})
		return
	}

	c.JSON(200, response)
}

// HandleWebSocket handles WebSocket connections for real-time chat
func (e *Engine) HandleWebSocket(c *gin.Context) {
	conn, err := e.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logrus.Errorf("Failed to upgrade connection: %v", err)
		return
	}

	// Register connection
	e.connectionsMutex.Lock()
	e.connections[conn] = true
	e.connectionsMutex.Unlock()

	// Send welcome message
	welcomeMsg := ChatMessage{
		Type:      "response",
		Content:   "Welcome to KaiaAnalyticsAI! Ask me anything about analytics, trading, or blockchain actions.",
		Timestamp: time.Now().Unix(),
	}
	conn.WriteJSON(welcomeMsg)

	// Handle messages
	go e.handleWebSocketMessages(conn)
}

// handleWebSocketMessages handles incoming WebSocket messages
func (e *Engine) handleWebSocketMessages(conn *websocket.Conn) {
	defer func() {
		e.connectionsMutex.Lock()
		delete(e.connections, conn)
		e.connectionsMutex.Unlock()
		conn.Close()
	}()

	for {
		select {
		case <-e.stopChan:
			return
		default:
			// Read message
			var msg ChatMessage
			err := conn.ReadJSON(&msg)
			if err != nil {
				logrus.Debugf("WebSocket read error: %v", err)
				return
			}

			// Process message
			response, err := e.processQuery(msg.Content, msg.UserID)
			if err != nil {
				errorMsg := ChatMessage{
					Type:      "error",
					Content:   "Failed to process query",
					Timestamp: time.Now().Unix(),
				}
				conn.WriteJSON(errorMsg)
				continue
			}

			// Send response
			responseMsg := ChatMessage{
				Type:      "response",
				Content:   response.Answer,
				Timestamp: time.Now().Unix(),
				Data:      response,
			}
			conn.WriteJSON(responseMsg)
		}
	}
}

// processQuery processes a user query and returns a response
func (e *Engine) processQuery(query, userID string) (*QueryResponse, error) {
	query = strings.ToLower(strings.TrimSpace(query))
	
	// Simple keyword-based query processing
	// In production, use proper NLP/LLM integration
	
	response := &QueryResponse{
		Timestamp: time.Now().Unix(),
	}

	// Check for yield-related queries
	if strings.Contains(query, "yield") || strings.Contains(query, "apy") || strings.Contains(query, "farming") {
		return e.handleYieldQuery(query, userID)
	}

	// Check for trading-related queries
	if strings.Contains(query, "trade") || strings.Contains(query, "buy") || strings.Contains(query, "sell") {
		return e.handleTradingQuery(query, userID)
	}

	// Check for governance-related queries
	if strings.Contains(query, "vote") || strings.Contains(query, "governance") || strings.Contains(query, "proposal") {
		return e.handleGovernanceQuery(query, userID)
	}

	// Check for staking-related queries
	if strings.Contains(query, "stake") || strings.Contains(query, "lock") {
		return e.handleStakingQuery(query, userID)
	}

	// Check for general analytics queries
	if strings.Contains(query, "volume") || strings.Contains(query, "gas") || strings.Contains(query, "price") {
		return e.handleAnalyticsQuery(query, userID)
	}

	// Default response
	response.Answer = "I can help you with yield farming opportunities, trading suggestions, governance voting, staking, and general analytics. What would you like to know?"
	response.Confidence = 0.8

	return response, nil
}

// handleYieldQuery handles yield farming related queries
func (e *Engine) handleYieldQuery(query, userID string) (*QueryResponse, error) {
	response := &QueryResponse{
		Answer:     "Here are the current best yield opportunities on Kaia:",
		Confidence: 0.9,
		Data: map[string]interface{}{
			"opportunities": []map[string]interface{}{
				{
					"pool":           "KAIA-USDC LP",
					"apy":            12.5,
					"tvl":            1000000.0,
					"risk_score":     0.3,
					"recommendation": "High yield with moderate risk",
				},
				{
					"pool":           "KAIA-ETH LP",
					"apy":            8.2,
					"tvl":            2500000.0,
					"risk_score":     0.2,
					"recommendation": "Stable yield with low risk",
				},
			},
		},
		Actions: []SuggestedAction{
			{
				Type:        "yield_farm",
				Description: "Deposit into KAIA-USDC LP pool",
				Parameters: map[string]interface{}{
					"pool":      "0x1234567890123456789012345678901234567890",
					"amount":    "1000",
					"lockPeriod": "30",
				},
				Confidence: 0.85,
			},
		},
	}

	return response, nil
}

// handleTradingQuery handles trading related queries
func (e *Engine) handleTradingQuery(query, userID string) (*QueryResponse, error) {
	response := &QueryResponse{
		Answer:     "Based on current market analysis, here are my trading suggestions:",
		Confidence: 0.8,
		Data: map[string]interface{}{
			"suggestions": []map[string]interface{}{
				{
					"token_pair":   "KAIA/USDC",
					"action":       "buy",
					"confidence":   0.75,
					"price":        1.25,
					"target_price": 1.40,
					"stop_loss":    1.15,
					"reasoning":    "Strong technical indicators and positive sentiment",
				},
			},
		},
		Actions: []SuggestedAction{
			{
				Type:        "swap",
				Description: "Buy KAIA with USDC",
				Parameters: map[string]interface{}{
					"tokenIn":     "0x1234567890123456789012345678901234567890", // USDC
					"tokenOut":    "0x2345678901234567890123456789012345678901", // KAIA
					"amountIn":    "1000",
					"minAmountOut": "800",
				},
				Confidence: 0.75,
			},
		},
	}

	return response, nil
}

// handleGovernanceQuery handles governance related queries
func (e *Engine) handleGovernanceQuery(query, userID string) (*QueryResponse, error) {
	response := &QueryResponse{
		Answer:     "Current governance proposals and sentiment:",
		Confidence: 0.85,
		Data: map[string]interface{}{
			"proposals": []map[string]interface{}{
				{
					"id":            1,
					"title":         "Increase Protocol Fee",
					"sentiment":     0.6,
					"support_votes": 1500,
					"against_votes": 500,
					"status":        "Active",
				},
				{
					"id":            2,
					"title":         "Add New Token Support",
					"sentiment":     0.8,
					"support_votes": 1800,
					"against_votes": 200,
					"status":        "Active",
				},
			},
		},
		Actions: []SuggestedAction{
			{
				Type:        "vote",
				Description: "Vote on Proposal #1",
				Parameters: map[string]interface{}{
					"proposalId": 1,
					"support":    true,
					"weight":     "100",
				},
				Confidence: 0.7,
			},
		},
	}

	return response, nil
}

// handleStakingQuery handles staking related queries
func (e *Engine) handleStakingQuery(query, userID string) (*QueryResponse, error) {
	response := &QueryResponse{
		Answer:     "Staking options and current rates:",
		Confidence: 0.9,
		Data: map[string]interface{}{
			"staking_options": []map[string]interface{}{
				{
					"pool":        "KAIA Staking Pool",
					"apy":         8.5,
					"lock_period": "30 days",
					"min_amount":  "100",
				},
			},
		},
		Actions: []SuggestedAction{
			{
				Type:        "stake",
				Description: "Stake KAIA tokens",
				Parameters: map[string]interface{}{
					"token":      "0x2345678901234567890123456789012345678901", // KAIA
					"amount":     "1000",
					"lockPeriod": "30",
				},
				Confidence: 0.8,
			},
		},
	}

	return response, nil
}

// handleAnalyticsQuery handles general analytics queries
func (e *Engine) handleAnalyticsQuery(query, userID string) (*QueryResponse, error) {
	response := &QueryResponse{
		Answer:     "Current blockchain analytics:",
		Confidence: 0.85,
		Data: map[string]interface{}{
			"transaction_volume": map[string]interface{}{
				"daily":   1800000.0,
				"weekly":  12000000.0,
				"trend":   "increasing",
			},
			"gas_price": map[string]interface{}{
				"current": 25.0,
				"trend":   "stable",
				"prediction": 26.0,
			},
			"active_addresses": 15000,
		},
	}

	return response, nil
}

// executeAction executes an on-chain action
func (e *Engine) executeAction(action SuggestedAction, userID string) error {
	// Convert action to blockchain transaction
	actionData, err := json.Marshal(action.Parameters)
	if err != nil {
		return err
	}

	// Create action on blockchain
	err = e.blockchainClient.CreateAction(action.Type, actionData)
	if err != nil {
		return fmt.Errorf("failed to execute action: %v", err)
	}

	logrus.Infof("Executed action %s for user %s", action.Type, userID)
	return nil
}

// BroadcastMessage broadcasts a message to all connected clients
func (e *Engine) BroadcastMessage(msg ChatMessage) {
	e.connectionsMutex.RLock()
	defer e.connectionsMutex.RUnlock()

	for conn := range e.connections {
		err := conn.WriteJSON(msg)
		if err != nil {
			logrus.Debugf("Failed to send message to client: %v", err)
		}
	}
}

// GetConnectionCount returns the number of active connections
func (e *Engine) GetConnectionCount() int {
	e.connectionsMutex.RLock()
	defer e.connectionsMutex.RUnlock()
	return len(e.connections)
}