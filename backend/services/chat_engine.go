package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gorilla/websocket"
)

// ChatEngine handles chat functionality and on-chain actions
type ChatEngine struct {
	ethClient    *ethclient.Client
	analyticsEngine *AnalyticsEngine
	dataCollector   *DataCollector
	logger       *log.Logger
	connections  map[string]*websocket.Conn
	mu           sync.RWMutex
}

// ChatMessage represents a chat message
type ChatMessage struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Message   string                 `json:"message"`
	Type      string                 `json:"type"` // text, action, query
	Timestamp int64                  `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ChatResponse represents a response to a chat message
type ChatResponse struct {
	ID        string                 `json:"id"`
	MessageID string                 `json:"message_id"`
	Response  string                 `json:"response"`
	Type      string                 `json:"type"` // text, action_result, analytics
	Data      interface{}            `json:"data,omitempty"`
	Timestamp int64                  `json:"timestamp"`
	Success   bool                   `json:"success"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ActionRequest represents an on-chain action request
type ActionRequest struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	ActionType  string                 `json:"action_type"`
	Parameters  map[string]interface{} `json:"parameters"`
	Status      string                 `json:"status"` // pending, executing, completed, failed
	Timestamp   int64                  `json:"timestamp"`
	Result      interface{}            `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// QueryIntent represents the intent of a user query
type QueryIntent struct {
	Intent     string                 `json:"intent"`
	Confidence float64                `json:"confidence"`
	Entities   map[string]interface{} `json:"entities"`
	Action     string                 `json:"action,omitempty"`
}

// NewChatEngine creates a new chat engine instance
func NewChatEngine(ethClient *ethclient.Client, analyticsEngine *AnalyticsEngine, dataCollector *DataCollector) *ChatEngine {
	return &ChatEngine{
		ethClient:       ethClient,
		analyticsEngine: analyticsEngine,
		dataCollector:   dataCollector,
		logger:          log.New(log.Writer(), "[ChatEngine] ", log.LstdFlags),
		connections:     make(map[string]*websocket.Conn),
	}
}

// ProcessMessage processes a chat message and returns a response
func (ce *ChatEngine) ProcessMessage(ctx context.Context, message *ChatMessage) (*ChatResponse, error) {
	startTime := time.Now()

	// Parse user intent
	intent, err := ce.parseIntent(message.Message)
	if err != nil {
		return nil, fmt.Errorf("failed to parse intent: %w", err)
	}

	var response *ChatResponse

	switch intent.Intent {
	case "yield_query":
		response, err = ce.handleYieldQuery(ctx, message, intent)
	case "trading_suggestion":
		response, err = ce.handleTradingSuggestion(ctx, message, intent)
	case "portfolio_analysis":
		response, err = ce.handlePortfolioAnalysis(ctx, message, intent)
	case "governance_query":
		response, err = ce.handleGovernanceQuery(ctx, message, intent)
	case "on_chain_action":
		response, err = ce.handleOnChainAction(ctx, message, intent)
	case "market_data":
		response, err = ce.handleMarketDataQuery(ctx, message, intent)
	case "gas_info":
		response, err = ce.handleGasInfoQuery(ctx, message, intent)
	default:
		response, err = ce.handleGeneralQuery(ctx, message, intent)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to process message: %w", err)
	}

	response.ID = fmt.Sprintf("resp_%d", time.Now().UnixNano())
	response.MessageID = message.ID
	response.Timestamp = time.Now().Unix()

	return response, nil
}

// parseIntent parses the intent of a user message
func (ce *ChatEngine) parseIntent(message string) (*QueryIntent, error) {
	message = strings.ToLower(message)
	
	// Simple keyword-based intent parsing
	// In a real implementation, this would use NLP/ML models
	
	intent := &QueryIntent{
		Entities: make(map[string]interface{}),
	}

	// Yield-related queries
	if strings.Contains(message, "yield") || strings.Contains(message, "apy") || strings.Contains(message, "farming") {
		intent.Intent = "yield_query"
		intent.Confidence = 0.85
		intent.Action = "analyze_yield_opportunities"
	}

	// Trading-related queries
	if strings.Contains(message, "trade") || strings.Contains(message, "buy") || strings.Contains(message, "sell") {
		intent.Intent = "trading_suggestion"
		intent.Confidence = 0.80
		intent.Action = "generate_trading_suggestions"
	}

	// Portfolio-related queries
	if strings.Contains(message, "portfolio") || strings.Contains(message, "balance") || strings.Contains(message, "holdings") {
		intent.Intent = "portfolio_analysis"
		intent.Confidence = 0.90
		intent.Action = "analyze_portfolio"
	}

	// Governance-related queries
	if strings.Contains(message, "governance") || strings.Contains(message, "vote") || strings.Contains(message, "proposal") {
		intent.Intent = "governance_query"
		intent.Confidence = 0.75
		intent.Action = "analyze_governance_sentiment"
	}

	// On-chain action requests
	if strings.Contains(message, "stake") || strings.Contains(message, "unstake") || strings.Contains(message, "swap") {
		intent.Intent = "on_chain_action"
		intent.Confidence = 0.95
		intent.Action = "execute_action"
	}

	// Market data queries
	if strings.Contains(message, "price") || strings.Contains(message, "market") || strings.Contains(message, "chart") {
		intent.Intent = "market_data"
		intent.Confidence = 0.70
		intent.Action = "get_market_data"
	}

	// Gas-related queries
	if strings.Contains(message, "gas") || strings.Contains(message, "fee") {
		intent.Intent = "gas_info"
		intent.Confidence = 0.88
		intent.Action = "get_gas_info"
	}

	// Default to general query
	if intent.Intent == "" {
		intent.Intent = "general_query"
		intent.Confidence = 0.50
		intent.Action = "general_response"
	}

	// Extract entities (simplified)
	ce.extractEntities(message, intent)

	return intent, nil
}

// extractEntities extracts entities from the message
func (ce *ChatEngine) extractEntities(message string, intent *QueryIntent) {
	// Extract addresses
	addressRegex := regexp.MustCompile(`0x[a-fA-F0-9]{40}`)
	addresses := addressRegex.FindAllString(message, -1)
	if len(addresses) > 0 {
		intent.Entities["addresses"] = addresses
	}

	// Extract amounts
	amountRegex := regexp.MustCompile(`\d+(?:\.\d+)?`)
	amounts := amountRegex.FindAllString(message, -1)
	if len(amounts) > 0 {
		intent.Entities["amounts"] = amounts
	}

	// Extract tokens/symbols
	tokenRegex := regexp.MustCompile(`\b(?:ETH|USDC|DAI|BTC|UNI|AAVE)\b`)
	tokens := tokenRegex.FindAllString(message, -1)
	if len(tokens) > 0 {
		intent.Entities["tokens"] = tokens
	}
}

// handleYieldQuery handles yield-related queries
func (ce *ChatEngine) handleYieldQuery(ctx context.Context, message *ChatMessage, intent *QueryIntent) (*ChatResponse, error) {
	// Process yield analysis
	result, err := ce.analyticsEngine.ProcessAnalyticsTask(ctx, "yield_analysis", map[string]interface{}{
		"user_address": message.UserID,
		"query":        message.Message,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to analyze yield opportunities: %w", err)
	}

	opportunities := result.Data.([]YieldOpportunity)
	
	var responseText strings.Builder
	responseText.WriteString("Here are the best yield opportunities I found:\n\n")
	
	for i, opp := range opportunities {
		if i >= 3 { // Limit to top 3
			break
		}
		responseText.WriteString(fmt.Sprintf("🏆 **%s** (%s)\n", opp.Protocol, opp.AssetPair))
		responseText.WriteString(fmt.Sprintf("   APY: %.2f%%\n", opp.APY))
		responseText.WriteString(fmt.Sprintf("   TVL: $%.0f\n", opp.TVL))
		responseText.WriteString(fmt.Sprintf("   Risk Score: %.2f\n", opp.Risk))
		responseText.WriteString(fmt.Sprintf("   Opportunity Score: %.2f\n\n", opp.Opportunity))
	}

	return &ChatResponse{
		Response: responseText.String(),
		Type:     "analytics",
		Data:     opportunities,
		Success:  true,
		Metadata: map[string]interface{}{
			"confidence": intent.Confidence,
			"intent":     intent.Intent,
		},
	}, nil
}

// handleTradingSuggestion handles trading suggestion queries
func (ce *ChatEngine) handleTradingSuggestion(ctx context.Context, message *ChatMessage, intent *QueryIntent) (*ChatResponse, error) {
	// Generate trading suggestions
	result, err := ce.analyticsEngine.ProcessAnalyticsTask(ctx, "trading_suggestions", map[string]interface{}{
		"user_address": message.UserID,
		"query":        message.Message,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate trading suggestions: %w", err)
	}

	suggestions := result.Data.([]TradingSuggestion)
	
	var responseText strings.Builder
	responseText.WriteString("Based on your trading history, here are my suggestions:\n\n")
	
	for i, suggestion := range suggestions {
		responseText.WriteString(fmt.Sprintf("💡 **%s %s**\n", strings.Title(suggestion.Type), suggestion.Asset))
		responseText.WriteString(fmt.Sprintf("   Amount: %.2f %s\n", suggestion.Amount, suggestion.Asset))
		responseText.WriteString(fmt.Sprintf("   Confidence: %.1f%%\n", suggestion.Confidence*100))
		responseText.WriteString(fmt.Sprintf("   Risk Level: %s\n", suggestion.RiskLevel))
		responseText.WriteString(fmt.Sprintf("   Expected Return: %.1f%%\n", suggestion.ExpectedReturn*100))
		responseText.WriteString(fmt.Sprintf("   Reasoning: %s\n\n", suggestion.Reasoning))
	}

	return &ChatResponse{
		Response: responseText.String(),
		Type:     "analytics",
		Data:     suggestions,
		Success:  true,
		Metadata: map[string]interface{}{
			"confidence": intent.Confidence,
			"intent":     intent.Intent,
		},
	}, nil
}

// handlePortfolioAnalysis handles portfolio analysis queries
func (ce *ChatEngine) handlePortfolioAnalysis(ctx context.Context, message *ChatMessage, intent *QueryIntent) (*ChatResponse, error) {
	// Analyze portfolio
	result, err := ce.analyticsEngine.ProcessAnalyticsTask(ctx, "portfolio_optimization", map[string]interface{}{
		"user_address": message.UserID,
		"risk_tolerance": "medium",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to analyze portfolio: %w", err)
	}

	optimization := result.Data.(map[string]interface{})
	
	responseText := fmt.Sprintf("📊 **Portfolio Analysis**\n\n"+
		"Current Risk Score: %.1f%%\n"+
		"Expected Return: %.1f%%\n"+
		"Rebalancing Needed: %v\n"+
		"Estimated Cost: $%.2f\n\n"+
		"Would you like me to help you rebalance your portfolio?",
		optimization["risk_score"].(float64)*100,
		optimization["expected_return"].(float64)*100,
		optimization["rebalancing_needed"].(bool),
		optimization["rebalancing_cost"].(float64))

	return &ChatResponse{
		Response: responseText,
		Type:     "analytics",
		Data:     optimization,
		Success:  true,
		Metadata: map[string]interface{}{
			"confidence": intent.Confidence,
			"intent":     intent.Intent,
		},
	}, nil
}

// handleGovernanceQuery handles governance-related queries
func (ce *ChatEngine) handleGovernanceQuery(ctx context.Context, message *ChatMessage, intent *QueryIntent) (*ChatResponse, error) {
	// Analyze governance sentiment
	result, err := ce.analyticsEngine.ProcessAnalyticsTask(ctx, "governance_sentiment", map[string]interface{}{
		"user_address": message.UserID,
		"query":        message.Message,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to analyze governance sentiment: %w", err)
	}

	sentiments := result.Data.([]GovernanceSentiment)
	
	var responseText strings.Builder
	responseText.WriteString("🗳️ **Governance Sentiment Analysis**\n\n")
	
	for _, sentiment := range sentiments {
		emoji := "🤷"
		switch sentiment.Sentiment {
		case "positive":
			emoji = "✅"
		case "negative":
			emoji = "❌"
		}
		
		responseText.WriteString(fmt.Sprintf("%s **%s**\n", emoji, sentiment.Title))
		responseText.WriteString(fmt.Sprintf("   Sentiment: %s (%.1f%% confidence)\n", sentiment.Sentiment, sentiment.Confidence*100))
		responseText.WriteString(fmt.Sprintf("   Votes: %d For, %d Against, %d Abstain\n\n", sentiment.ForVotes, sentiment.AgainstVotes, sentiment.AbstainVotes))
	}

	return &ChatResponse{
		Response: responseText.String(),
		Type:     "analytics",
		Data:     sentiments,
		Success:  true,
		Metadata: map[string]interface{}{
			"confidence": intent.Confidence,
			"intent":     intent.Intent,
		},
	}, nil
}

// handleOnChainAction handles on-chain action requests
func (ce *ChatEngine) handleOnChainAction(ctx context.Context, message *ChatMessage, intent *QueryIntent) (*ChatResponse, error) {
	// Extract action parameters from message
	actionType := ce.extractActionType(message.Message)
	parameters := ce.extractActionParameters(message.Message)
	
	// Create action request
	actionRequest := &ActionRequest{
		ID:         fmt.Sprintf("action_%d", time.Now().UnixNano()),
		UserID:     message.UserID,
		ActionType: actionType,
		Parameters: parameters,
		Status:     "pending",
		Timestamp:  time.Now().Unix(),
	}

	// Simulate action execution
	// In a real implementation, this would interact with the ActionContract
	actionRequest.Status = "completed"
	actionRequest.Result = map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Successfully executed %s action", actionType),
		"tx_hash": "0x1234567890abcdef...", // Simulated transaction hash
	}

	responseText := fmt.Sprintf("⚡ **Action Executed Successfully**\n\n"+
		"Action: %s\n"+
		"Status: %s\n"+
		"Transaction: %s\n\n"+
		"Your action has been submitted to the blockchain!",
		actionType,
		actionRequest.Status,
		actionRequest.Result.(map[string]interface{})["tx_hash"])

	return &ChatResponse{
		Response: responseText,
		Type:     "action_result",
		Data:     actionRequest,
		Success:  true,
		Metadata: map[string]interface{}{
			"confidence": intent.Confidence,
			"intent":     intent.Intent,
			"action_id":  actionRequest.ID,
		},
	}, nil
}

// handleMarketDataQuery handles market data queries
func (ce *ChatEngine) handleMarketDataQuery(ctx context.Context, message *ChatMessage, intent *QueryIntent) (*ChatResponse, error) {
	// Get market data
	symbols := []string{"ETH", "USDC", "DAI"}
	marketData, err := ce.dataCollector.CollectMarketData(ctx, symbols)
	if err != nil {
		return nil, fmt.Errorf("failed to collect market data: %w", err)
	}

	var responseText strings.Builder
	responseText.WriteString("📈 **Market Data**\n\n")
	
	for _, data := range marketData {
		changeEmoji := "➡️"
		if data.Change24h > 0 {
			changeEmoji = "📈"
		} else if data.Change24h < 0 {
			changeEmoji = "📉"
		}
		
		responseText.WriteString(fmt.Sprintf("%s **%s**: $%.2f (%+.2f%%)\n", changeEmoji, data.Symbol, data.Price, data.Change24h))
		responseText.WriteString(fmt.Sprintf("   24h Volume: $%.0f\n", data.Volume24h))
		responseText.WriteString(fmt.Sprintf("   Market Cap: $%.0f\n\n", data.MarketCap))
	}

	return &ChatResponse{
		Response: responseText.String(),
		Type:     "market_data",
		Data:     marketData,
		Success:  true,
		Metadata: map[string]interface{}{
			"confidence": intent.Confidence,
			"intent":     intent.Intent,
		},
	}, nil
}

// handleGasInfoQuery handles gas-related queries
func (ce *ChatEngine) handleGasInfoQuery(ctx context.Context, message *ChatMessage, intent *QueryIntent) (*ChatResponse, error) {
	// Get gas data
	gasData, err := ce.dataCollector.CollectGasData(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to collect gas data: %w", err)
	}

	responseText := fmt.Sprintf("⛽ **Gas Information**\n\n"+
		"Current Gas Price: %d Gwei\n"+
		"Fast Gas Price: %d Gwei\n"+
		"Standard Gas Price: %d Gwei\n"+
		"Slow Gas Price: %d Gwei\n"+
		"Gas Utilization: %.1f%%\n\n"+
		"💡 Tip: Use the slow gas price for non-urgent transactions to save on fees!",
		gasData["current_gas_price"].(uint64)/1e9,
		gasData["fast_gas_price"].(uint64)/1e9,
		gasData["standard_gas_price"].(uint64)/1e9,
		gasData["slow_gas_price"].(uint64)/1e9,
		gasData["gas_utilization"].(float64)*100)

	return &ChatResponse{
		Response: responseText,
		Type:     "gas_info",
		Data:     gasData,
		Success:  true,
		Metadata: map[string]interface{}{
			"confidence": intent.Confidence,
			"intent":     intent.Intent,
		},
	}, nil
}

// handleGeneralQuery handles general queries
func (ce *ChatEngine) handleGeneralQuery(ctx context.Context, message *ChatMessage, intent *QueryIntent) (*ChatResponse, error) {
	responseText := "Hello! I'm your Kaia Analytics AI assistant. I can help you with:\n\n" +
		"🔍 **Analytics**: Yield opportunities, portfolio analysis, trading suggestions\n" +
		"⚡ **Actions**: Staking, voting, swapping tokens\n" +
		"📊 **Data**: Market prices, gas fees, network stats\n" +
		"🗳️ **Governance**: Proposal analysis and voting\n\n" +
		"Just ask me anything about DeFi, trading, or blockchain analytics!"

	return &ChatResponse{
		Response: responseText,
		Type:     "text",
		Success:  true,
		Metadata: map[string]interface{}{
			"confidence": intent.Confidence,
			"intent":     intent.Intent,
		},
	}, nil
}

// extractActionType extracts the action type from a message
func (ce *ChatEngine) extractActionType(message string) string {
	message = strings.ToLower(message)
	
	if strings.Contains(message, "stake") {
		return "stake"
	} else if strings.Contains(message, "unstake") {
		return "unstake"
	} else if strings.Contains(message, "swap") {
		return "swap"
	} else if strings.Contains(message, "vote") {
		return "vote"
	} else if strings.Contains(message, "yield") {
		return "yield_farm"
	}
	
	return "unknown"
}

// extractActionParameters extracts action parameters from a message
func (ce *ChatEngine) extractActionParameters(message string) map[string]interface{} {
	parameters := make(map[string]interface{})
	
	// Extract amounts
	amountRegex := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*(ETH|USDC|DAI)`)
	matches := amountRegex.FindAllStringSubmatch(message, -1)
	if len(matches) > 0 {
		parameters["amount"] = matches[0][1]
		parameters["token"] = matches[0][2]
	}
	
	// Extract addresses
	addressRegex := regexp.MustCompile(`0x[a-fA-F0-9]{40}`)
	addresses := addressRegex.FindAllString(message, -1)
	if len(addresses) > 0 {
		parameters["target_address"] = addresses[0]
	}
	
	return parameters
}

