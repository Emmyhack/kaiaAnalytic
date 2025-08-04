package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"gonum.org/v1/gonum/stat"
)

// AnalyticsEngine generates actionable analytics from raw data
type AnalyticsEngine struct {
	logger         *logrus.Logger
	workerPool     *ants.Pool
	llm            llms.LLM
	dataCollector  *DataCollector
	mu             sync.RWMutex
	analytics      map[string]*AnalyticsResult
	yieldCache     map[string]*YieldAnalysis
	tradeCache     map[string]*TradeOptimization
	sentimentCache map[string]*SentimentAnalysis
}

type AnalyticsResult struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Data        interface{}            `json:"data"`
	Confidence  float64                `json:"confidence"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata"`
	Expires     time.Time              `json:"expires"`
}

type YieldAnalysis struct {
	Protocol        string  `json:"protocol"`
	PoolAddress     string  `json:"pool_address"`
	APR             float64 `json:"apr"`
	APY             float64 `json:"apy"`
	TVL             float64 `json:"tvl"`
	RiskScore       int     `json:"risk_score"`
	Strategy        string  `json:"strategy"`
	MinDeposit      float64 `json:"min_deposit"`
	ImpermanentLoss float64 `json:"impermanent_loss_risk"`
	Recommendation  string    `json:"recommendation"`
	Confidence      float64   `json:"confidence"`
	AnalysisDate    time.Time `json:"analysis_date"`
}

type TradeOptimization struct {
	TokenPair       string    `json:"token_pair"`
	Action          string    `json:"action"` // "buy", "sell", "hold"
	TargetPrice     float64   `json:"target_price"`
	CurrentPrice    float64   `json:"current_price"`
	PriceChange24h  float64   `json:"price_change_24h"`
	Volume24h       float64   `json:"volume_24h"`
	RSI             float64   `json:"rsi"`
	MACD            float64   `json:"macd"`
	BollingerBands  []float64 `json:"bollinger_bands"`
	SupportLevels   []float64 `json:"support_levels"`
	ResistanceLevels []float64 `json:"resistance_levels"`
	Recommendation  string    `json:"recommendation"`
	RiskLevel       string    `json:"risk_level"`
	Confidence      float64   `json:"confidence"`
	OptimalAmount   float64   `json:"optimal_amount"`
	GasCost         float64   `json:"estimated_gas_cost"`
	AnalysisDate    time.Time `json:"analysis_date"`
}

type SentimentAnalysis struct {
	ProposalID       string    `json:"proposal_id"`
	ProposalTitle    string    `json:"proposal_title"`
	SentimentScore   float64   `json:"sentiment_score"` // -1.0 to 1.0
	ParticipationRate float64  `json:"participation_rate"`
	VotingPower      float64   `json:"voting_power"`
	PredictedOutcome string    `json:"predicted_outcome"`
	KeyTopics        []string  `json:"key_topics"`
	Concerns         []string  `json:"concerns"`
	Support          []string  `json:"support"`
	Confidence       float64   `json:"confidence"`
	AnalysisDate     time.Time `json:"analysis_date"`
}

type MarketData struct {
	Price          float64   `json:"price"`
	Volume         float64   `json:"volume"`
	MarketCap      float64   `json:"market_cap"`
	PriceChange24h float64   `json:"price_change_24h"`
	Timestamp      time.Time `json:"timestamp"`
}

// NewAnalyticsEngine creates a new analytics engine
func NewAnalyticsEngine(logger *logrus.Logger, dataCollector *DataCollector, openaiKey string) (*AnalyticsEngine, error) {
	// Create worker pool for concurrent processing
	pool, err := ants.NewPool(10)
	if err != nil {
		return nil, fmt.Errorf("failed to create worker pool: %w", err)
	}

	// Initialize LLM for sentiment analysis
	var llm llms.LLM
	if openaiKey != "" {
		llm, err = openai.New(openai.WithToken(openaiKey))
		if err != nil {
			logger.WithError(err).Warn("Failed to initialize OpenAI client, sentiment analysis will be limited")
		}
	}

	engine := &AnalyticsEngine{
		logger:         logger,
		workerPool:     pool,
		llm:            llm,
		dataCollector:  dataCollector,
		analytics:      make(map[string]*AnalyticsResult),
		yieldCache:     make(map[string]*YieldAnalysis),
		tradeCache:     make(map[string]*TradeOptimization),
		sentimentCache: make(map[string]*SentimentAnalysis),
	}

	// Start background workers
	go engine.startBackgroundAnalysis()
	go engine.cleanupExpiredCache()

	return engine, nil
}

