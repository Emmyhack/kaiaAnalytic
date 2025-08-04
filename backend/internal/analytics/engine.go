package analytics

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/panjf2000/ants/v2"
	"github.com/sirupsen/logrus"
	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/gonum/stat/distuv"
	"kaia-analytics-ai/internal/config"
	"kaia-analytics-ai/internal/contracts"
)

// Engine handles analytics computations
type Engine struct {
	config           *config.Config
	blockchainClient *contracts.BlockchainClient
	workerPool       *ants.Pool
	stopChan         chan struct{}
	mu               sync.RWMutex
	
	// Analytics data cache
	yieldData      []YieldOpportunity
	governanceData []GovernanceSentiment
	tradingData    []TradingSuggestion
	volumeData     []TransactionVolume
	gasData        []GasTrend
}

// YieldOpportunity represents a yield farming opportunity
type YieldOpportunity struct {
	PoolAddress    string  `json:"poolAddress"`
	PoolName       string  `json:"poolName"`
	APY            float64 `json:"apy"`
	TVL            float64 `json:"tvl"`
	RiskScore      float64 `json:"riskScore"`
	TokenPair      string  `json:"tokenPair"`
	LastUpdated    int64   `json:"lastUpdated"`
	Recommendation string  `json:"recommendation"`
}

// GovernanceSentiment represents governance proposal sentiment
type GovernanceSentiment struct {
	ProposalID    uint64  `json:"proposalId"`
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	Sentiment     float64 `json:"sentiment"` // -1 to 1
	SupportVotes  uint64  `json:"supportVotes"`
	AgainstVotes  uint64  `json:"againstVotes"`
	TotalVotes    uint64  `json:"totalVotes"`
	Participation float64 `json:"participation"`
	Status        string  `json:"status"`
}

// TradingSuggestion represents a trading recommendation
type TradingSuggestion struct {
	TokenPair      string  `json:"tokenPair"`
	Action         string  `json:"action"` // "buy", "sell", "hold"
	Confidence     float64 `json:"confidence"`
	Price          float64 `json:"price"`
	TargetPrice    float64 `json:"targetPrice"`
	StopLoss       float64 `json:"stopLoss"`
	Reasoning      string  `json:"reasoning"`
	RiskLevel      string  `json:"riskLevel"`
	TimeHorizon    string  `json:"timeHorizon"`
	LastUpdated    int64   `json:"lastUpdated"`
}

// TransactionVolume represents transaction volume data
type TransactionVolume struct {
	Timestamp int64   `json:"timestamp"`
	Volume    float64 `json:"volume"`
	Count     uint64  `json:"count"`
	Average   float64 `json:"average"`
	Trend     string  `json:"trend"`
}

// GasTrend represents gas price trends
type GasTrend struct {
	Timestamp int64   `json:"timestamp"`
	GasPrice  float64 `json:"gasPrice"`
	Trend     string  `json:"trend"`
	Prediction float64 `json:"prediction"`
}

// NewEngine creates a new analytics engine
func NewEngine(cfg *config.Config, bc *contracts.BlockchainClient) *Engine {
	// Create worker pool for concurrent analytics processing
	workerPool, err := ants.NewPool(cfg.AnalyticsWorkerPoolSize)
	if err != nil {
		logrus.Fatalf("Failed to create worker pool: %v", err)
	}

	engine := &Engine{
		config:           cfg,
		blockchainClient: bc,
		workerPool:       workerPool,
		stopChan:         make(chan struct{}),
	}

	return engine
}

// Start starts the analytics engine
func (e *Engine) Start() {
	logrus.Info("Starting analytics engine")
	
	// Start periodic analytics updates
	go e.runAnalyticsUpdates()
}

// Stop stops the analytics engine
func (e *Engine) Stop() {
	logrus.Info("Stopping analytics engine")
	close(e.stopChan)
	e.workerPool.Release()
}

// runAnalyticsUpdates runs periodic analytics updates
func (e *Engine) runAnalyticsUpdates() {
	ticker := time.NewTicker(e.config.AnalyticsUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			e.updateAnalytics()
		case <-e.stopChan:
			return
		}
	}
}

// updateAnalytics updates all analytics data
func (e *Engine) updateAnalytics() {
	// Update yield opportunities
	e.workerPool.Submit(func() {
		e.updateYieldOpportunities()
	})

	// Update governance sentiment
	e.workerPool.Submit(func() {
		e.updateGovernanceSentiment()
	})

	// Update trading suggestions
	e.workerPool.Submit(func() {
		e.updateTradingSuggestions()
	})

	// Update transaction volume
	e.workerPool.Submit(func() {
		e.updateTransactionVolume()
	})

	// Update gas trends
	e.workerPool.Submit(func() {
		e.updateGasTrends()
	})
}

