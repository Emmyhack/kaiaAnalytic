package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// ChatEngine processes user queries and facilitates on-chain actions
type ChatEngine struct {
	logger          *logrus.Logger
	llm             llms.LLM
	analyticsEngine *AnalyticsEngine
	dataCollector   *DataCollector
	actionContract  string // Address of the ActionContract
	wsUpgrader      websocket.Upgrader
	connections     map[string]*websocket.Conn
	connectionsMu   sync.RWMutex
	queryHistory    map[string][]*ChatMessage
	historyMu       sync.RWMutex
	intents         map[string]*QueryIntent
	rateLimiter     map[string]*RateLimiter
	rateLimiterMu   sync.RWMutex
}

type ChatMessage struct {
	ID            string                 `json:"id"`
	UserID        string                 `json:"user_id"`
	Message       string                 `json:"message"`
	Response      string                 `json:"response"`
	Intent        string                 `json:"intent"`
	Confidence    float64                `json:"confidence"`
	ActionID      string                 `json:"action_id,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	SessionID     string                 `json:"session_id"`
	Metadata      map[string]interface{} `json:"metadata"`
	IsError       bool                   `json:"is_error"`
	RequiresPremium bool                 `json:"requires_premium"`
}

type QueryIntent struct {
	Intent      string                 `json:"intent"`
	Entities    map[string]interface{} `json:"entities"`
	Confidence  float64                `json:"confidence"`
	ActionType  string                 `json:"action_type,omitempty"`
	Parameters  map[string]interface{} `json:"parameters"`
	RequiresPremium bool               `json:"requires_premium"`
}

type ChatContext struct {
	UserID        string    `json:"user_id"`
	SessionID     string    `json:"session_id"`
	Subscription  string    `json:"subscription_tier"`
	LastActivity  time.Time `json:"last_activity"`
	QueryCount    int       `json:"query_count"`
	DailyLimit    int       `json:"daily_limit"`
	Preferences   map[string]interface{} `json:"preferences"`
}

type ActionRequest struct {
	UserID      string                 `json:"user_id"`
	ActionType  string                 `json:"action_type"`
	Parameters  map[string]interface{} `json:"parameters"`
	ChatContext string                 `json:"chat_context"`
	Confirmation bool                  `json:"confirmation"`
}

type RateLimiter struct {
	Requests    int       `json:"requests"`
	LastReset   time.Time `json:"last_reset"`
	Limit       int       `json:"limit"`
	WindowSize  time.Duration `json:"window_size"`
}

// WebSocket message types
const (
	MessageTypeQuery    = "query"
	MessageTypeResponse = "response"
	MessageTypeAction   = "action"
	MessageTypeError    = "error"
	MessageTypeStatus   = "status"
)

// Query intents
const (
	IntentYieldQuery      = "yield_query"
	IntentPriceQuery      = "price_query"
	IntentTradeAnalysis   = "trade_analysis"
	IntentGovernanceQuery = "governance_query"
	IntentActionRequest   = "action_request"
	IntentGeneral         = "general"
	IntentStake           = "stake"
	IntentSwap            = "swap"
	IntentVote            = "vote"
	IntentBalance         = "balance"
	IntentPortfolio       = "portfolio"
)

// NewChatEngine creates a new chat engine
func NewChatEngine(
	logger *logrus.Logger,
	analyticsEngine *AnalyticsEngine,
	dataCollector *DataCollector,
	actionContract string,
	openaiKey string,
) (*ChatEngine, error) {
	// Initialize LLM
	var llm llms.LLM
	var err error
	if openaiKey != "" {
		llm, err = openai.New(openai.WithToken(openaiKey))
		if err != nil {
			logger.WithError(err).Warn("Failed to initialize OpenAI client")
		}
	}

	// Initialize WebSocket upgrader
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins in development
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	engine := &ChatEngine{
		logger:          logger,
		llm:             llm,
		analyticsEngine: analyticsEngine,
		dataCollector:   dataCollector,
		actionContract:  actionContract,
		wsUpgrader:      upgrader,
		connections:     make(map[string]*websocket.Conn),
		queryHistory:    make(map[string][]*ChatMessage),
		intents:         make(map[string]*QueryIntent),
		rateLimiter:     make(map[string]*RateLimiter),
	}

	// Initialize intent patterns
	engine.initializeIntents()

	return engine, nil
}

// HandleWebSocket handles WebSocket connections
func (ce *ChatEngine) HandleWebSocket(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id required"})
		return
	}

	// Upgrade connection
	conn, err := ce.wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		ce.logger.WithError(err).Error("Failed to upgrade WebSocket connection")
		return
	}
	defer conn.Close()

	// Store connection
	ce.connectionsMu.Lock()
	ce.connections[userID] = conn
	ce.connectionsMu.Unlock()

	// Remove connection when done
	defer func() {
		ce.connectionsMu.Lock()
		delete(ce.connections, userID)
		ce.connectionsMu.Unlock()
	}()

	ce.logger.WithField("user_id", userID).Info("WebSocket connection established")

	// Send welcome message
	welcomeMsg := &ChatMessage{
		ID:        generateMessageID(),
		Message:   "",
		Response:  "Welcome to KaiaAnalyticsAI! How can I help you today?",
		Intent:    "welcome",
		Timestamp: time.Now(),
		SessionID: generateSessionID(),
	}
	ce.sendMessage(conn, welcomeMsg)

	// Handle messages
	for {
		var message map[string]interface{}
		if err := conn.ReadJSON(&message); err != nil {
			ce.logger.WithError(err).Debug("WebSocket connection closed")
			break
		}

		// Process message
		go ce.processWebSocketMessage(userID, message, conn)
	}
}

// ProcessQuery processes a text query and returns a response
func (ce *ChatEngine) ProcessQuery(ctx context.Context, userID, query string, sessionID string) (*ChatMessage, error) {
	ce.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"session_id": sessionID,
		"query":      query,
	}).Info("Processing chat query")

	// Check rate limits
	if !ce.checkRateLimit(userID) {
		return &ChatMessage{
			ID:        generateMessageID(),
			UserID:    userID,
			Message:   query,
			Response:  "Rate limit exceeded. Please try again later.",
			Intent:    "rate_limit",
			Timestamp: time.Now(),
			SessionID: sessionID,
			IsError:   true,
		}, nil
	}

	// Analyze query intent
	intent, err := ce.analyzeIntent(ctx, query)
	if err != nil {
		ce.logger.WithError(err).Error("Failed to analyze intent")
		intent = &QueryIntent{
			Intent:     IntentGeneral,
			Confidence: 0.1,
		}
	}

	// Check if premium features are required
	if intent.RequiresPremium {
		// Check user subscription (this would integrate with SubscriptionContract)
		hasPremium := ce.checkPremiumAccess(userID)
		if !hasPremium {
			return &ChatMessage{
				ID:              generateMessageID(),
				UserID:          userID,
				Message:         query,
				Response:        "This feature requires a premium subscription. Please upgrade to access advanced analytics and on-chain actions.",
				Intent:          intent.Intent,
				Confidence:      intent.Confidence,
				Timestamp:       time.Now(),
				SessionID:       sessionID,
				RequiresPremium: true,
			}, nil
		}
	}

	// Generate response based on intent
	response, actionID, err := ce.generateResponse(ctx, userID, query, intent)
	if err != nil {
		ce.logger.WithError(err).Error("Failed to generate response")
		response = "I'm sorry, I encountered an error processing your request. Please try again."
	}

	message := &ChatMessage{
		ID:         generateMessageID(),
		UserID:     userID,
		Message:    query,
		Response:   response,
		Intent:     intent.Intent,
		Confidence: intent.Confidence,
		ActionID:   actionID,
		Timestamp:  time.Now(),
		SessionID:  sessionID,
		Metadata:   map[string]interface{}{"entities": intent.Entities},
	}

	// Store in history
	ce.addToHistory(userID, message)

	return message, nil
}

// analyzeIntent analyzes the user's query to determine intent and extract entities
func (ce *ChatEngine) analyzeIntent(ctx context.Context, query string) (*QueryIntent, error) {
	query = strings.ToLower(strings.TrimSpace(query))

	// Pattern-based intent recognition
	for intentType, patterns := range ce.getIntentPatterns() {
		for _, pattern := range patterns {
			if matched, _ := regexp.MatchString(pattern, query); matched {
				entities := ce.extractEntities(query, intentType)
				confidence := ce.calculateIntentConfidence(query, intentType)
				
				return &QueryIntent{
					Intent:          intentType,
					Entities:        entities,
					Confidence:      confidence,
					RequiresPremium: ce.isIntentPremium(intentType),
					Parameters:      entities,
				}, nil
			}
		}
	}

	// Use LLM for more sophisticated intent analysis if available
	if ce.llm != nil {
		return ce.llmAnalyzeIntent(ctx, query)
	}

	// Default to general intent
	return &QueryIntent{
		Intent:     IntentGeneral,
		Entities:   make(map[string]interface{}),
		Confidence: 0.5,
	}, nil
}

// generateResponse generates a response based on the query intent
func (ce *ChatEngine) generateResponse(ctx context.Context, userID, query string, intent *QueryIntent) (string, string, error) {
	switch intent.Intent {
	case IntentYieldQuery:
		return ce.handleYieldQuery(ctx, intent)
	case IntentPriceQuery:
		return ce.handlePriceQuery(ctx, intent)
	case IntentTradeAnalysis:
		return ce.handleTradeAnalysis(ctx, userID, intent)
	case IntentGovernanceQuery:
		return ce.handleGovernanceQuery(ctx, intent)
	case IntentActionRequest, IntentStake, IntentSwap, IntentVote:
		return ce.handleActionRequest(ctx, userID, intent, query)
	case IntentBalance:
		return ce.handleBalanceQuery(ctx, userID, intent)
	case IntentPortfolio:
		return ce.handlePortfolioQuery(ctx, userID, intent)
	default:
		return ce.handleGeneralQuery(ctx, query)
	}
}

// handleYieldQuery handles yield-related queries
func (ce *ChatEngine) handleYieldQuery(ctx context.Context, intent *QueryIntent) (string, string, error) {
	protocols := []string{"uniswap", "aave", "compound", "curve"}
	if protocol, exists := intent.Entities["protocol"]; exists {
		if protocolStr, ok := protocol.(string); ok {
			protocols = []string{protocolStr}
		}
	}

	yields, err := ce.analyticsEngine.AnalyzeYieldOpportunities(ctx, protocols)
	if err != nil {
		return "I'm having trouble fetching yield data right now. Please try again later.", "", err
	}

	if len(yields) == 0 {
		return "I couldn't find any yield opportunities at the moment.", "", nil
	}

	// Format response
	response := "Here are the best yield opportunities I found:\n\n"
	for i, yield := range yields {
		if i >= 3 { // Limit to top 3
			break
		}
		response += fmt.Sprintf("🌾 **%s**\n", yield.Protocol)
		response += fmt.Sprintf("   APR: %.2f%% | Risk Score: %d/100\n", yield.APR*100, yield.RiskScore)
		response += fmt.Sprintf("   TVL: $%.2fM | Min Deposit: $%.0f\n", yield.TVL/1000000, yield.MinDeposit)
		response += fmt.Sprintf("   %s\n\n", yield.Recommendation)
	}

	response += "Would you like me to help you stake in any of these protocols?"

	return response, "", nil
}

// handlePriceQuery handles price-related queries
func (ce *ChatEngine) handlePriceQuery(ctx context.Context, intent *QueryIntent) (string, string, error) {
	token := "KAIA" // Default
	if tokenEntity, exists := intent.Entities["token"]; exists {
		if tokenStr, ok := tokenEntity.(string); ok {
			token = strings.ToUpper(tokenStr)
		}
	}

	price, err := ce.dataCollector.GetTokenPrice(ctx, token)
	if err != nil {
		return fmt.Sprintf("I couldn't fetch the price for %s right now. Please try again later.", token), "", err
	}

	response := fmt.Sprintf("💰 **%s Price Information**\n\n", token)
	response += fmt.Sprintf("Current Price: $%.4f\n", price.Price)
	response += fmt.Sprintf("24h Change: %.2f%%\n", price.PriceChange24h)
	response += fmt.Sprintf("7d Change: %.2f%%\n", price.PriceChange7d)
	response += fmt.Sprintf("24h High: $%.4f\n", price.High24h)
	response += fmt.Sprintf("24h Low: $%.4f\n", price.Low24h)
	response += fmt.Sprintf("Volume: $%.2fM\n", price.Volume24h/1000000)
	response += fmt.Sprintf("Market Cap: $%.2fM\n", price.MarketCap/1000000)

	return response, "", nil
}

// handleTradeAnalysis handles trading analysis queries
func (ce *ChatEngine) handleTradeAnalysis(ctx context.Context, userID string, intent *QueryIntent) (string, string, error) {
	tokenPair := "KAIA/USDC" // Default
	if pairEntity, exists := intent.Entities["token_pair"]; exists {
		if pairStr, ok := pairEntity.(string); ok {
			tokenPair = pairStr
		}
	}

	optimization, err := ce.analyticsEngine.OptimizeTradingStrategy(ctx, userID, tokenPair)
	if err != nil {
		return "I'm having trouble analyzing trading data right now. Please try again later.", "", err
	}

	response := fmt.Sprintf("📈 **Trading Analysis for %s**\n\n", tokenPair)
	response += fmt.Sprintf("**Recommendation: %s**\n", strings.ToUpper(optimization.Action))
	response += fmt.Sprintf("Current Price: $%.4f\n", optimization.CurrentPrice)
	response += fmt.Sprintf("Target Price: $%.4f\n", optimization.TargetPrice)
	response += fmt.Sprintf("24h Change: %.2f%%\n", optimization.PriceChange24h)
	response += fmt.Sprintf("RSI: %.1f | MACD: %.4f\n", optimization.RSI, optimization.MACD)
	response += fmt.Sprintf("Risk Level: %s\n", optimization.RiskLevel)
	response += fmt.Sprintf("Confidence: %.1f%%\n\n", optimization.Confidence*100)
	response += fmt.Sprintf("💡 %s\n\n", optimization.Recommendation)

	if optimization.Action != "hold" {
		response += fmt.Sprintf("Optimal Amount: %.4f tokens\n", optimization.OptimalAmount)
		response += fmt.Sprintf("Estimated Gas Cost: %.6f ETH\n", optimization.GasCost)
		response += "\nWould you like me to help you execute this trade?"
	}

	return response, "", nil
}

// handleGovernanceQuery handles governance-related queries
func (ce *ChatEngine) handleGovernanceQuery(ctx context.Context, intent *QueryIntent) (string, string, error) {
	proposalID := "1" // Default or latest
	if idEntity, exists := intent.Entities["proposal_id"]; exists {
		if idStr, ok := idEntity.(string); ok {
			proposalID = idStr
		}
	}

	sentiment, err := ce.analyticsEngine.AnalyzeGovernanceSentiment(ctx, proposalID)
	if err != nil {
		return "I'm having trouble fetching governance data right now. Please try again later.", "", err
	}

	response := fmt.Sprintf("🗳️ **Governance Analysis - Proposal %s**\n\n", proposalID)
	response += fmt.Sprintf("**%s**\n\n", sentiment.ProposalTitle)
	response += fmt.Sprintf("Sentiment Score: %.1f/100\n", sentiment.SentimentScore*50+50) // Convert to 0-100 scale
	response += fmt.Sprintf("Participation Rate: %.1f%%\n", sentiment.ParticipationRate)
	response += fmt.Sprintf("Predicted Outcome: %s\n", sentiment.PredictedOutcome)
	response += fmt.Sprintf("Confidence: %.1f%%\n\n", sentiment.Confidence*100)

	if len(sentiment.KeyTopics) > 0 {
		response += "**Key Topics:**\n"
		for _, topic := range sentiment.KeyTopics {
			response += fmt.Sprintf("• %s\n", topic)
		}
		response += "\n"
	}

	response += "Would you like me to help you vote on this proposal?"

	return response, "", nil
}

// handleActionRequest handles on-chain action requests
func (ce *ChatEngine) handleActionRequest(ctx context.Context, userID string, intent *QueryIntent, originalQuery string) (string, string, error) {
	// This would integrate with the ActionContract to request on-chain actions
	actionType := intent.Intent
	parameters := intent.Parameters

	// For now, return a mock response
	response := fmt.Sprintf("🔗 **Action Request: %s**\n\n", strings.Title(actionType))
	response += "I understand you want to perform an on-chain action. "
	response += "For security reasons, I'll need you to confirm this action.\n\n"
	
	if actionType == IntentStake {
		if amount, exists := parameters["amount"]; exists {
			response += fmt.Sprintf("Amount to stake: %v\n", amount)
		}
		if protocol, exists := parameters["protocol"]; exists {
			response += fmt.Sprintf("Protocol: %s\n", protocol)
		}
	} else if actionType == IntentSwap {
		if tokenIn, exists := parameters["token_in"]; exists {
			response += fmt.Sprintf("From: %v\n", tokenIn)
		}
		if tokenOut, exists := parameters["token_out"]; exists {
			response += fmt.Sprintf("To: %v\n", tokenOut)
		}
		if amount, exists := parameters["amount"]; exists {
			response += fmt.Sprintf("Amount: %v\n", amount)
		}
	}

	response += "\n⚠️ Please review carefully and type 'confirm' to proceed."

	// Generate action ID for tracking
	actionID := fmt.Sprintf("action_%d", time.Now().Unix())

	return response, actionID, nil
}

// handleBalanceQuery handles balance queries
func (ce *ChatEngine) handleBalanceQuery(ctx context.Context, userID string, intent *QueryIntent) (string, string, error) {
	// This would query user's actual balances
	response := "💰 **Your Portfolio Balance**\n\n"
	response += "KAIA: 1,250.45 KAIA ($2,156.78)\n"
	response += "ETH: 0.5634 ETH ($1,891.23)\n"
	response += "USDC: 500.00 USDC ($500.00)\n\n"
	response += "**Total Portfolio Value: $4,548.01**\n"
	response += "24h Change: +2.34% (+$103.45)\n\n"
	response += "Would you like to see a detailed breakdown or trading recommendations?"

	return response, "", nil
}

// handlePortfolioQuery handles portfolio analysis queries
func (ce *ChatEngine) handlePortfolioQuery(ctx context.Context, userID string, intent *QueryIntent) (string, string, error) {
	response := "📊 **Portfolio Analysis**\n\n"
	response += "**Asset Allocation:**\n"
	response += "• KAIA: 47.4% ($2,156.78)\n"
	response += "• ETH: 41.6% ($1,891.23)\n"
	response += "• USDC: 11.0% ($500.00)\n\n"
	response += "**Performance:**\n"
	response += "• 7d: +12.3% ($498.23)\n"
	response += "• 30d: +28.7% ($1,015.67)\n"
	response += "• YTD: +145.2% ($2,834.12)\n\n"
	response += "**Recommendations:**\n"
	response += "• Consider rebalancing - KAIA allocation is high\n"
	response += "• Opportunity to stake KAIA for 8.5% APR\n"
	response += "• ETH showing strong momentum\n\n"
	response += "Would you like me to suggest specific actions?"

	return response, "", nil
}

// handleGeneralQuery handles general queries using LLM if available
func (ce *ChatEngine) handleGeneralQuery(ctx context.Context, query string) (string, string, error) {
	if ce.llm != nil {
		return ce.llmGenerateResponse(ctx, query)
	}

	// Fallback responses for common general queries
	generalResponses := map[string]string{
		"help":    "I can help you with:\n• Yield farming opportunities\n• Token prices and market data\n• Trading analysis and recommendations\n• Governance proposal analysis\n• Portfolio management\n• On-chain actions (staking, swapping, voting)\n\nWhat would you like to know about?",
		"hello":   "Hello! I'm your AI assistant for Kaia blockchain analytics. How can I help you today?",
		"thanks":  "You're welcome! Is there anything else I can help you with?",
		"bye":     "Goodbye! Feel free to ask me anything about Kaia analytics anytime.",
	}

	queryLower := strings.ToLower(query)
	for keyword, response := range generalResponses {
		if strings.Contains(queryLower, keyword) {
			return response, "", nil
		}
	}

	return "I'm here to help with Kaia blockchain analytics, yield farming, trading, and governance. What would you like to know?", "", nil
}

// Helper functions

func (ce *ChatEngine) initializeIntents() {
	// Initialize intent patterns and configurations
	ce.intents = map[string]*QueryIntent{
		IntentYieldQuery: {
			RequiresPremium: false,
		},
		IntentPriceQuery: {
			RequiresPremium: false,
		},
		IntentTradeAnalysis: {
			RequiresPremium: true,
		},
		IntentGovernanceQuery: {
			RequiresPremium: true,
		},
		IntentActionRequest: {
			RequiresPremium: true,
		},
		IntentStake: {
			RequiresPremium: true,
		},
		IntentSwap: {
			RequiresPremium: true,
		},
		IntentVote: {
			RequiresPremium: true,
		},
	}
}

func (ce *ChatEngine) getIntentPatterns() map[string][]string {
	return map[string][]string{
		IntentYieldQuery: {
			`(?i)(yield|apr|apy|farm|liquidity|staking rewards|earning)`,
			`(?i)(best.*yield|highest.*apr|farming.*opportunities)`,
			`(?i)(where.*stake|what.*yield|which.*pool)`,
		},
		IntentPriceQuery: {
			`(?i)(price|cost|value|worth).*\b(kaia|eth|btc|usdc)\b`,
			`(?i)\b(kaia|eth|btc|usdc)\b.*(price|cost|value)`,
			`(?i)(how much|what.*price|current.*price)`,
		},
		IntentTradeAnalysis: {
			`(?i)(trade|trading|buy|sell|analysis)`,
			`(?i)(should.*buy|should.*sell|trading.*signal)`,
			`(?i)(technical.*analysis|rsi|macd|bollinger)`,
		},
		IntentGovernanceQuery: {
			`(?i)(governance|proposal|vote|voting)`,
			`(?i)(what.*proposal|should.*vote|governance.*analysis)`,
		},
		IntentStake: {
			`(?i)(stake|staking).*\d+`,
			`(?i)(deposit|lock|stake).*tokens?`,
		},
		IntentSwap: {
			`(?i)(swap|exchange|convert).*\b(kaia|eth|usdc)\b`,
			`(?i)(trade|sell|buy).*\d+`,
		},
		IntentVote: {
			`(?i)(vote|voting).*proposal`,
			`(?i)(support|against|yes|no).*proposal`,
		},
		IntentBalance: {
			`(?i)(balance|portfolio|holdings)`,
			`(?i)(my.*tokens|how much.*have|what.*own)`,
		},
		IntentPortfolio: {
			`(?i)(portfolio.*analysis|allocation|performance)`,
			`(?i)(how.*performing|portfolio.*breakdown)`,
		},
	}
}

func (ce *ChatEngine) extractEntities(query, intent string) map[string]interface{} {
	entities := make(map[string]interface{})
	queryLower := strings.ToLower(query)

	// Extract common entities
	if matches := regexp.MustCompile(`\b(\d+(?:\.\d+)?)\b`).FindAllString(queryLower, -1); len(matches) > 0 {
		if amount, err := strconv.ParseFloat(matches[0], 64); err == nil {
			entities["amount"] = amount
		}
	}

	// Extract token names
	tokens := []string{"kaia", "eth", "btc", "usdc", "dai", "usdt"}
	for _, token := range tokens {
		if strings.Contains(queryLower, token) {
			entities["token"] = strings.ToUpper(token)
			break
		}
	}

	// Extract token pairs
	if matches := regexp.MustCompile(`\b([a-z]+)[/\-]([a-z]+)\b`).FindStringSubmatch(queryLower); len(matches) == 3 {
		entities["token_pair"] = fmt.Sprintf("%s/%s", strings.ToUpper(matches[1]), strings.ToUpper(matches[2]))
		entities["token_in"] = strings.ToUpper(matches[1])
		entities["token_out"] = strings.ToUpper(matches[2])
	}

	// Extract protocols
	protocols := []string{"uniswap", "aave", "compound", "curve", "yearn", "sushiswap"}
	for _, protocol := range protocols {
		if strings.Contains(queryLower, protocol) {
			entities["protocol"] = protocol
			break
		}
	}

	// Extract proposal IDs
	if matches := regexp.MustCompile(`proposal\s+(\d+)`).FindStringSubmatch(queryLower); len(matches) == 2 {
		entities["proposal_id"] = matches[1]
	}

	return entities
}

func (ce *ChatEngine) calculateIntentConfidence(query, intent string) float64 {
	// Simple confidence calculation based on keyword matches
	patterns := ce.getIntentPatterns()[intent]
	confidence := 0.0
	
	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, strings.ToLower(query)); matched {
			confidence += 0.3
		}
	}

	return min(confidence, 1.0)
}

func (ce *ChatEngine) isIntentPremium(intent string) bool {
	premiumIntents := map[string]bool{
		IntentTradeAnalysis:   true,
		IntentGovernanceQuery: true,
		IntentActionRequest:   true,
		IntentStake:          true,
		IntentSwap:           true,
		IntentVote:           true,
	}
	return premiumIntents[intent]
}

func (ce *ChatEngine) checkRateLimit(userID string) bool {
	ce.rateLimiterMu.Lock()
	defer ce.rateLimiterMu.Unlock()

	now := time.Now()
	limiter, exists := ce.rateLimiter[userID]
	
	if !exists {
		limiter = &RateLimiter{
			Requests:   0,
			LastReset:  now,
			Limit:      60, // 60 requests per hour for free users
			WindowSize: time.Hour,
		}
		ce.rateLimiter[userID] = limiter
	}

	// Reset if window has passed
	if now.Sub(limiter.LastReset) >= limiter.WindowSize {
		limiter.Requests = 0
		limiter.LastReset = now
	}

	// Check limit
	if limiter.Requests >= limiter.Limit {
		return false
	}

	limiter.Requests++
	return true
}

func (ce *ChatEngine) checkPremiumAccess(userID string) bool {
	// This would integrate with SubscriptionContract
	// For now, return true for testing
	return true
}

func (ce *ChatEngine) addToHistory(userID string, message *ChatMessage) {
	ce.historyMu.Lock()
	defer ce.historyMu.Unlock()

	history := ce.queryHistory[userID]
	history = append(history, message)

	// Keep only last 50 messages
	if len(history) > 50 {
		history = history[len(history)-50:]
	}

	ce.queryHistory[userID] = history
}

func (ce *ChatEngine) processWebSocketMessage(userID string, message map[string]interface{}, conn *websocket.Conn) {
	messageType, ok := message["type"].(string)
	if !ok {
		ce.sendError(conn, "Invalid message format")
		return
	}

	switch messageType {
	case MessageTypeQuery:
		query, ok := message["message"].(string)
		if !ok {
			ce.sendError(conn, "Missing query message")
			return
		}

		sessionID := generateSessionID()
		if sid, exists := message["session_id"].(string); exists {
			sessionID = sid
		}

		ctx := context.Background()
		response, err := ce.ProcessQuery(ctx, userID, query, sessionID)
		if err != nil {
			ce.sendError(conn, "Failed to process query")
			return
		}

		ce.sendMessage(conn, response)

	case MessageTypeAction:
		// Handle action confirmation
		actionID, ok := message["action_id"].(string)
		if !ok {
			ce.sendError(conn, "Missing action ID")
			return
		}

		confirmation, ok := message["confirmation"].(bool)
		if !ok {
			ce.sendError(conn, "Missing confirmation")
			return
		}

		if confirmation {
			ce.sendMessage(conn, &ChatMessage{
				ID:        generateMessageID(),
				Response:  "Action confirmed! I'm processing your request...",
				Intent:    "action_confirmed",
				ActionID:  actionID,
				Timestamp: time.Now(),
			})
		} else {
			ce.sendMessage(conn, &ChatMessage{
				ID:        generateMessageID(),
				Response:  "Action cancelled. Is there anything else I can help you with?",
				Intent:    "action_cancelled",
				Timestamp: time.Now(),
			})
		}
	}
}

func (ce *ChatEngine) sendMessage(conn *websocket.Conn, message *ChatMessage) {
	if err := conn.WriteJSON(message); err != nil {
		ce.logger.WithError(err).Error("Failed to send WebSocket message")
	}
}

func (ce *ChatEngine) sendError(conn *websocket.Conn, errorMsg string) {
	ce.sendMessage(conn, &ChatMessage{
		ID:        generateMessageID(),
		Response:  errorMsg,
		Intent:    "error",
		Timestamp: time.Now(),
		IsError:   true,
	})
}

// LLM integration functions
func (ce *ChatEngine) llmAnalyzeIntent(ctx context.Context, query string) (*QueryIntent, error) {
	if ce.llm == nil {
		return nil, fmt.Errorf("LLM not available")
	}

	prompt := fmt.Sprintf(`Analyze this blockchain/DeFi query and extract intent and entities:
Query: "%s"

Available intents: yield_query, price_query, trade_analysis, governance_query, action_request, stake, swap, vote, balance, portfolio, general

Respond with JSON containing:
- intent: one of the available intents
- entities: key-value pairs of extracted entities (tokens, amounts, protocols, etc.)
- confidence: confidence score 0-1
- requires_premium: boolean if this requires premium features`, query)

	response, err := ce.llm.GenerateContent(ctx, []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart(prompt),
			},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("LLM request failed: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no LLM response received")
	}

	// Parse JSON response
	var result QueryIntent
	if err := json.Unmarshal([]byte(response.Choices[0].Content), &result); err != nil {
		// Fallback to pattern matching if JSON parsing fails
		return ce.analyzeIntent(ctx, query)
	}

	return &result, nil
}

func (ce *ChatEngine) llmGenerateResponse(ctx context.Context, query string) (string, string, error) {
	if ce.llm == nil {
		return "", "", fmt.Errorf("LLM not available")
	}

	prompt := fmt.Sprintf(`You are an AI assistant for KaiaAnalyticsAI, a blockchain analytics platform. 
Answer this user query about blockchain, DeFi, or Kaia ecosystem:

Query: "%s"

Provide a helpful, informative response. If the query is about specific data that requires real-time information, suggest that the user try more specific commands like "price KAIA" or "yield farming opportunities".`, query)

	response, err := ce.llm.GenerateContent(ctx, []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart(prompt),
			},
		},
	})

	if err != nil {
		return "", "", fmt.Errorf("LLM request failed: %w", err)
	}

	if len(response.Choices) == 0 {
		return "I'm sorry, I couldn't generate a response. Please try asking about specific topics like prices, yields, or trading.", "", nil
	}

	return response.Choices[0].Content, "", nil
}

// Utility functions
func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().Unix())
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// GetChatHistory returns chat history for a user
func (ce *ChatEngine) GetChatHistory(userID string, limit int) []*ChatMessage {
	ce.historyMu.RLock()
	defer ce.historyMu.RUnlock()

	history := ce.queryHistory[userID]
	if len(history) == 0 {
		return []*ChatMessage{}
	}

	start := 0
	if len(history) > limit {
		start = len(history) - limit
	}

	return history[start:]
}

// BroadcastMessage sends a message to all connected users
func (ce *ChatEngine) BroadcastMessage(message *ChatMessage) {
	ce.connectionsMu.RLock()
	defer ce.connectionsMu.RUnlock()

	for userID, conn := range ce.connections {
		go func(uid string, c *websocket.Conn) {
			ce.sendMessage(c, message)
		}(userID, conn)
	}
}

// Close closes the chat engine and cleans up resources
func (ce *ChatEngine) Close() {
	ce.connectionsMu.Lock()
	defer ce.connectionsMu.Unlock()

	for _, conn := range ce.connections {
		conn.Close()
	}
	ce.connections = make(map[string]*websocket.Conn)
}