// AnalyzeYieldOpportunities identifies and analyzes yield opportunities
func (ae *AnalyticsEngine) AnalyzeYieldOpportunities(ctx context.Context, protocols []string) ([]*YieldAnalysis, error) {
	ae.logger.Info("Starting yield opportunity analysis")

	var results []*YieldAnalysis
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, protocol := range protocols {
		wg.Add(1)
		ae.workerPool.Submit(func() {
			defer wg.Done()

			analysis, err := ae.analyzeProtocolYield(ctx, protocol)
			if err != nil {
				ae.logger.WithError(err).WithField("protocol", protocol).Error("Failed to analyze protocol yield")
				return
			}

			mu.Lock()
			results = append(results, analysis)
			ae.yieldCache[protocol] = analysis
			mu.Unlock()
		})
	}

	wg.Wait()

	// Sort by APR descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].APR > results[j].APR
	})

	ae.logger.WithField("count", len(results)).Info("Completed yield opportunity analysis")
	return results, nil
}

// OptimizeTradingStrategy generates optimized trading suggestions
func (ae *AnalyticsEngine) OptimizeTradingStrategy(ctx context.Context, userAddress string, tokenPair string) (*TradeOptimization, error) {
	ae.logger.WithFields(logrus.Fields{
		"user":      userAddress,
		"tokenPair": tokenPair,
	}).Info("Optimizing trading strategy")

	// Check cache first
	ae.mu.RLock()
	if cached, exists := ae.tradeCache[tokenPair]; exists && time.Since(cached.AnalysisDate) < 5*time.Minute {
		ae.mu.RUnlock()
		return cached, nil
	}
	ae.mu.RUnlock()

	// Get historical data
	historicalData, err := ae.dataCollector.GetHistoricalPrices(ctx, tokenPair, 30) // 30 days
	if err != nil {
		return nil, fmt.Errorf("failed to get historical data: %w", err)
	}

	// Get user's trading history for personalization
	userHistory, err := ae.dataCollector.GetUserTradeHistory(ctx, userAddress)
	if err != nil {
		ae.logger.WithError(err).Warn("Failed to get user history, using generic analysis")
	}

	optimization := ae.calculateTradeOptimization(historicalData, userHistory, tokenPair)

	// Cache the result
	ae.mu.Lock()
	ae.tradeCache[tokenPair] = optimization
	ae.mu.Unlock()

	return optimization, nil
}

// AnalyzeGovernanceSentiment analyzes sentiment of governance proposals
func (ae *AnalyticsEngine) AnalyzeGovernanceSentiment(ctx context.Context, proposalID string) (*SentimentAnalysis, error) {
	ae.logger.WithField("proposalID", proposalID).Info("Analyzing governance sentiment")

	// Check cache first
	ae.mu.RLock()
	if cached, exists := ae.sentimentCache[proposalID]; exists && time.Since(cached.AnalysisDate) < 1*time.Hour {
		ae.mu.RUnlock()
		return cached, nil
	}
	ae.mu.RUnlock()

	// Get proposal data
	proposalData, err := ae.dataCollector.GetGovernanceProposal(ctx, proposalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposal data: %w", err)
	}

	// Get community discussions
	discussions, err := ae.dataCollector.GetProposalDiscussions(ctx, proposalID)
	if err != nil {
		ae.logger.WithError(err).Warn("Failed to get discussions, using basic analysis")
	}

	sentiment := ae.analyzeSentiment(proposalData, discussions)

	// Cache the result
	ae.mu.Lock()
	ae.sentimentCache[proposalID] = sentiment
	ae.mu.Unlock()

	return sentiment, nil
}

