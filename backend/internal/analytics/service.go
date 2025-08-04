package analytics

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"time"

	"kaia-analytics-ai/internal/contracts"
	"kaia-analytics-ai/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/panjf2000/ants/v2"
	"github.com/sirupsen/logrus"
	"gonum.org/v1/gonum/stat"
)

// Service handles analytics processing and generation
type Service struct {
	config          *config.Config
	db              *sql.DB
	redis           *redis.Client
	contractManager *contracts.Manager
	logger          *logrus.Logger
	workerPool      *ants.Pool
}

// AnalyticsResult represents the result of an analytics computation
type AnalyticsResult struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	TTL       time.Duration          `json:"ttl"`
}

// YieldAnalysis represents yield farming analysis
type YieldAnalysis struct {
	Protocol      string  `json:"protocol"`
	TokenPair     string  `json:"token_pair"`
	APY           float64 `json:"apy"`
	TVL           float64 `json:"tvl"`
	RiskScore     int     `json:"risk_score"`
	Category      string  `json:"category"`
	Recommendation string `json:"recommendation"`
	Confidence    float64 `json:"confidence"`
}

// TradingSuggestion represents AI-generated trading suggestions
type TradingSuggestion struct {
	TokenPair     string  `json:"token_pair"`
	Action        string  `json:"action"` // "buy", "sell", "hold"
	Confidence    float64 `json:"confidence"`
	PriceTarget   float64 `json:"price_target"`
	StopLoss      float64 `json:"stop_loss"`
	Reasoning     string  `json:"reasoning"`
	TimeHorizon   string  `json:"time_horizon"`
	RiskLevel     string  `json:"risk_level"`
}

// GovernanceAnalysis represents governance sentiment analysis
type GovernanceAnalysis struct {
	ProposalID        string  `json:"proposal_id"`
	Title             string  `json:"title"`
	SentimentScore    float64 `json:"sentiment_score"`
	ParticipationRate float64 `json:"participation_rate"`
	Outcome           string  `json:"predicted_outcome"`
	KeyTopics         []string `json:"key_topics"`
	Community         string  `json:"community_sentiment"`
}

// MarketTrend represents market trend analysis
type MarketTrend struct {
	Asset         string    `json:"asset"`
	Trend         string    `json:"trend"` // "bullish", "bearish", "sideways"
	Strength      float64   `json:"strength"`
	Duration      string    `json:"duration"`
	KeyIndicators []string  `json:"key_indicators"`
	LastUpdated   time.Time `json:"last_updated"`
}

// NewService creates a new analytics service
func NewService(
	config *config.Config,
	db *sql.DB,
	redis *redis.Client,
	contractManager *contracts.Manager,
	logger *logrus.Logger,
) *Service {
	workerPool, _ := ants.NewPool(config.WorkerPoolSize)
	
	return &Service{
		config:          config,
		db:              db,
		redis:           redis,
		contractManager: contractManager,
		logger:          logger.WithField("service", "analytics"),
		workerPool:      workerPool,
	}
}

// Start starts the analytics service
func (s *Service) Start(ctx context.Context) error {
	s.logger.Info("Starting Analytics Engine")

	// Start background processors
	go s.processAnalyticsTasks(ctx)
	go s.updateYieldAnalysis(ctx)
	go s.generateTradingSuggestions(ctx)
	go s.analyzeGovernanceData(ctx)

	<-ctx.Done()
	s.logger.Info("Analytics Engine stopped")
	return nil
}

// HTTP Handlers

// GetYieldOpportunities returns yield farming opportunities
func (s *Service) GetYieldOpportunities(c *gin.Context) {
	protocol := c.Query("protocol")
	category := c.Query("category")
	minAPY := c.Query("min_apy")
	maxRisk := c.Query("max_risk")

	// Get cached results first
	cacheKey := fmt.Sprintf("yield_opportunities:%s:%s:%s:%s", protocol, category, minAPY, maxRisk)
	cached, err := s.redis.Get(c.Request.Context(), cacheKey).Result()
	if err == nil {
		var opportunities []YieldAnalysis
		if json.Unmarshal([]byte(cached), &opportunities) == nil {
			c.JSON(http.StatusOK, gin.H{
				"data":      opportunities,
				"cached":    true,
				"timestamp": time.Now(),
			})
			return
		}
	}

	// Generate fresh analysis
	opportunities, err := s.analyzeYieldOpportunities(c.Request.Context(), protocol, category, minAPY, maxRisk)
	if err != nil {
		s.logger.WithError(err).Error("Failed to analyze yield opportunities")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze yield opportunities"})
		return
	}

	// Cache results
	if data, err := json.Marshal(opportunities); err == nil {
		s.redis.Set(c.Request.Context(), cacheKey, data, 5*time.Minute)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      opportunities,
		"cached":    false,
		"timestamp": time.Now(),
	})
}

