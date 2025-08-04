package contracts

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"kaia-analytics-ai/internal/config"
)

// BlockchainClient handles all blockchain interactions
type BlockchainClient struct {
	client           *ethclient.Client
	config           *config.Config
	contracts        *ContractInstances
	subscriptionChan chan *types.Header
	stopChan         chan struct{}
}

// ContractInstances holds references to deployed contracts
type ContractInstances struct {
	AnalyticsRegistry  *AnalyticsRegistry
	DataContract       *DataContract
	SubscriptionContract *SubscriptionContract
	ActionContract     *ActionContract
}

// NewBlockchainClient creates a new blockchain client
func NewBlockchainClient(cfg *config.Config) (*BlockchainClient, error) {
	// Connect to Kaia blockchain
	client, err := ethclient.Dial(cfg.KaiaRPCURL)
	if err != nil {
		return nil, err
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Initialize contract instances
	contracts, err := initializeContracts(client, cfg)
	if err != nil {
		return nil, err
	}

	bc := &BlockchainClient{
		client:           client,
		config:           cfg,
		contracts:        contracts,
		subscriptionChan: make(chan *types.Header),
		stopChan:         make(chan struct{}),
	}

	// Start blockchain monitoring
	go bc.monitorBlockchain()

	return bc, nil
}

// initializeContracts creates contract instances
func initializeContracts(client *ethclient.Client, cfg *config.Config) (*ContractInstances, error) {
	contracts := &ContractInstances{}

	// AnalyticsRegistry
	if cfg.ContractAddresses.AnalyticsRegistry != "" {
		analyticsRegistry, err := NewAnalyticsRegistry(
			common.HexToAddress(cfg.ContractAddresses.AnalyticsRegistry),
			client,
		)
		if err != nil {
			return nil, err
		}
		contracts.AnalyticsRegistry = analyticsRegistry
	}

	// DataContract
	if cfg.ContractAddresses.DataContract != "" {
		dataContract, err := NewDataContract(
			common.HexToAddress(cfg.ContractAddresses.DataContract),
			client,
		)
		if err != nil {
			return nil, err
		}
		contracts.DataContract = dataContract
	}

	// SubscriptionContract
	if cfg.ContractAddresses.SubscriptionContract != "" {
		subscriptionContract, err := NewSubscriptionContract(
			common.HexToAddress(cfg.ContractAddresses.SubscriptionContract),
			client,
		)
		if err != nil {
			return nil, err
		}
		contracts.SubscriptionContract = subscriptionContract
	}

	// ActionContract
	if cfg.ContractAddresses.ActionContract != "" {
		actionContract, err := NewActionContract(
			common.HexToAddress(cfg.ContractAddresses.ActionContract),
			client,
		)
		if err != nil {
			return nil, err
		}
		contracts.ActionContract = actionContract
	}

	return contracts, nil
}

// monitorBlockchain monitors blockchain for new blocks and events
func (bc *BlockchainClient) monitorBlockchain() {
	headers := make(chan *types.Header)
	sub, err := bc.client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		logrus.Errorf("Failed to subscribe to new headers: %v", err)
		return
	}
	defer sub.Unsubscribe()

	for {
		select {
		case err := <-sub.Err():
			logrus.Errorf("Blockchain subscription error: %v", err)
			return
		case header := <-headers:
			bc.handleNewBlock(header)
		case <-bc.stopChan:
			return
		}
	}
}

// handleNewBlock processes new blockchain blocks
func (bc *BlockchainClient) handleNewBlock(header *types.Header) {
	logrus.Debugf("New block: %d", header.Number.Uint64())
	
	// Process block for analytics data
	go bc.processBlockForAnalytics(header)
}

// processBlockForAnalytics extracts analytics data from new blocks
func (bc *BlockchainClient) processBlockForAnalytics(header *types.Header) {
	block, err := bc.client.BlockByHash(context.Background(), header.Hash())
	if err != nil {
		logrus.Errorf("Failed to get block: %v", err)
		return
	}

	// Extract transaction data for analytics
	for _, tx := range block.Transactions() {
		bc.processTransaction(tx, block.Number())
	}
}

// processTransaction processes individual transactions for analytics
func (bc *BlockchainClient) processTransaction(tx *types.Transaction, blockNumber *big.Int) {
	// Extract transaction metadata
	txData := map[string]interface{}{
		"hash":      tx.Hash().Hex(),
		"from":      "", // Would need to recover from signature
		"to":        tx.To().Hex(),
		"value":     tx.Value().String(),
		"gas":       tx.Gas(),
		"gasPrice":  tx.GasPrice().String(),
		"blockNumber": blockNumber.String(),
		"timestamp": time.Now().Unix(),
	}

	logrus.Debugf("Processing transaction: %s", tx.Hash().Hex())
	
	// Store transaction data for analytics
	// This would typically involve storing to a database or cache
}

// GetSubscriptionPlans returns available subscription plans
func (bc *BlockchainClient) GetSubscriptionPlans(c *gin.Context) {
	if bc.contracts.SubscriptionContract == nil {
		c.JSON(500, gin.H{"error": "Subscription contract not available"})
		return
	}

	// Get total number of plans
	totalPlans, err := bc.contracts.SubscriptionContract.GetTotalPlans(nil)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get total plans"})
		return
	}

	var plans []map[string]interface{}
	for i := uint64(1); i <= totalPlans.Uint64(); i++ {
		plan, err := bc.contracts.SubscriptionContract.GetPlan(nil, big.NewInt(int64(i)))
		if err != nil {
			continue
		}

		plans = append(plans, map[string]interface{}{
			"planId":    plan.PlanId.String(),
			"name":      plan.Name,
			"price":     plan.Price.String(),
			"duration":  plan.Duration.String(),
			"isActive":  plan.IsActive,
			"features":  plan.Features,
		})
	}

	c.JSON(200, gin.H{
		"plans": plans,
		"total": len(plans),
	})
}