// GetAnalyticsResult retrieves a cached analytics result
func (ae *AnalyticsEngine) GetAnalyticsResult(id string) (*AnalyticsResult, bool) {
	ae.mu.RLock()
	defer ae.mu.RUnlock()
	
	result, exists := ae.analytics[id]
	if exists && time.Now().Before(result.Expires) {
		return result, true
	}
	
	return nil, false
}

// analyzeProtocolYield analyzes yield opportunities for a specific protocol
func (ae *AnalyticsEngine) analyzeProtocolYield(ctx context.Context, protocol string) (*YieldAnalysis, error) {
	// Get protocol data
	protocolData, err := ae.dataCollector.GetProtocolData(ctx, protocol)
	if err != nil {
		return nil, fmt.Errorf("failed to get protocol data: %w", err)
	}

	// Calculate APR and APY
	apr := ae.calculateAPR(protocolData)
	apy := ae.calculateAPY(apr)

	// Assess risk
	riskScore := ae.assessRisk(protocolData)

	// Calculate impermanent loss risk for LP tokens
	impermanentLossRisk := ae.calculateImpermanentLossRisk(protocolData)

	// Generate recommendation
	recommendation := ae.generateYieldRecommendation(apr, float64(riskScore), impermanentLossRisk)

	analysis := &YieldAnalysis{
		Protocol:        protocol,
		PoolAddress:     protocolData["pool_address"].(string),
		APR:             apr,
		APY:             apy,
		TVL:             protocolData["tvl"].(float64),
		RiskScore:       riskScore,
		Strategy:        protocolData["strategy"].(string),
		MinDeposit:      protocolData["min_deposit"].(float64),
		ImpermanentLoss: impermanentLossRisk,
		Recommendation:  recommendation,
		Confidence:      ae.calculateConfidence(protocolData),
		AnalysisDate:    time.Now(),
	}

	return analysis, nil
}

// calculateTradeOptimization performs technical analysis
func (ae *AnalyticsEngine) calculateTradeOptimization(historicalData []MarketData, userHistory interface{}, tokenPair string) *TradeOptimization {
	if len(historicalData) < 14 {
		// Not enough data for proper analysis
		return &TradeOptimization{
			TokenPair:      tokenPair,
			Action:         "hold",
			Recommendation: "Insufficient historical data for analysis",
			Confidence:     0.1,
		}
	}

	prices := make([]float64, len(historicalData))
	volumes := make([]float64, len(historicalData))
	
	for i, data := range historicalData {
		prices[i] = data.Price
		volumes[i] = data.Volume
	}

	currentPrice := prices[len(prices)-1]
	
	// Calculate technical indicators
	rsi := ae.calculateRSI(prices, 14)
	macd := ae.calculateMACD(prices)
	bollingerBands := ae.calculateBollingerBands(prices, 20)
	supportLevels := ae.findSupportLevels(prices)
	resistanceLevels := ae.findResistanceLevels(prices)

	// Generate trading signal
	action := ae.generateTradingSignal(rsi, macd, currentPrice, bollingerBands)
	
	// Calculate target price
	targetPrice := ae.calculateTargetPrice(currentPrice, action, supportLevels, resistanceLevels)
	
	// Assess risk
	riskLevel := ae.assessTradingRisk(rsi, prices)
	
	// Calculate optimal position size
	optimalAmount := ae.calculateOptimalPositionSize(currentPrice, riskLevel, userHistory)

	return &TradeOptimization{
		TokenPair:        tokenPair,
		Action:           action,
		TargetPrice:      targetPrice,
		CurrentPrice:     currentPrice,
		PriceChange24h:   ((currentPrice - prices[len(prices)-2]) / prices[len(prices)-2]) * 100,
		Volume24h:        volumes[len(volumes)-1],
		RSI:              rsi,
		MACD:             macd,
		BollingerBands:   bollingerBands,
		SupportLevels:    supportLevels,
		ResistanceLevels: resistanceLevels,
		Recommendation:   ae.generateTradeRecommendation(action, rsi, macd),
		RiskLevel:        riskLevel,
		Confidence:       ae.calculateTradingConfidence(rsi, macd, len(historicalData)),
		OptimalAmount:    optimalAmount,
		GasCost:          ae.estimateGasCost(),
		AnalysisDate:     time.Now(),
	}
}

