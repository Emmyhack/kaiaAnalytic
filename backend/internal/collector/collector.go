package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"kaia-analytics-ai/internal/config"
	"kaia-analytics-ai/internal/contracts"
)

// Collector handles data collection from various sources
type Collector struct {
	config           *config.Config
	blockchainClient *contracts.BlockchainClient
	stopChan         chan struct{}
	mu               sync.RWMutex
	
	// Data caches
	blockchainData map[string]interface{}
	marketData     map[string]interface{}
	historicalData map[string]interface{}
	
	// HTTP client
	httpClient *http.Client
}

// MarketData represents market data from external APIs
type MarketData struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Change24h float64 `json:"change_24h"`
	Volume24h float64 `json:"volume_24h"`
	MarketCap float64 `json:"market_cap"`
	Timestamp int64   `json:"timestamp"`
}

// HistoricalData represents historical blockchain data
type HistoricalData struct {
	BlockNumber uint64    `json:"blockNumber"`
	Timestamp   int64     `json:"timestamp"`
	GasUsed     uint64    `json:"gasUsed"`
	GasPrice    float64   `json:"gasPrice"`
	TxCount     uint64    `json:"txCount"`
	Volume      float64   `json:"volume"`
}

// NewCollector creates a new data collector
func NewCollector(cfg *config.Config, bc *contracts.BlockchainClient) *Collector {
	collector := &Collector{
		config:           cfg,
		blockchainClient: bc,
		stopChan:         make(chan struct{}),
		blockchainData:   make(map[string]interface{}),
		marketData:       make(map[string]interface{}),
		historicalData:   make(map[string]interface{}),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	return collector
}

// Start starts the data collector
func (e *Collector) Start() {
	logrus.Info("Starting data collector")
	
	// Start periodic data collection
	go e.runDataCollection()
}

// Stop stops the data collector
func (e *Collector) Stop() {
	logrus.Info("Stopping data collector")
	close(e.stopChan)
}

// runDataCollection runs periodic data collection
func (e *Collector) runDataCollection() {
	ticker := time.NewTicker(e.config.DataCollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			e.collectAllData()
		case <-e.stopChan:
			return
		}
	}
}

// collectAllData collects data from all sources
func (e *Collector) collectAllData() {
	// Collect blockchain data
	go e.collectBlockchainData()
	
	// Collect market data
	go e.collectMarketData()
	
	// Collect historical data
	go e.collectHistoricalData()
}

// collectBlockchainData collects real-time blockchain data
func (e *Collector) collectBlockchainData() {
	data, err := e.blockchainClient.GetBlockchainData()
	if err != nil {
		logrus.Errorf("Failed to collect blockchain data: %v", err)
		return
	}

	e.mu.Lock()
	e.blockchainData = data
	e.mu.Unlock()

	logrus.Debug("Updated blockchain data")
}

// collectMarketData collects market data from external APIs
func (e *Collector) collectMarketData() {
	// Collect from CoinGecko API
	marketData, err := e.fetchCoinGeckoData()
	if err != nil {
		logrus.Errorf("Failed to fetch CoinGecko data: %v", err)
		return
	}

	e.mu.Lock()
	e.marketData = marketData
	e.mu.Unlock()

	logrus.Debug("Updated market data")
}

// collectHistoricalData collects historical blockchain data
func (e *Collector) collectHistoricalData() {
	// Collect from Kaiascan API
	historicalData, err := e.fetchKaiascanData()
	if err != nil {
		logrus.Errorf("Failed to fetch Kaiascan data: %v", err)
		return
	}

	e.mu.Lock()
	e.historicalData = historicalData
	e.mu.Unlock()

	logrus.Debug("Updated historical data")
}

// fetchCoinGeckoData fetches market data from CoinGecko API
func (e *Collector) fetchCoinGeckoData() (map[string]interface{}, error) {
	// Mock CoinGecko data - in real implementation, make actual API calls
	marketData := map[string]interface{}{
		"kaia": map[string]interface{}{
			"symbol":      "kaia",
			"price":       1.25,
			"change_24h":  5.2,
			"volume_24h":  1000000.0,
			"market_cap":  50000000.0,
			"timestamp":   time.Now().Unix(),
		},
		"ethereum": map[string]interface{}{
			"symbol":      "eth",
			"price":       2000.0,
			"change_24h":  2.1,
			"volume_24h":  5000000.0,
			"market_cap":  240000000000.0,
			"timestamp":   time.Now().Unix(),
		},
	}

	return marketData, nil
}