// GetTradingSuggestions returns AI-generated trading suggestions
func (s *Service) GetTradingSuggestions(c *gin.Context) {
	userAddress := c.Query("user")
	timeHorizon := c.Query("time_horizon")
	riskLevel := c.Query("risk_level")

	suggestions, err := s.generateUserTradingSuggestions(c.Request.Context(), userAddress, timeHorizon, riskLevel)
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate trading suggestions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate trading suggestions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      suggestions,
		"timestamp": time.Now(),
	})
}

// GetGovernanceData returns governance sentiment analysis
func (s *Service) GetGovernanceData(c *gin.Context) {
	category := c.Query("category")
	limit := c.Query("limit")

	limitInt := 10
	if limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			limitInt = l
		}
	}

	analysis, err := s.getGovernanceAnalysis(c.Request.Context(), category, limitInt)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get governance analysis")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get governance analysis"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      analysis,
		"timestamp": time.Now(),
	})
}

// GetMarketTrends returns market trend analysis
func (s *Service) GetMarketTrends(c *gin.Context) {
	asset := c.Query("asset")
	timeframe := c.Query("timeframe")

	trends, err := s.analyzeMarketTrends(c.Request.Context(), asset, timeframe)
	if err != nil {
		s.logger.WithError(err).Error("Failed to analyze market trends")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze market trends"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      trends,
		"timestamp": time.Now(),
	})
}