// analyzeSentiment performs sentiment analysis on governance proposals
func (ae *AnalyticsEngine) analyzeSentiment(proposalData map[string]interface{}, discussions []string) *SentimentAnalysis {
	proposalID := proposalData["id"].(string)
	proposalTitle := proposalData["title"].(string)

	// Basic sentiment analysis using keyword matching
	sentimentScore := ae.calculateBasicSentiment(discussions)
	
	// Use LLM for enhanced analysis if available
	if ae.llm != nil {
		enhancedScore, topics, concerns, support := ae.performLLMSentimentAnalysis(proposalTitle, discussions)
		if enhancedScore != 0 {
			sentimentScore = enhancedScore
		}
		
		return &SentimentAnalysis{
			ProposalID:        proposalID,
			ProposalTitle:     proposalTitle,
			SentimentScore:    sentimentScore,
			ParticipationRate: proposalData["participation_rate"].(float64),
			VotingPower:       proposalData["voting_power"].(float64),
			PredictedOutcome:  ae.predictOutcome(sentimentScore),
			KeyTopics:         topics,
			Concerns:          concerns,
			Support:           support,
			Confidence:        ae.calculateSentimentConfidence(len(discussions)),
			AnalysisDate:      time.Now(),
		}
	}

	return &SentimentAnalysis{
		ProposalID:        proposalID,
		ProposalTitle:     proposalTitle,
		SentimentScore:    sentimentScore,
		ParticipationRate: proposalData["participation_rate"].(float64),
		VotingPower:       proposalData["voting_power"].(float64),
		PredictedOutcome:  ae.predictOutcome(sentimentScore),
		KeyTopics:         []string{},
		Concerns:          []string{},
		Support:           []string{},
		Confidence:        ae.calculateSentimentConfidence(len(discussions)),
		AnalysisDate:      time.Now(),
	}
}

// Technical Analysis Functions

// calculateRSI calculates the Relative Strength Index
func (ae *AnalyticsEngine) calculateRSI(prices []float64, period int) float64 {
	if len(prices) < period+1 {
		return 50.0 // Neutral RSI
	}

	gains := 0.0
	losses := 0.0

	// Calculate initial average gain and loss
	for i := 1; i <= period; i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			gains += change
		} else {
			losses += math.Abs(change)
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	// Calculate subsequent averages using exponential smoothing
	for i := period + 1; i < len(prices); i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			avgGain = (avgGain*float64(period-1) + change) / float64(period)
			avgLoss = (avgLoss * float64(period-1)) / float64(period)
		} else {
			avgGain = (avgGain * float64(period-1)) / float64(period)
			avgLoss = (avgLoss*float64(period-1) + math.Abs(change)) / float64(period)
		}
	}

	if avgLoss == 0 {
		return 100.0
	}

	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))

	return rsi
}

// calculateMACD calculates the MACD indicator
func (ae *AnalyticsEngine) calculateMACD(prices []float64) float64 {
	if len(prices) < 26 {
		return 0.0
	}

	// Calculate 12-period EMA
	ema12 := ae.calculateEMA(prices, 12)
	// Calculate 26-period EMA
	ema26 := ae.calculateEMA(prices, 26)

	return ema12 - ema26
}

// calculateEMA calculates Exponential Moving Average
func (ae *AnalyticsEngine) calculateEMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return stat.Mean(prices, nil)
	}

	multiplier := 2.0 / (float64(period) + 1.0)
	ema := stat.Mean(prices[:period], nil)

	for i := period; i < len(prices); i++ {
		ema = (prices[i] * multiplier) + (ema * (1 - multiplier))
	}

	return ema
}

// calculateBollingerBands calculates Bollinger Bands
func (ae *AnalyticsEngine) calculateBollingerBands(prices []float64, period int) []float64 {
	if len(prices) < period {
		mean := stat.Mean(prices, nil)
		return []float64{mean, mean, mean} // middle, upper, lower
	}

	recentPrices := prices[len(prices)-period:]
	mean := stat.Mean(recentPrices, nil)
	stdDev := stat.StdDev(recentPrices, nil)

	upper := mean + (2 * stdDev)
	lower := mean - (2 * stdDev)

	return []float64{mean, upper, lower}
}