// updateYieldOpportunities analyzes yield farming opportunities
func (e *Engine) updateYieldOpportunities() {
	// Mock yield data - in real implementation, fetch from DeFi protocols
	yieldData := []YieldOpportunity{
		{
			PoolAddress:    "0x1234567890123456789012345678901234567890",
			PoolName:       "KAIA-USDC LP",
			APY:            12.5,
			TVL:            1000000.0,
			RiskScore:      0.3,
			TokenPair:      "KAIA/USDC",
			LastUpdated:    time.Now().Unix(),
			Recommendation: "High yield with moderate risk",
		},
		{
			PoolAddress:    "0x2345678901234567890123456789012345678901",
			PoolName:       "KAIA-ETH LP",
			APY:            8.2,
			TVL:            2500000.0,
			RiskScore:      0.2,
			TokenPair:      "KAIA/ETH",
			LastUpdated:    time.Now().Unix(),
			Recommendation: "Stable yield with low risk",
		},
	}

	e.mu.Lock()
	e.yieldData = yieldData
	e.mu.Unlock()

	logrus.Debug("Updated yield opportunities")
}

// updateGovernanceSentiment analyzes governance proposal sentiment
func (e *Engine) updateGovernanceSentiment() {
	// Mock governance data - in real implementation, fetch from governance contracts
	governanceData := []GovernanceSentiment{
		{
			ProposalID:    1,
			Title:         "Increase Protocol Fee",
			Description:   "Proposal to increase protocol fee from 0.1% to 0.15%",
			Sentiment:     0.6,
			SupportVotes:  1500,
			AgainstVotes:  500,
			TotalVotes:    2000,
			Participation: 0.75,
			Status:        "Active",
		},
		{
			ProposalID:    2,
			Title:         "Add New Token Support",
			Description:   "Proposal to add support for new token pairs",
			Sentiment:     0.8,
			SupportVotes:  1800,
			AgainstVotes:  200,
			TotalVotes:    2000,
			Participation: 0.80,
			Status:        "Active",
		},
	}

	e.mu.Lock()
	e.governanceData = governanceData
	e.mu.Unlock()

	logrus.Debug("Updated governance sentiment")
}

// updateTradingSuggestions generates trading recommendations
func (e *Engine) updateTradingSuggestions() {
	// Mock trading data - in real implementation, use ML models
	tradingData := []TradingSuggestion{
		{
			TokenPair:   "KAIA/USDC",
			Action:      "buy",
			Confidence:  0.75,
			Price:       1.25,
			TargetPrice: 1.40,
			StopLoss:    1.15,
			Reasoning:   "Strong technical indicators and positive sentiment",
			RiskLevel:   "medium",
			TimeHorizon: "1 week",
			LastUpdated: time.Now().Unix(),
		},
		{
			TokenPair:   "KAIA/ETH",
			Action:      "hold",
			Confidence:  0.60,
			Price:       0.0008,
			TargetPrice: 0.0009,
			StopLoss:    0.0007,
			Reasoning:   "Sideways movement expected",
			RiskLevel:   "low",
			TimeHorizon: "3 days",
			LastUpdated: time.Now().Unix(),
		},
	}

	e.mu.Lock()
	e.tradingData = tradingData
	e.mu.Unlock()

	logrus.Debug("Updated trading suggestions")
}

// updateTransactionVolume analyzes transaction volume trends
func (e *Engine) updateTransactionVolume() {
	// Generate mock volume data with trend analysis
	volumes := []float64{1000, 1200, 1100, 1400, 1300, 1600, 1500, 1800}
	
	// Calculate trend
	trend := "increasing"
	if len(volumes) >= 2 {
		if volumes[len(volumes)-1] < volumes[len(volumes)-2] {
			trend = "decreasing"
		}
	}

	volumeData := []TransactionVolume{
		{
			Timestamp: time.Now().Add(-7 * 24 * time.Hour).Unix(),
			Volume:    volumes[len(volumes)-1],
			Count:     1500,
			Average:   volumes[len(volumes)-1] / 1500,
			Trend:     trend,
		},
	}

	e.mu.Lock()
	e.volumeData = volumeData
	e.mu.Unlock()

	logrus.Debug("Updated transaction volume")
}

