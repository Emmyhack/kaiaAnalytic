package main

import (
	"context"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gin-gonic/gin"
)

// Response structures
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

type BlockResponse struct {
	Number       string   `json:"number"`
	Hash         string   `json:"hash"`
	ParentHash   string   `json:"parent_hash"`
	Timestamp    uint64   `json:"timestamp"`
	GasUsed      uint64   `json:"gas_used"`
	GasLimit     uint64   `json:"gas_limit"`
	Transactions int      `json:"transaction_count"`
	Size         string   `json:"size"`
}

type TransactionResponse struct {
	Hash             string `json:"hash"`
	BlockNumber      string `json:"block_number"`
	BlockHash        string `json:"block_hash"`
	TransactionIndex uint   `json:"transaction_index"`
	From             string `json:"from"`
	To               string `json:"to"`
	Value            string `json:"value"`
	Gas              uint64 `json:"gas"`
	GasPrice         string `json:"gas_price"`
	GasUsed          uint64 `json:"gas_used,omitempty"`
	Status           uint64 `json:"status,omitempty"`
}

type BalanceResponse struct {
	Address string `json:"address"`
	Balance string `json:"balance"`
	BalanceEth string `json:"balance_eth"`
}

type NetworkStatsResponse struct {
	LatestBlock    uint64 `json:"latest_block"`
	NetworkID      string `json:"network_id"`
	ChainID        string `json:"chain_id"`
	SyncProgress   bool   `json:"is_syncing"`
	PeerCount      uint64 `json:"peer_count"`
}

type ContractInfoResponse struct {
	Address     string `json:"address"`
	Code        string `json:"code"`
	CodeSize    int    `json:"code_size"`
	IsContract  bool   `json:"is_contract"`
}

// healthCheck returns the health status of the service
func (a *App) healthCheck(c *gin.Context) {
	// Test Ethereum connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	_, err := a.ethClient.NetworkID(ctx)
	status := "healthy"
	if err != nil {
		status = "unhealthy"
		a.logger.WithError(err).Error("Health check failed")
	}

	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Version:   "1.0.0",
	}

	if status == "healthy" {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusServiceUnavailable, response)
	}
}

// getBlockByNumber retrieves block information by block number
func (a *App) getBlockByNumber(c *gin.Context) {
	blockNumberStr := c.Param("number")
	
	var blockNumber *big.Int
	if blockNumberStr == "latest" {
		blockNumber = nil
	} else {
		num, err := strconv.ParseInt(blockNumberStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_block_number",
				Message: "Block number must be a valid integer or 'latest'",
			})
			return
		}
		blockNumber = big.NewInt(num)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	block, err := a.ethClient.BlockByNumber(ctx, blockNumber)
	if err != nil {
		a.logger.WithError(err).Error("Failed to get block")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "block_fetch_failed",
			Message: "Failed to retrieve block information",
		})
		return
	}

	response := BlockResponse{
		Number:       block.Number().String(),
		Hash:         block.Hash().Hex(),
		ParentHash:   block.ParentHash().Hex(),
		Timestamp:    block.Time(),
		GasUsed:      block.GasUsed(),
		GasLimit:     block.GasLimit(),
		Transactions: len(block.Transactions()),
		Size:         strconv.FormatUint(block.Size(), 10),
	}

	c.JSON(http.StatusOK, response)
}

// getTransactionByHash retrieves transaction information by hash
func (a *App) getTransactionByHash(c *gin.Context) {
	txHashStr := c.Param("hash")
	
	if !common.IsHexAddress(txHashStr) && len(txHashStr) != 66 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_transaction_hash",
			Message: "Transaction hash must be a valid hex string",
		})
		return
	}

	txHash := common.HexToHash(txHashStr)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, isPending, err := a.ethClient.TransactionByHash(ctx, txHash)
	if err != nil {
		a.logger.WithError(err).Error("Failed to get transaction")
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "transaction_not_found",
			Message: "Transaction not found",
		})
		return
	}

	// Get transaction receipt for additional info
	var gasUsed uint64
	var status uint64
	if !isPending {
		receipt, err := a.ethClient.TransactionReceipt(ctx, txHash)
		if err == nil {
			gasUsed = receipt.GasUsed
			status = receipt.Status
		}
	}

	response := TransactionResponse{
		Hash:             tx.Hash().Hex(),
		From:             getFromAddress(tx).Hex(),
		Gas:              tx.Gas(),
		GasPrice:         tx.GasPrice().String(),
		Value:            tx.Value().String(),
		GasUsed:          gasUsed,
		Status:           status,
	}

	if tx.To() != nil {
		response.To = tx.To().Hex()
	}

	c.JSON(http.StatusOK, response)
}