// Additional helper functions for comprehensive analysis
func (ae *AnalyticsEngine) findSupportLevels(prices []float64) []float64 {
	// Simple support level detection using local minima
	var supports []float64
	for i := 1; i < len(prices)-1; i++ {
		if prices[i] < prices[i-1] && prices[i] < prices[i+1] {
			supports = append(supports, prices[i])
		}
	}
	
	// Return top 3 support levels
	sort.Float64s(supports)
	if len(supports) > 3 {
		return supports[:3]
	}
	return supports
}

func (ae *AnalyticsEngine) findResistanceLevels(prices []float64) []float64 {
	// Simple resistance level detection using local maxima
	var resistance []float64
	for i := 1; i < len(prices)-1; i++ {
		if prices[i] > prices[i-1] && prices[i] > prices[i+1] {
			resistance = append(resistance, prices[i])
		}
	}
	
	// Return top 3 resistance levels
	sort.Sort(sort.Reverse(sort.Float64Slice(resistance)))
	if len(resistance) > 3 {
		return resistance[:3]
	}
	return resistance
}

// Additional utility functions
func (ae *AnalyticsEngine) calculateAPR(protocolData map[string]interface{}) float64 {
	// Mock APR calculation - in real implementation, this would use protocol-specific formulas
	baseRate := protocolData["base_rate"].(float64)
	utilization := protocolData["utilization"].(float64)
	return baseRate * (1 + utilization/100)
}

func (ae *AnalyticsEngine) calculateAPY(apr float64) float64 {
	// Convert APR to APY (compounding daily)
	return math.Pow(1+apr/365, 365) - 1
}

func (ae *AnalyticsEngine) assessRisk(protocolData map[string]interface{}) int {
	// Risk assessment based on various factors
	age := protocolData["age_days"].(float64)
	tvl := protocolData["tvl"].(float64)
	auditScore := protocolData["audit_score"].(float64)

	riskScore := 50 // Base risk
	
	if age > 365 {
		riskScore -= 10 // Mature protocol
	}
	if tvl > 100000000 { // $100M TVL
		riskScore -= 15
	}
	if auditScore > 80 {
		riskScore -= 10
	}

	if riskScore < 0 {
		riskScore = 0
	}
	if riskScore > 100 {
		riskScore = 100
	}

	return riskScore
}

func (ae *AnalyticsEngine) calculateImpermanentLossRisk(protocolData map[string]interface{}) float64 {
	// Calculate impermanent loss risk for LP tokens
	volatility := protocolData["volatility"].(float64)
	correlation := protocolData["correlation"].(float64)
	
	// High volatility and low correlation increase impermanent loss risk
	risk := volatility * (1 - correlation)
	return math.Min(risk, 100.0)
}

func (ae *AnalyticsEngine) generateYieldRecommendation(apr, riskScore float64, impermanentLoss float64) string {
	if apr > 20 && riskScore < 30 {
		return "Highly Recommended - High yield with acceptable risk"
	} else if apr > 10 && riskScore < 50 {
		return "Recommended - Good balance of yield and risk"
	} else if riskScore > 70 {
		return "High Risk - Proceed with caution"
	} else if apr < 5 {
		return "Low Yield - Consider other opportunities"
	}
	return "Moderate - Standard yield opportunity"
}

func (ae *AnalyticsEngine) generateTradingSignal(rsi, macd, currentPrice float64, bollingerBands []float64) string {
	signals := 0

	// RSI signals
	if rsi < 30 {
		signals++ // Oversold - buy signal
	} else if rsi > 70 {
		signals-- // Overbought - sell signal
	}

	// MACD signals
	if macd > 0 {
		signals++ // Positive MACD - buy signal
	} else {
		signals-- // Negative MACD - sell signal
	}

	// Bollinger Band signals
	if currentPrice < bollingerBands[2] { // Below lower band
		signals++ // Buy signal
	} else if currentPrice > bollingerBands[1] { // Above upper band
		signals-- // Sell signal
	}

	if signals > 0 {
		return "buy"
	} else if signals < 0 {
		return "sell"
	}
	return "hold"
}