// PurchaseSubscription handles subscription purchase
func (bc *BlockchainClient) PurchaseSubscription(c *gin.Context) {
	var req struct {
		PlanID uint64 `json:"planId" binding:"required"`
		UserAddress string `json:"userAddress" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// Validate user address
	if !common.IsHexAddress(req.UserAddress) {
		c.JSON(400, gin.H{"error": "Invalid user address"})
		return
	}

	// Get plan details
	plan, err := bc.contracts.SubscriptionContract.GetPlan(nil, big.NewInt(int64(req.PlanID)))
	if err != nil {
		c.JSON(404, gin.H{"error": "Plan not found"})
		return
	}

	if !plan.IsActive {
		c.JSON(400, gin.H{"error": "Plan is not active"})
		return
	}

	// In a real implementation, you would:
	// 1. Create a transaction to purchase subscription
	// 2. Sign the transaction with user's private key
	// 3. Submit the transaction to the blockchain
	// 4. Wait for confirmation

	c.JSON(200, gin.H{
		"message": "Subscription purchase initiated",
		"planId":  req.PlanID,
		"price":   plan.Price.String(),
	})
}

// GetSubscriptionStatus returns user's subscription status
func (bc *BlockchainClient) GetSubscriptionStatus(c *gin.Context) {
	userAddress := c.Param("address")
	if !common.IsHexAddress(userAddress) {
		c.JSON(400, gin.H{"error": "Invalid address"})
		return
	}

	address := common.HexToAddress(userAddress)
	
	// Check if user has active subscription
	hasActive, err := bc.contracts.SubscriptionContract.HasActiveSubscription(nil, address)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to check subscription status"})
		return
	}

	if !hasActive {
		c.JSON(200, gin.H{
			"hasActiveSubscription": false,
			"subscription":         nil,
		})
		return
	}

	// Get subscription details
	subscription, err := bc.contracts.SubscriptionContract.GetUserActiveSubscription(nil, address)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get subscription details"})
		return
	}

	// Get plan details
	plan, err := bc.contracts.SubscriptionContract.GetPlan(nil, subscription.PlanId)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get plan details"})
		return
	}

	c.JSON(200, gin.H{
		"hasActiveSubscription": true,
		"subscription": map[string]interface{}{
			"subscriptionId": subscription.SubscriptionId.String(),
			"planId":         subscription.PlanId.String(),
			"startTime":      subscription.StartTime.String(),
			"endTime":        subscription.EndTime.String(),
			"isActive":       subscription.IsActive,
			"lastPayment":    subscription.LastPayment.String(),
			"plan": map[string]interface{}{
				"name":     plan.Name,
				"price":    plan.Price.String(),
				"duration": plan.Duration.String(),
				"features": plan.Features,
			},
		},
	})
}

// RegisterAnalyticsTask registers a new analytics task
func (bc *BlockchainClient) RegisterAnalyticsTask(taskType, description string, reward *big.Int) error {
	if bc.contracts.AnalyticsRegistry == nil {
		return fmt.Errorf("AnalyticsRegistry contract not available")
	}

	// In a real implementation, you would create and submit a transaction
	// For now, we'll just log the task registration
	logrus.Infof("Registering analytics task: %s - %s (reward: %s)", taskType, description, reward.String())
	return nil
}

// StoreAnalyticsResult stores analytics result on-chain
func (bc *BlockchainClient) StoreAnalyticsResult(taskID uint64, dataType string, dataHash [32]byte) error {
	if bc.contracts.DataContract == nil {
		return fmt.Errorf("DataContract not available")
	}

	// In a real implementation, you would create and submit a transaction
	logrus.Infof("Storing analytics result: taskID=%d, type=%s, hash=%x", taskID, dataType, dataHash)
	return nil
}

// CreateAction creates an on-chain action
func (bc *BlockchainClient) CreateAction(actionType string, actionData []byte) error {
	if bc.contracts.ActionContract == nil {
		return fmt.Errorf("ActionContract not available")
	}

	// In a real implementation, you would create and submit a transaction
	logrus.Infof("Creating action: type=%s, data=%x", actionType, actionData)
	return nil
}

// GetBlockchainData returns current blockchain state
func (bc *BlockchainClient) GetBlockchainData() (map[string]interface{}, error) {
	// Get latest block
	header, err := bc.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	// Get gas price
	gasPrice, err := bc.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	// Get network ID
	chainID, err := bc.client.NetworkID(context.Background())
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"latestBlockNumber": header.Number.String(),
		"latestBlockHash":   header.Hash().Hex(),
		"gasPrice":          gasPrice.String(),
		"chainId":           chainID.String(),
		"timestamp":         time.Now().Unix(),
	}, nil
}

// Close closes the blockchain client
func (bc *BlockchainClient) Close() {
	close(bc.stopChan)
	if bc.client != nil {
		bc.client.Close()
	}
}