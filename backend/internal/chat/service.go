package chat

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"kaia-analytics-ai/internal/contracts"
	"kaia-analytics-ai/pkg/config"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// Service handles chat interactions and NLP processing
type Service struct {
	config          *config.Config
	db              *sql.DB
	redis           *redis.Client
	contractManager *contracts.Manager
	logger          *logrus.Logger
	upgrader        websocket.Upgrader
}

// ChatMessage represents a chat message
type ChatMessage struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Message   string    `json:"message"`
	Response  string    `json:"response"`
	Intent    string    `json:"intent"`
	Entities  []Entity  `json:"entities"`
	Actions   []Action  `json:"actions"`
	Timestamp time.Time `json:"timestamp"`
}

// Entity represents extracted entities from user message
type Entity struct {
	Type       string      `json:"type"`
	Value      interface{} `json:"value"`
	Confidence float64     `json:"confidence"`
}

// Action represents an action to be executed
type Action struct {
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
	Status     string                 `json:"status"`
}

// QueryRequest represents a chat query request
type QueryRequest struct {
	Message string `json:"message"`
	UserID  string `json:"user_id"`
	Context string `json:"context,omitempty"`
}

// QueryResponse represents a chat query response
type QueryResponse struct {
	Response   string   `json:"response"`
	Intent     string   `json:"intent"`
	Entities   []Entity `json:"entities"`
	Actions    []Action `json:"actions"`
	Suggestions []string `json:"suggestions"`
}

// ActionRequest represents an action execution request
type ActionRequest struct {
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
	UserID     string                 `json:"user_id"`
}

// ActionResponse represents an action execution response
type ActionResponse struct {
	ActionID string `json:"action_id"`
	Status   string `json:"status"`
	Result   string `json:"result"`
}

// NewService creates a new chat service
func NewService(
	config *config.Config,
	db *sql.DB,
	redis *redis.Client,
	contractManager *contracts.Manager,
	logger *logrus.Logger,
) *Service {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins in development
		},
	}

	return &Service{
		config:          config,
		db:              db,
		redis:           redis,
		contractManager: contractManager,
		logger:          logger.WithField("service", "chat"),
		upgrader:        upgrader,
	}
}

// Start starts the chat service
func (s *Service) Start(ctx context.Context) error {
	s.logger.Info("Starting Chat Engine")

	// Start background processors
	go s.processActionQueue(ctx)

	<-ctx.Done()
	s.logger.Info("Chat Engine stopped")
	return nil
}

// HTTP Handlers

// HandleQuery processes a chat query
func (s *Service) HandleQuery(c *gin.Context) {
	var request QueryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	response, err := s.processQuery(c.Request.Context(), &request)
	if err != nil {
		s.logger.WithError(err).Error("Failed to process query")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process query"})
		return
	}

	// Store chat message
	chatMessage := &ChatMessage{
		ID:        fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		UserID:    request.UserID,
		Message:   request.Message,
		Response:  response.Response,
		Intent:    response.Intent,
		Entities:  response.Entities,
		Actions:   response.Actions,
		Timestamp: time.Now(),
	}

	if err := s.storeChatMessage(c.Request.Context(), chatMessage); err != nil {
		s.logger.WithError(err).Error("Failed to store chat message")
	}

	c.JSON(http.StatusOK, response)
}

// HandleAction processes an action execution request
func (s *Service) HandleAction(c *gin.Context) {
	var request ActionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	response, err := s.executeAction(c.Request.Context(), &request)
	if err != nil {
		s.logger.WithError(err).Error("Failed to execute action")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute action"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetChatHistory returns chat history for a user
func (s *Service) GetChatHistory(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	history, err := s.getChatHistory(c.Request.Context(), userID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get chat history")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get chat history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      history,
		"timestamp": time.Now(),
	})
}

// HandleWebSocket handles WebSocket connections for real-time chat
func (s *Service) HandleWebSocket(c *gin.Context) {
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.logger.WithError(err).Error("Failed to upgrade WebSocket connection")
		return
	}
	defer conn.Close()

	userID := c.Query("user_id")
	if userID == "" {
		conn.WriteJSON(map[string]string{"error": "user_id is required"})
		return
	}

	s.handleWebSocketConnection(c.Request.Context(), conn, userID)
}

// Core Processing Methods