func (ae *AnalyticsEngine) calculateTargetPrice(currentPrice float64, action string, supportLevels, resistanceLevels []float64) float64 {
	switch action {
	case "buy":
		if len(resistanceLevels) > 0 {
			return resistanceLevels[0] // Target first resistance
		}
		return currentPrice * 1.05 // 5% profit target
	case "sell":
		if len(supportLevels) > 0 {
			return supportLevels[0] // Target first support
		}
		return currentPrice * 0.95 // 5% stop loss
	default:
		return currentPrice
	}
}

func (ae *AnalyticsEngine) assessTradingRisk(rsi float64, prices []float64) string {
	volatility := stat.StdDev(prices[len(prices)-10:], nil) / stat.Mean(prices[len(prices)-10:], nil)
	
	if volatility > 0.1 || rsi > 80 || rsi < 20 {
		return "high"
	} else if volatility > 0.05 || rsi > 70 || rsi < 30 {
		return "medium"
	}
	return "low"
}

func (ae *AnalyticsEngine) calculateOptimalPositionSize(currentPrice float64, riskLevel string, userHistory interface{}) float64 {
	// Simple position sizing based on risk level
	baseAmount := 1000.0 // Base amount in USD
	
	switch riskLevel {
	case "low":
		return baseAmount / currentPrice
	case "medium":
		return (baseAmount * 0.7) / currentPrice
	case "high":
		return (baseAmount * 0.3) / currentPrice
	default:
		return baseAmount / currentPrice
	}
}

func (ae *AnalyticsEngine) generateTradeRecommendation(action string, rsi, macd float64) string {
	switch action {
	case "buy":
		return fmt.Sprintf("Consider buying - RSI: %.2f, MACD: %.2f indicates upward momentum", rsi, macd)
	case "sell":
		return fmt.Sprintf("Consider selling - RSI: %.2f, MACD: %.2f indicates downward momentum", rsi, macd)
	default:
		return "Hold position - Mixed signals, wait for clearer trend"
	}
}

func (ae *AnalyticsEngine) calculateTradingConfidence(rsi, macd float64, dataPoints int) float64 {
	confidence := 0.5 // Base confidence
	
	// More data points increase confidence
	confidence += float64(dataPoints) / 100.0
	
	// Strong RSI signals increase confidence
	if rsi < 20 || rsi > 80 {
		confidence += 0.2
	}
	
	// Strong MACD signals increase confidence
	if math.Abs(macd) > 0.05 {
		confidence += 0.1
	}
	
	return math.Min(confidence, 1.0)
}

func (ae *AnalyticsEngine) estimateGasCost() float64 {
	// Mock gas cost estimation - in real implementation, this would query current gas prices
	return 0.002 // ETH
}

func (ae *AnalyticsEngine) calculateBasicSentiment(discussions []string) float64 {
	if len(discussions) == 0 {
		return 0.0
	}

	positiveWords := []string{"support", "great", "excellent", "approve", "yes", "agree", "good", "positive"}
	negativeWords := []string{"against", "terrible", "bad", "no", "disagree", "reject", "oppose", "negative"}

	totalWords := 0
	sentimentSum := 0.0

	for _, discussion := range discussions {
		words := len(discussion) / 5 // Rough word count estimate
		totalWords += words

		for _, word := range positiveWords {
			if contains(discussion, word) {
				sentimentSum += 1.0
			}
		}
		for _, word := range negativeWords {
			if contains(discussion, word) {
				sentimentSum -= 1.0
			}
		}
	}

	if totalWords == 0 {
		return 0.0
	}

	return sentimentSum / float64(totalWords) * 100
}