// fetchKaiascanData fetches historical data from Kaiascan API
func (e *Collector) fetchKaiascanData() (map[string]interface{}, error) {
	// Mock Kaiascan data - in real implementation, make actual API calls
	historicalData := map[string]interface{}{
		"blocks": []map[string]interface{}{
			{
				"blockNumber": 1000000,
				"timestamp":   time.Now().Add(-1 * time.Hour).Unix(),
				"gasUsed":     15000000,
				"gasPrice":    25.0,
				"txCount":     150,
				"volume":      500000.0,
			},
			{
				"blockNumber": 999999,
				"timestamp":   time.Now().Add(-2 * time.Hour).Unix(),
				"gasUsed":     14800000,
				"gasPrice":    24.0,
				"txCount":     145,
				"volume":      480000.0,
			},
		},
		"transactions": []map[string]interface{}{
			{
				"hash":        "0x1234567890123456789012345678901234567890",
				"blockNumber": 1000000,
				"from":        "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
				"to":          "0xfedcbafedcbafedcbafedcbafedcbafedcbafedc",
				"value":       "1000000000000000000",
				"gasUsed":     21000,
				"gasPrice":    "25000000000",
				"timestamp":   time.Now().Add(-1 * time.Hour).Unix(),
			},
		},
	}

	return historicalData, nil
}

// HTTP Handlers

// GetBlockchainData returns current blockchain data
func (e *Collector) GetBlockchainData(c *gin.Context) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	c.JSON(200, gin.H{
		"data":      e.blockchainData,
		"timestamp": time.Now().Unix(),
	})
}

// GetMarketData returns current market data
func (e *Collector) GetMarketData(c *gin.Context) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	c.JSON(200, gin.H{
		"data":      e.marketData,
		"timestamp": time.Now().Unix(),
	})
}

// GetHistoricalData returns historical blockchain data
func (e *Collector) GetHistoricalData(c *gin.Context) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	c.JSON(200, gin.H{
		"data":      e.historicalData,
		"timestamp": time.Now().Unix(),
	})
}

// Utility Functions

// makeHTTPRequest makes an HTTP request with retry logic
func (e *Collector) makeHTTPRequest(url string) ([]byte, error) {
	var lastErr error
	
	for i := 0; i < e.config.MaxRetries; i++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		// Add headers
		req.Header.Set("User-Agent", "KaiaAnalyticsAI/1.0")
		if e.config.KaiascanAPIKey != "" {
			req.Header.Set("Authorization", "Bearer "+e.config.KaiascanAPIKey)
		}

		resp, err := e.httpClient.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		return body, nil
	}

	return nil, fmt.Errorf("failed after %d retries: %v", e.config.MaxRetries, lastErr)
}

// parseJSONResponse parses JSON response
func (e *Collector) parseJSONResponse(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// calculateMovingAverage calculates moving average for a slice of values
func (e *Collector) calculateMovingAverage(values []float64, window int) []float64 {
	if len(values) < window {
		return values
	}

	result := make([]float64, len(values)-window+1)
	for i := 0; i <= len(values)-window; i++ {
		sum := 0.0
		for j := i; j < i+window; j++ {
			sum += values[j]
		}
		result[i] = sum / float64(window)
	}

	return result
}

// calculateExponentialMovingAverage calculates exponential moving average
func (e *Collector) calculateExponentialMovingAverage(values []float64, alpha float64) []float64 {
	if len(values) == 0 {
		return values
	}

	result := make([]float64, len(values))
	result[0] = values[0]

	for i := 1; i < len(values); i++ {
		result[i] = alpha*values[i] + (1-alpha)*result[i-1]
	}

	return result
}

// detectAnomalies detects anomalies in time series data
func (e *Collector) detectAnomalies(values []float64, threshold float64) []bool {
	if len(values) < 2 {
		return make([]bool, len(values))
	}

	anomalies := make([]bool, len(values))
	mean := 0.0
	variance := 0.0

	// Calculate mean
	for _, v := range values {
		mean += v
	}
	mean /= float64(len(values))

	// Calculate variance
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	variance /= float64(len(values))
	stdDev := sqrt(variance)

	// Detect anomalies
	for i, v := range values {
		zScore := abs(v-mean) / stdDev
		anomalies[i] = zScore > threshold
	}

	return anomalies
}

// Helper functions
func sqrt(x float64) float64 {
	return float64(int(x*100)) / 100
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}