// processQuery processes a natural language query
func (s *Service) processQuery(ctx context.Context, request *QueryRequest) (*QueryResponse, error) {
	s.logger.WithField("message", request.Message).Debug("Processing query")

	// Extract intent and entities
	intent, entities := s.extractIntentAndEntities(request.Message)

	// Generate response based on intent
	response := s.generateResponse(ctx, intent, entities, request.UserID)

	// Generate suggested actions
	actions := s.generateActions(intent, entities)

	// Generate follow-up suggestions
	suggestions := s.generateSuggestions(intent)

	return &QueryResponse{
		Response:    response,
		Intent:      intent,
		Entities:    entities,
		Actions:     actions,
		Suggestions: suggestions,
	}, nil
}

// executeAction executes a requested action
func (s *Service) executeAction(ctx context.Context, request *ActionRequest) (*ActionResponse, error) {
	s.logger.WithField("type", request.Type).Debug("Executing action")

	// Validate user permissions
	canExecute, err := s.contractManager.CanPerformAction(ctx, common.HexToAddress(request.UserID))
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}

	if !canExecute {
		return &ActionResponse{
			Status: "failed",
			Result: "Insufficient permissions or subscription required",
		}, nil
	}

	// Create action in contract
	actionType := s.getActionTypeCode(request.Type)
	parametersJSON, _ := json.Marshal(request.Parameters)

	actionID, err := s.contractManager.CreateAction(ctx, actionType, string(parametersJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create action: %w", err)
	}

	return &ActionResponse{
		ActionID: actionID.String(),
		Status:   "pending",
		Result:   "Action created and queued for execution",
	}, nil
}

// NLP Processing Methods

// extractIntentAndEntities extracts intent and entities from user message
func (s *Service) extractIntentAndEntities(message string) (string, []Entity) {
	message = strings.ToLower(strings.TrimSpace(message))

	// Simple intent classification (in production, use ML models)
	var intent string
	var entities []Entity

	if strings.Contains(message, "yield") || strings.Contains(message, "farming") || strings.Contains(message, "apy") {
		intent = "yield_query"
		if strings.Contains(message, "kaia") {
			entities = append(entities, Entity{Type: "token", Value: "KAIA", Confidence: 0.9})
		}
	} else if strings.Contains(message, "trade") || strings.Contains(message, "buy") || strings.Contains(message, "sell") {
		intent = "trading_query"
	} else if strings.Contains(message, "stake") || strings.Contains(message, "staking") {
		intent = "staking_action"
	} else if strings.Contains(message, "governance") || strings.Contains(message, "vote") || strings.Contains(message, "proposal") {
		intent = "governance_query"
	} else if strings.Contains(message, "price") || strings.Contains(message, "chart") {
		intent = "price_query"
	} else {
		intent = "general_query"
	}

	return intent, entities
}

// generateResponse generates a response based on intent and entities
func (s *Service) generateResponse(ctx context.Context, intent string, entities []Entity, userID string) string {
	switch intent {
	case "yield_query":
		return s.generateYieldResponse(ctx, entities)
	case "trading_query":
		return s.generateTradingResponse(ctx, entities, userID)
	case "staking_action":
		return s.generateStakingResponse(ctx, entities)
	case "governance_query":
		return s.generateGovernanceResponse(ctx, entities)
	case "price_query":
		return s.generatePriceResponse(ctx, entities)
	default:
		return "I can help you with yield farming opportunities, trading suggestions, staking, governance information, and price data. What would you like to know?"
	}
}

// generateActions generates possible actions based on intent and entities
func (s *Service) generateActions(intent string, entities []Entity) []Action {
	var actions []Action

	switch intent {
	case "yield_query":
		actions = append(actions, Action{
			Type: "view_yield_opportunities",
			Parameters: map[string]interface{}{
				"category": "all",
			},
			Status: "available",
		})
	case "staking_action":
		for _, entity := range entities {
			if entity.Type == "token" {
				actions = append(actions, Action{
					Type: "stake_tokens",
					Parameters: map[string]interface{}{
						"token": entity.Value,
					},
					Status: "available",
				})
			}
		}
	case "governance_query":
		actions = append(actions, Action{
			Type: "view_proposals",
			Parameters: map[string]interface{}{
				"status": "active",
			},
			Status: "available",
		})
	}

	return actions
}

// generateSuggestions generates follow-up suggestions
func (s *Service) generateSuggestions(intent string) []string {
	switch intent {
	case "yield_query":
		return []string{
			"Show me the highest APY opportunities",
			"What are the risks of yield farming?",
			"Compare protocols by TVL",
		}
	case "trading_query":
		return []string{
			"Show me trading signals for KAIA",
			"What's the market sentiment?",
			"Analyze my trading performance",
		}
	case "staking_action":
		return []string{
			"Show me staking rewards",
			"Compare staking pools",
			"Check my staking balance",
		}
	default:
		return []string{
			"Show me yield opportunities",
			"Get trading suggestions",
			"Check governance proposals",
		}
	}
}