// HandleCustomQuery processes custom analytics queries
func (s *Service) HandleCustomQuery(c *gin.Context) {
	var request struct {
		Query      string                 `json:"query"`
		Parameters map[string]interface{} `json:"parameters"`
		UserID     string                 `json:"user_id"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	result, err := s.processCustomQuery(c.Request.Context(), request.Query, request.Parameters, request.UserID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to process custom query")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process custom query"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      result,
		"timestamp": time.Now(),
	})
}

// Core Analytics Methods

// analyzeYieldOpportunities performs yield farming analysis
func (s *Service) analyzeYieldOpportunities(ctx context.Context, protocol, category, minAPY, maxRisk string) ([]YieldAnalysis, error) {
	// Get yield data from contracts
	var opportunities []YieldAnalysis

	// Mock implementation - in production, this would query real data
	opportunities = []YieldAnalysis{
		{
			Protocol:       "KaiaSwap",
			TokenPair:      "KAIA/USDC",
			APY:            12.5,
			TVL:            1500000,
			RiskScore:      25,
			Category:       "farming",
			Recommendation: "Strong Buy",
			Confidence:     0.85,
		},
		{
			Protocol:       "KaiaLend",
			TokenPair:      "KAIA",
			APY:            8.2,
			TVL:            5000000,
			RiskScore:      15,
			Category:       "lending",
			Recommendation: "Buy",
			Confidence:     0.75,
		},
	}

	// Apply filters
	filtered := s.filterYieldOpportunities(opportunities, protocol, category, minAPY, maxRisk)
	
	// Sort by APY descending
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].APY > filtered[j].APY
	})

	return filtered, nil
}

// generateUserTradingSuggestions creates personalized trading suggestions
func (s *Service) generateUserTradingSuggestions(ctx context.Context, userAddress, timeHorizon, riskLevel string) ([]TradingSuggestion, error) {
	// In production, this would:
	// 1. Analyze user's trading history
	// 2. Consider market conditions
	// 3. Apply ML models for predictions
	// 4. Generate personalized suggestions

	suggestions := []TradingSuggestion{
		{
			TokenPair:     "KAIA/USDC",
			Action:        "buy",
			Confidence:    0.78,
			PriceTarget:   1.25,
			StopLoss:      0.95,
			Reasoning:     "Strong technical indicators and positive market sentiment",
			TimeHorizon:   "1-2 weeks",
			RiskLevel:     "medium",
		},
		{
			TokenPair:     "ETH/KAIA",
			Action:        "hold",
			Confidence:    0.65,
			PriceTarget:   0.0,
			StopLoss:      0.0,
			Reasoning:     "Consolidation phase, wait for breakout confirmation",
			TimeHorizon:   "1 month",
			RiskLevel:     "low",
		},
	}

	return suggestions, nil
}

// getGovernanceAnalysis retrieves governance sentiment analysis
func (s *Service) getGovernanceAnalysis(ctx context.Context, category string, limit int) ([]GovernanceAnalysis, error) {
	// Mock implementation
	analysis := []GovernanceAnalysis{
		{
			ProposalID:        "KIP-001",
			Title:             "Increase Block Gas Limit",
			SentimentScore:    0.75,
			ParticipationRate: 0.68,
			Outcome:           "likely_pass",
			KeyTopics:         []string{"scalability", "gas_fees", "performance"},
			Community:         "positive",
		},
	}

	return analysis, nil
}

// analyzeMarketTrends performs market trend analysis
func (s *Service) analyzeMarketTrends(ctx context.Context, asset, timeframe string) ([]MarketTrend, error) {
	// Mock implementation
	trends := []MarketTrend{
		{
			Asset:         "KAIA",
			Trend:         "bullish",
			Strength:      0.72,
			Duration:      "2 weeks",
			KeyIndicators: []string{"RSI", "MACD", "Volume"},
			LastUpdated:   time.Now(),
		},
	}

	return trends, nil
}

// processCustomQuery handles custom analytics queries
func (s *Service) processCustomQuery(ctx context.Context, query string, parameters map[string]interface{}, userID string) (interface{}, error) {
	// This would implement a query engine for custom analytics
	// For now, return a mock response
	return map[string]interface{}{
		"query":      query,
		"parameters": parameters,
		"result":     "Custom query processed successfully",
		"user_id":    userID,
	}, nil
}

// Background Processing Methods

// processAnalyticsTasks processes pending analytics tasks from the registry
func (s *Service) processAnalyticsTasks(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.processPendingTasks(ctx)
		}
	}
}

// processPendingTasks processes all pending tasks
func (s *Service) processPendingTasks(ctx context.Context) {
	tasks, err := s.contractManager.GetPendingTasks(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get pending tasks")
		return
	}

	for _, task := range tasks {
		s.workerPool.Submit(func() {
			if err := s.processTask(ctx, task); err != nil {
				s.logger.WithError(err).WithField("task_id", task.ID).Error("Failed to process task")
			}
		})
	}
}

// processTask processes a single analytics task
func (s *Service) processTask(ctx context.Context, task *contracts.AnalyticsTask) error {
	s.logger.WithField("task_id", task.ID).WithField("type", task.TaskType).Info("Processing analytics task")

	var result interface{}
	var err error

	switch task.TaskType {
	case "yield_analysis":
		result, err = s.processYieldAnalysisTask(ctx, task)
	case "governance_sentiment":
		result, err = s.processGovernanceSentimentTask(ctx, task)
	case "trade_optimization":
		result, err = s.processTradeOptimizationTask(ctx, task)
	default:
		return fmt.Errorf("unknown task type: %s", task.TaskType)
	}

	if err != nil {
		return fmt.Errorf("failed to process task: %w", err)
	}

	// Store result and mark task as completed
	resultData, _ := json.Marshal(result)
	resultHash := fmt.Sprintf("result_%s_%d", task.TaskType, task.ID.Int64())
	
	if err := s.contractManager.CompleteTask(ctx, task.ID, resultHash); err != nil {
		return fmt.Errorf("failed to complete task: %w", err)
	}

	// Cache result
	s.redis.Set(ctx, resultHash, resultData, 24*time.Hour)

	return nil
}

// Task processors
func (s *Service) processYieldAnalysisTask(ctx context.Context, task *contracts.AnalyticsTask) (interface{}, error) {
	// Process yield analysis task
	return s.analyzeYieldOpportunities(ctx, "", "", "", "")
}

func (s *Service) processGovernanceSentimentTask(ctx context.Context, task *contracts.AnalyticsTask) (interface{}, error) {
	// Process governance sentiment task
	return s.getGovernanceAnalysis(ctx, "", 10)
}

func (s *Service) processTradeOptimizationTask(ctx context.Context, task *contracts.AnalyticsTask) (interface{}, error) {
	// Process trade optimization task
	return s.generateUserTradingSuggestions(ctx, "", "", "")
}

// Periodic update methods

// updateYieldAnalysis updates yield analysis data periodically
func (s *Service) updateYieldAnalysis(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.refreshYieldData(ctx); err != nil {
				s.logger.WithError(err).Error("Failed to refresh yield data")
			}
		}
	}
}

// generateTradingSuggestions generates trading suggestions periodically
func (s *Service) generateTradingSuggestions(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.refreshTradingSuggestions(ctx); err != nil {
				s.logger.WithError(err).Error("Failed to refresh trading suggestions")
			}
		}
	}
}

// analyzeGovernanceData analyzes governance data periodically
func (s *Service) analyzeGovernanceData(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.refreshGovernanceAnalysis(ctx); err != nil {
				s.logger.WithError(err).Error("Failed to refresh governance analysis")
			}
		}
	}
}

// Helper methods

// filterYieldOpportunities filters yield opportunities based on criteria
func (s *Service) filterYieldOpportunities(opportunities []YieldAnalysis, protocol, category, minAPY, maxRisk string) []YieldAnalysis {
	var filtered []YieldAnalysis

	for _, opp := range opportunities {
		// Apply filters
		if protocol != "" && opp.Protocol != protocol {
			continue
		}
		if category != "" && opp.Category != category {
			continue
		}
		if minAPY != "" {
			if min, err := strconv.ParseFloat(minAPY, 64); err == nil && opp.APY < min {
				continue
			}
		}
		if maxRisk != "" {
			if max, err := strconv.Atoi(maxRisk); err == nil && opp.RiskScore > max {
				continue
			}
		}

		filtered = append(filtered, opp)
	}

	return filtered
}

// calculateRiskScore calculates risk score for yield opportunities
func (s *Service) calculateRiskScore(tvl, apy float64, protocol string) int {
	// Simple risk scoring algorithm
	score := 50 // Base score

	// TVL factor (higher TVL = lower risk)
	if tvl > 10000000 {
		score -= 20
	} else if tvl > 1000000 {
		score -= 10
	}

	// APY factor (higher APY = higher risk)
	if apy > 20 {
		score += 20
	} else if apy > 10 {
		score += 10
	}

	// Protocol factor (established protocols = lower risk)
	establishedProtocols := []string{"KaiaSwap", "KaiaLend"}
	for _, p := range establishedProtocols {
		if protocol == p {
			score -= 10
			break
		}
	}

	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}

	return score
}

// calculateMovingAverage calculates moving average for trend analysis
func (s *Service) calculateMovingAverage(prices []float64, period int) []float64 {
	if len(prices) < period {
		return nil
	}

	var ma []float64
	for i := period - 1; i < len(prices); i++ {
		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			sum += prices[j]
		}
		ma = append(ma, sum/float64(period))
	}

	return ma
}

// calculateVolatility calculates price volatility
func (s *Service) calculateVolatility(prices []float64) float64 {
	if len(prices) < 2 {
		return 0
	}

	var returns []float64
	for i := 1; i < len(prices); i++ {
		returns = append(returns, (prices[i]-prices[i-1])/prices[i-1])
	}

	mean := stat.Mean(returns, nil)
	variance := stat.Variance(returns, nil)
	
	return math.Sqrt(variance) * math.Sqrt(252) // Annualized volatility
}

// Refresh methods for periodic updates

func (s *Service) refreshYieldData(ctx context.Context) error {
	s.logger.Debug("Refreshing yield data")
	// Implementation would fetch fresh data and update cache
	return nil
}

func (s *Service) refreshTradingSuggestions(ctx context.Context) error {
	s.logger.Debug("Refreshing trading suggestions")
	// Implementation would generate fresh suggestions
	return nil
}

func (s *Service) refreshGovernanceAnalysis(ctx context.Context) error {
	s.logger.Debug("Refreshing governance analysis")
	// Implementation would analyze latest governance data
	return nil
}