// getAddressBalance retrieves the balance of an Ethereum address
func (a *App) getAddressBalance(c *gin.Context) {
	addressStr := c.Param("address")
	
	if !common.IsHexAddress(addressStr) {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_address",
			Message: "Address must be a valid Ethereum address",
		})
		return
	}

	address := common.HexToAddress(addressStr)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	balance, err := a.ethClient.BalanceAt(ctx, address, nil)
	if err != nil {
		a.logger.WithError(err).Error("Failed to get balance")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "balance_fetch_failed",
			Message: "Failed to retrieve address balance",
		})
		return
	}

	// Convert Wei to Ether
	balanceEth := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(1e18))

	response := BalanceResponse{
		Address:    address.Hex(),
		Balance:    balance.String(),
		BalanceEth: balanceEth.String(),
	}

	c.JSON(http.StatusOK, response)
}

// getNetworkStats retrieves network statistics
func (a *App) getNetworkStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get latest block number
	latestBlock, err := a.ethClient.BlockNumber(ctx)
	if err != nil {
		a.logger.WithError(err).Error("Failed to get latest block")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "network_stats_failed",
			Message: "Failed to retrieve network statistics",
		})
		return
	}

	// Get network ID
	networkID, err := a.ethClient.NetworkID(ctx)
	if err != nil {
		a.logger.WithError(err).Error("Failed to get network ID")
		networkID = big.NewInt(0)
	}

	// Get chain ID
	chainID, err := a.ethClient.ChainID(ctx)
	if err != nil {
		a.logger.WithError(err).Error("Failed to get chain ID")
		chainID = big.NewInt(0)
	}

	// Check sync status
	syncProgress, err := a.ethClient.SyncProgress(ctx)
	isSyncing := err == nil && syncProgress != nil

	response := NetworkStatsResponse{
		LatestBlock:  latestBlock,
		NetworkID:    networkID.String(),
		ChainID:      chainID.String(),
		SyncProgress: isSyncing,
		PeerCount:    0, // Note: This requires admin API access
	}

	c.JSON(http.StatusOK, response)
}

// getContractInfo retrieves information about a smart contract
func (a *App) getContractInfo(c *gin.Context) {
	addressStr := c.Param("address")
	
	if !common.IsHexAddress(addressStr) {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_address",
			Message: "Address must be a valid Ethereum address",
		})
		return
	}

	address := common.HexToAddress(addressStr)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get contract code
	code, err := a.ethClient.CodeAt(ctx, address, nil)
	if err != nil {
		a.logger.WithError(err).Error("Failed to get contract code")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "contract_info_failed",
			Message: "Failed to retrieve contract information",
		})
		return
	}

	isContract := len(code) > 0
	codeHex := common.Bytes2Hex(code)

	response := ContractInfoResponse{
		Address:    address.Hex(),
		Code:       codeHex,
		CodeSize:   len(code),
		IsContract: isContract,
	}

	c.JSON(http.StatusOK, response)
}

// Helper function to extract from address from transaction
func getFromAddress(tx *types.Transaction) common.Address {
	from, err := types.Sender(types.NewEIP155Signer(tx.ChainId()), tx)
	if err != nil {
		return common.Address{}
	}
	return from
}

// KaiaAnalyticsAI specific handlers