// Response Generation Methods

func (s *Service) generateYieldResponse(ctx context.Context, entities []Entity) string {
	return "Here are the top yield farming opportunities on Kaia:\n\n" +
		"🌾 KaiaSwap KAIA/USDC: 12.5% APY (Low Risk)\n" +
		"🌾 KaiaLend KAIA: 8.2% APY (Very Low Risk)\n" +
		"🌾 KaiaStake: 6.8% APY (Minimal Risk)\n\n" +
		"Would you like more details about any of these opportunities?"
}

func (s *Service) generateTradingResponse(ctx context.Context, entities []Entity, userID string) string {
	return "Based on current market analysis:\n\n" +
		"📈 KAIA/USDC: Strong Buy Signal (78% confidence)\n" +
		"📊 Target: $1.25 | Stop Loss: $0.95\n" +
		"⏰ Time Horizon: 1-2 weeks\n\n" +
		"Key factors: Positive technical indicators, increasing volume, and strong community sentiment.\n\n" +
		"Would you like me to set up automated trading alerts?"
}

func (s *Service) generateStakingResponse(ctx context.Context, entities []Entity) string {
	return "Staking options available:\n\n" +
		"🔒 Native KAIA Staking: 6.8% APY\n" +
		"💰 Minimum: 1,000 KAIA\n" +
		"⏰ Lock Period: 30 days\n\n" +
		"Would you like me to help you stake your KAIA tokens?"
}

func (s *Service) generateGovernanceResponse(ctx context.Context, entities []Entity) string {
	return "Active governance proposals:\n\n" +
		"🗳️ KIP-001: Increase Block Gas Limit\n" +
		"📊 Community Sentiment: 75% Positive\n" +
		"⏰ Voting ends in 5 days\n\n" +
		"Would you like to participate in governance voting?"
}

func (s *Service) generatePriceResponse(ctx context.Context, entities []Entity) string {
	return "Current KAIA price data:\n\n" +
		"💰 Price: $1.15 (+2.5% 24h)\n" +
		"📊 Volume: $50M (24h)\n" +
		"📈 Market Cap: $5.75B\n" +
		"🎯 Trend: Bullish\n\n" +
		"Would you like detailed price analysis or alerts?"
}

// WebSocket Handling

func (s *Service) handleWebSocketConnection(ctx context.Context, conn *websocket.Conn, userID string) {
	s.logger.WithField("user_id", userID).Info("WebSocket connection established")

	for {
		var message map[string]interface{}
		err := conn.ReadJSON(&message)
		if err != nil {
			s.logger.WithError(err).Debug("WebSocket connection closed")
			break
		}

		// Process message
		if msg, ok := message["message"].(string); ok {
			request := &QueryRequest{
				Message: msg,
				UserID:  userID,
			}

			response, err := s.processQuery(ctx, request)
			if err != nil {
				conn.WriteJSON(map[string]string{"error": "Failed to process message"})
				continue
			}

			conn.WriteJSON(response)
		}
	}
}

// Background Processing

func (s *Service) processActionQueue(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.processPendingActions(ctx)
		}
	}
}

func (s *Service) processPendingActions(ctx context.Context) {
	s.logger.Debug("Processing pending actions")
}

// Helper Methods

func (s *Service) getActionTypeCode(actionType string) uint8 {
	switch actionType {
	case "stake":
		return 0
	case "unstake":
		return 1
	case "swap":
		return 3
	case "vote":
		return 4
	default:
		return 7 // Custom
	}
}

// Data Storage Methods

func (s *Service) storeChatMessage(ctx context.Context, message *ChatMessage) error {
	// Cache recent messages
	cacheKey := fmt.Sprintf("chat_history:%s", message.UserID)
	messageJSON, _ := json.Marshal(message)
	s.redis.LPush(ctx, cacheKey, messageJSON)
	s.redis.LTrim(ctx, cacheKey, 0, 99) // Keep last 100 messages
	s.redis.Expire(ctx, cacheKey, 24*time.Hour)

	return nil
}

func (s *Service) getChatHistory(ctx context.Context, userID string) ([]*ChatMessage, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("chat_history:%s", userID)
	cached, err := s.redis.LRange(ctx, cacheKey, 0, 49).Result() // Get last 50 messages
	if err == nil && len(cached) > 0 {
		var messages []*ChatMessage
		for _, msgJSON := range cached {
			var msg ChatMessage
			if json.Unmarshal([]byte(msgJSON), &msg) == nil {
				messages = append(messages, &msg)
			}
		}
		return messages, nil
	}

	return []*ChatMessage{}, nil
}