// updateGasTrends analyzes gas price trends
func (e *Engine) updateGasTrends() {
	// Mock gas data with prediction
	gasPrices := []float64{20, 25, 22, 28, 30, 27, 32, 35}
	
	// Simple linear regression for prediction
	prediction := e.predictGasPrice(gasPrices)
	
	trend := "stable"
	if len(gasPrices) >= 2 {
		if gasPrices[len(gasPrices)-1] > gasPrices[len(gasPrices)-2] {
			trend = "increasing"
		} else if gasPrices[len(gasPrices)-1] < gasPrices[len(gasPrices)-2] {
			trend = "decreasing"
		}
	}

	gasData := []GasTrend{
		{
			Timestamp:  time.Now().Unix(),
			GasPrice:   gasPrices[len(gasPrices)-1],
			Trend:      trend,
			Prediction: prediction,
		},
	}

	e.mu.Lock()
	e.gasData = gasData
	e.mu.Unlock()

	logrus.Debug("Updated gas trends")
}

// predictGasPrice uses simple linear regression to predict gas price
func (e *Engine) predictGasPrice(prices []float64) float64 {
	if len(prices) < 2 {
		return prices[len(prices)-1]
	}

	// Simple linear regression
	x := make([]float64, len(prices))
	for i := range x {
		x[i] = float64(i)
	}

	slope, intercept := stat.LinearRegression(x, prices, nil)
	
	// Predict next value
	nextX := float64(len(prices))
	prediction := slope*nextX + intercept
	
	return math.Max(prediction, 0) // Gas price can't be negative
}

// HTTP Handlers

// GetYieldOpportunities returns yield farming opportunities
func (e *Engine) GetYieldOpportunities(c *gin.Context) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	c.JSON(200, gin.H{
		"opportunities": e.yieldData,
		"total":         len(e.yieldData),
		"timestamp":     time.Now().Unix(),
	})
}

// GetGovernanceSentiment returns governance sentiment analysis
func (e *Engine) GetGovernanceSentiment(c *gin.Context) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	c.JSON(200, gin.H{
		"proposals": e.governanceData,
		"total":     len(e.governanceData),
		"timestamp": time.Now().Unix(),
	})
}

// GetTradingSuggestions returns trading recommendations
func (e *Engine) GetTradingSuggestions(c *gin.Context) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	c.JSON(200, gin.H{
		"suggestions": e.tradingData,
		"total":       len(e.tradingData),
		"timestamp":   time.Now().Unix(),
	})
}

// GetTransactionVolume returns transaction volume data
func (e *Engine) GetTransactionVolume(c *gin.Context) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	c.JSON(200, gin.H{
		"volumes":   e.volumeData,
		"total":     len(e.volumeData),
		"timestamp": time.Now().Unix(),
	})
}

// GetGasTrends returns gas price trends
func (e *Engine) GetGasTrends(c *gin.Context) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	c.JSON(200, gin.H{
		"trends":    e.gasData,
		"total":     len(e.gasData),
		"timestamp": time.Now().Unix(),
	})
}

// Statistical Analysis Functions

// CalculateVolatility calculates price volatility
func (e *Engine) CalculateVolatility(prices []float64) float64 {
	if len(prices) < 2 {
		return 0
	}

	// Calculate returns
	returns := make([]float64, len(prices)-1)
	for i := 0; i < len(prices)-1; i++ {
		returns[i] = (prices[i+1] - prices[i]) / prices[i]
	}

	// Calculate standard deviation
	mean := stat.Mean(returns, nil)
	variance := stat.Variance(returns, nil)
	
	return math.Sqrt(variance)
}

// CalculateSharpeRatio calculates Sharpe ratio for risk-adjusted returns
func (e *Engine) CalculateSharpeRatio(returns []float64, riskFreeRate float64) float64 {
	if len(returns) < 2 {
		return 0
	}

	meanReturn := stat.Mean(returns, nil)
	stdDev := math.Sqrt(stat.Variance(returns, nil))
	
	if stdDev == 0 {
		return 0
	}
	
	return (meanReturn - riskFreeRate) / stdDev
}

// CalculateCorrelation calculates correlation between two datasets
func (e *Engine) CalculateCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0
	}

	return stat.Correlation(x, y, nil)
}

// GenerateMonteCarloSimulation generates Monte Carlo simulation for price prediction
func (e *Engine) GenerateMonteCarloSimulation(initialPrice, volatility, drift float64, steps, simulations int) []float64 {
	results := make([]float64, simulations)
	
	for i := 0; i < simulations; i++ {
		price := initialPrice
		for j := 0; j < steps; j++ {
			// Generate random normal distribution
			normal := distuv.Normal{Mu: 0, Sigma: 1}
			random := normal.Rand()
			
			// Update price using geometric Brownian motion
			price = price * math.Exp((drift-0.5*volatility*volatility)*1.0/365 + volatility*math.Sqrt(1.0/365)*random)
		}
		results[i] = price
	}
	
	return results
}