// getKaiaNetworkStats handles requests for Kaia network statistics
func (a *App) getKaiaNetworkStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stats, err := a.dataCollector.GetKaiaNetworkStats(ctx)
	if err != nil {
		a.logger.WithError(err).Error("Failed to get Kaia network stats")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "kaia_stats_failed",
			Message: "Failed to retrieve Kaia network statistics",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// getYieldAnalytics handles requests for yield farming analytics
func (a *App) getYieldAnalytics(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	protocols := []string{"uniswap", "aave", "compound", "curve", "yearn"}
	if protocolParam := c.Query("protocol"); protocolParam != "" {
		protocols = []string{protocolParam}
	}

	yields, err := a.analyticsEngine.AnalyzeYieldOpportunities(ctx, protocols)
	if err != nil {
		a.logger.WithError(err).Error("Failed to analyze yield opportunities")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "yield_analysis_failed",
			Message: "Failed to analyze yield opportunities",
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"yields": yields,
		"count":  len(yields),
	})
}

// getTradeAnalysis handles requests for trading analysis
func (a *App) getTradeAnalysis(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pair := c.Param("pair")
	userAddress := c.Query("user")

	optimization, err := a.analyticsEngine.OptimizeTradingStrategy(ctx, userAddress, pair)
	if err != nil {
		a.logger.WithError(err).Error("Failed to optimize trading strategy")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "trade_analysis_failed",
			Message: "Failed to analyze trading strategy",
		})
		return
	}

	c.JSON(http.StatusOK, optimization)
}

// getGovernanceAnalysis handles requests for governance sentiment analysis
func (a *App) getGovernanceAnalysis(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	proposalID := c.Param("proposalId")

	sentiment, err := a.analyticsEngine.AnalyzeGovernanceSentiment(ctx, proposalID)
	if err != nil {
		a.logger.WithError(err).Error("Failed to analyze governance sentiment")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "governance_analysis_failed",
			Message: "Failed to analyze governance sentiment",
		})
		return
	}

	c.JSON(http.StatusOK, sentiment)
}

// getTokenPrice handles requests for token price information
func (a *App) getTokenPrice(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	symbol := c.Param("symbol")

	price, err := a.dataCollector.GetTokenPrice(ctx, symbol)
	if err != nil {
		a.logger.WithError(err).Error("Failed to get token price")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "price_fetch_failed",
			Message: "Failed to retrieve token price",
		})
		return
	}

	c.JSON(http.StatusOK, price)
}

// getHistoricalPrices handles requests for historical price data
func (a *App) getHistoricalPrices(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pair := c.Param("pair")
	daysStr := c.DefaultQuery("days", "30")
	
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 || days > 365 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_days",
			Message: "Days parameter must be a positive integer between 1 and 365",
		})
		return
	}

	historicalData, err := a.dataCollector.GetHistoricalPrices(ctx, pair, days)
	if err != nil {
		a.logger.WithError(err).Error("Failed to get historical prices")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "historical_data_failed",
			Message: "Failed to retrieve historical price data",
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"pair":  pair,
		"days":  days,
		"data":  historicalData,
		"count": len(historicalData),
	})
}

// processChatQuery handles chat query requests
func (a *App) processChatQuery(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var request struct {
		UserID    string `json:"user_id" binding:"required"`
		Query     string `json:"query" binding:"required"`
		SessionID string `json:"session_id"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request format",
		})
		return
	}

	response, err := a.chatEngine.ProcessQuery(ctx, request.UserID, request.Query, request.SessionID)
	if err != nil {
		a.logger.WithError(err).Error("Failed to process chat query")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "chat_processing_failed",
			Message: "Failed to process chat query",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// getChatHistory handles requests for chat history
func (a *App) getChatHistory(c *gin.Context) {
	userID := c.Param("userId")
	limitStr := c.DefaultQuery("limit", "20")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_limit",
			Message: "Limit parameter must be a positive integer between 1 and 100",
		})
		return
	}

	history := a.chatEngine.GetChatHistory(userID, limit)

	c.JSON(http.StatusOK, map[string]interface{}{
		"user_id": userID,
		"history": history,
		"count":   len(history),
	})
}