func (ae *AnalyticsEngine) performLLMSentimentAnalysis(title string, discussions []string) (float64, []string, []string, []string) {
	if ae.llm == nil {
		return 0.0, nil, nil, nil
	}

	// Combine discussions into analysis prompt
	prompt := fmt.Sprintf("Analyze the sentiment of this governance proposal: '%s'\n\nDiscussions:\n%s\n\nProvide sentiment score (-1 to 1), key topics, concerns, and support points.", 
		title, joinStrings(discussions, "\n"))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := ae.llm.GenerateContent(ctx, []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart(prompt),
			},
		},
	})

	if err != nil {
		ae.logger.WithError(err).Error("Failed to analyze sentiment with LLM")
		return 0.0, nil, nil, nil
	}

	// Parse LLM response (simplified)
	sentimentScore := 0.0
	topics := []string{}
	concerns := []string{}
	support := []string{}

	// In a real implementation, you would parse the structured response
	// For now, return mock data
	if len(response.Choices) > 0 {
		content := response.Choices[0].Content
		if contains(content, "positive") {
			sentimentScore = 0.6
		} else if contains(content, "negative") {
			sentimentScore = -0.4
		}
	}

	return sentimentScore, topics, concerns, support
}

func (ae *AnalyticsEngine) predictOutcome(sentimentScore float64) string {
	if sentimentScore > 0.3 {
		return "Likely to Pass"
	} else if sentimentScore < -0.3 {
		return "Likely to Fail"
	}
	return "Uncertain"
}

func (ae *AnalyticsEngine) calculateConfidence(data map[string]interface{}) float64 {
	// Calculate confidence based on data quality and completeness
	dataPoints := len(data)
	if dataPoints > 10 {
		return 0.9
	} else if dataPoints > 5 {
		return 0.7
	}
	return 0.5
}

func (ae *AnalyticsEngine) calculateSentimentConfidence(discussionCount int) float64 {
	if discussionCount > 100 {
		return 0.9
	} else if discussionCount > 50 {
		return 0.7
	} else if discussionCount > 10 {
		return 0.5
	}
	return 0.3
}

// Background processes
func (ae *AnalyticsEngine) startBackgroundAnalysis() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// Refresh popular yield opportunities
		ae.refreshYieldAnalysis()
		
		// Update trading signals for popular pairs
		ae.refreshTradingSignals()
		
		// Update governance sentiment for active proposals
		ae.refreshGovernanceSentiment()
	}
}

func (ae *AnalyticsEngine) cleanupExpiredCache() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		ae.mu.Lock()
		
		// Clean expired analytics
		for id, result := range ae.analytics {
			if time.Now().After(result.Expires) {
				delete(ae.analytics, id)
			}
		}
		
		// Clean expired yield cache (older than 30 minutes)
		for protocol, analysis := range ae.yieldCache {
			if time.Since(analysis.AnalysisDate) > 30*time.Minute {
				delete(ae.yieldCache, protocol)
			}
		}
		
		// Clean expired trade cache (older than 5 minutes)
		for pair, optimization := range ae.tradeCache {
			if time.Since(optimization.AnalysisDate) > 5*time.Minute {
				delete(ae.tradeCache, pair)
			}
		}
		
		// Clean expired sentiment cache (older than 1 hour)
		for proposalID, sentiment := range ae.sentimentCache {
			if time.Since(sentiment.AnalysisDate) > 1*time.Hour {
				delete(ae.sentimentCache, proposalID)
			}
		}
		
		ae.mu.Unlock()
	}
}

func (ae *AnalyticsEngine) refreshYieldAnalysis() {
	// Refresh analysis for popular protocols
	popularProtocols := []string{"uniswap", "aave", "compound", "curve", "yearn"}
	ctx := context.Background()
	
	ae.AnalyzeYieldOpportunities(ctx, popularProtocols)
}

func (ae *AnalyticsEngine) refreshTradingSignals() {
	// Refresh signals for popular trading pairs
	popularPairs := []string{"ETH/USDC", "BTC/USDC", "KAIA/USDC", "KAIA/ETH"}
	ctx := context.Background()
	
	for _, pair := range popularPairs {
		ae.OptimizeTradingStrategy(ctx, "", pair) // Empty user address for general analysis
	}
}

func (ae *AnalyticsEngine) refreshGovernanceSentiment() {
	// This would refresh sentiment for active governance proposals
	// Implementation depends on governance data source
}

// Cleanup resources
func (ae *AnalyticsEngine) Close() {
	if ae.workerPool != nil {
		ae.workerPool.Release()
	}
}

// Utility functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}
	
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}