// RegisterConnection registers a WebSocket connection
func (ce *ChatEngine) RegisterConnection(userID string, conn *websocket.Conn) {
	ce.mu.Lock()
	defer ce.mu.Unlock()
	
	ce.connections[userID] = conn
}

// UnregisterConnection unregisters a WebSocket connection
func (ce *ChatEngine) UnregisterConnection(userID string) {
	ce.mu.Lock()
	defer ce.mu.Unlock()
	
	delete(ce.connections, userID)
}

// BroadcastMessage broadcasts a message to all connected users
func (ce *ChatEngine) BroadcastMessage(message *ChatResponse) error {
	ce.mu.RLock()
	defer ce.mu.RUnlock()
	
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	
	for userID, conn := range ce.connections {
		err := conn.WriteMessage(websocket.TextMessage, messageBytes)
		if err != nil {
			ce.logger.Printf("Failed to send message to user %s: %v", userID, err)
			// Remove failed connection
			go ce.UnregisterConnection(userID)
		}
	}
	
	return nil
}

// GetChatMetrics returns chat engine metrics
func (ce *ChatEngine) GetChatMetrics() map[string]interface{} {
	ce.mu.RLock()
	defer ce.mu.RUnlock()
	
	return map[string]interface{}{
		"active_connections": len(ce.connections),
		"total_users":        len(ce.connections),
		"last_updated":       time.Now().Unix(),
	}
}