// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";

/**
 * @title DataContract
 * @dev Stores analytics results and anonymized user trade history on-chain
 * @notice This contract provides decentralized storage for analytics data with privacy protection
 * @author Kaia Analytics AI Team
 */
contract DataContract is Ownable, ReentrancyGuard {
    
    /// @notice Structure for analytics result
    struct AnalyticsResult {
        uint256 resultId;
        uint256 taskId;
        string dataHash;
        string metadata;
        uint256 timestamp;
        address submitter;
        bool isValid;
    }
    
    /// @notice Structure for anonymized trade data
    struct TradeData {
        uint256 tradeId;
        bytes32 userHash; // Anonymized user identifier
        string tradeType;
        string assetPair;
        uint256 amount;
        uint256 price;
        uint256 timestamp;
        string protocol;
        bool isAnonymized;
    }
    
    /// @notice Mapping from result ID to analytics result
    mapping(uint256 => AnalyticsResult) public analyticsResults;
    
    /// @notice Mapping from trade ID to trade data
    mapping(uint256 => TradeData) public tradeData;
    
    /// @notice Mapping from task ID to result IDs
    mapping(uint256 => uint256[]) public taskResults;
    
    /// @notice Mapping from user hash to trade IDs
    mapping(bytes32 => uint256[]) public userTrades;
    
    /// @notice Total number of analytics results
    uint256 public totalResults;
    
    /// @notice Total number of trade records
    uint256 public totalTrades;
    
    /// @notice Fee for storing analytics results
    uint256 public analyticsStorageFee;
    
    /// @notice Fee for storing trade data
    uint256 public tradeStorageFee;
    
    /// @notice Emitted when analytics result is stored
    /// @param resultId The unique result identifier
    /// @param taskId The associated task ID
    /// @param dataHash The hash of the analytics data
    /// @param submitter The address that submitted the result
    /// @param timestamp The submission timestamp
    event AnalyticsResultStored(
        uint256 indexed resultId,
        uint256 indexed taskId,
        string dataHash,
        address indexed submitter,
        uint256 timestamp
    );
    
    /// @notice Emitted when trade data is stored
    /// @param tradeId The unique trade identifier
    /// @param userHash The anonymized user hash
    /// @param tradeType The type of trade
    /// @param assetPair The trading pair
    /// @param amount The trade amount
    /// @param price The trade price
    /// @param timestamp The trade timestamp
    event TradeDataStored(
        uint256 indexed tradeId,
        bytes32 indexed userHash,
        string tradeType,
        string assetPair,
        uint256 amount,
        uint256 price,
        uint256 timestamp
    );
    
    /// @notice Emitted when storage fees are updated
    /// @param analyticsFee The new analytics storage fee
    /// @param tradeFee The new trade storage fee
    event StorageFeesUpdated(uint256 analyticsFee, uint256 tradeFee);
    
    /// @dev Thrown when data hash is empty
    error EmptyDataHash();
    
    /// @dev Thrown when trade data is invalid
    error InvalidTradeData();
    
    /// @dev Thrown when result ID doesn't exist
    error ResultNotFound();
    
    /// @dev Thrown when trade ID doesn't exist
    error TradeNotFound();
    
    /// @dev Thrown when storage fee is insufficient
    error InsufficientStorageFee();
    
    /// @dev Thrown when caller is not authorized
    error NotAuthorized();

    /**
     * @notice Creates a new DataContract
     * @param _analyticsStorageFee The fee for storing analytics results
     * @param _tradeStorageFee The fee for storing trade data
     */
    constructor(uint256 _analyticsStorageFee, uint256 _tradeStorageFee) {
        analyticsStorageFee = _analyticsStorageFee;
        tradeStorageFee = _tradeStorageFee;
    }
    
    /**
     * @notice Stores analytics result data
     * @param _taskId The associated task ID
     * @param _dataHash The hash of the analytics data
     * @param _metadata Additional metadata about the result
     * @return resultId The unique identifier for the stored result
     */
    function storeAnalyticsResult(
        uint256 _taskId,
        string memory _dataHash,
        string memory _metadata
    ) external payable nonReentrant returns (uint256 resultId) {
        if (bytes(_dataHash).length == 0) {
            revert EmptyDataHash();
        }
        
        if (msg.value < analyticsStorageFee) {
            revert InsufficientStorageFee();
        }
        
        resultId = totalResults + 1;
        totalResults = resultId;
        
        analyticsResults[resultId] = AnalyticsResult({
            resultId: resultId,
            taskId: _taskId,
            dataHash: _dataHash,
            metadata: _metadata,
            timestamp: block.timestamp,
            submitter: msg.sender,
            isValid: true
        });
        
        taskResults[_taskId].push(resultId);
        
        emit AnalyticsResultStored(resultId, _taskId, _dataHash, msg.sender, block.timestamp);
    }
    
    /**
     * @notice Stores anonymized trade data
     * @param _userHash The anonymized user identifier
     * @param _tradeType The type of trade (buy, sell, swap, etc.)
     * @param _assetPair The trading pair (e.g., "ETH/USDC")
     * @param _amount The trade amount
     * @param _price The trade price
     * @param _protocol The protocol used for the trade
     * @return tradeId The unique identifier for the stored trade
     */
    function storeTradeData(
        bytes32 _userHash,
        string memory _tradeType,
        string memory _assetPair,
        uint256 _amount,
        uint256 _price,
        string memory _protocol
    ) external payable nonReentrant returns (uint256 tradeId) {
        if (bytes(_tradeType).length == 0 || bytes(_assetPair).length == 0) {
            revert InvalidTradeData();
        }
        
        if (_amount == 0 || _price == 0) {
            revert InvalidTradeData();
        }
        
        if (msg.value < tradeStorageFee) {
            revert InsufficientStorageFee();
        }
        
        tradeId = totalTrades + 1;
        totalTrades = tradeId;
        
        tradeData[tradeId] = TradeData({
            tradeId: tradeId,
            userHash: _userHash,
            tradeType: _tradeType,
            assetPair: _assetPair,
            amount: _amount,
            price: _price,
            timestamp: block.timestamp,
            protocol: _protocol,
            isAnonymized: true
        });
        
        userTrades[_userHash].push(tradeId);
        
        emit TradeDataStored(
            tradeId,
            _userHash,
            _tradeType,
            _assetPair,
            _amount,
            _price,
            block.timestamp
        );
    }
    
    /**
     * @notice Gets analytics result by ID
     * @param _resultId The result ID to query
     * @return result The complete analytics result structure
     */
    function getAnalyticsResult(uint256 _resultId) external view returns (AnalyticsResult memory result) {
        if (_resultId == 0 || _resultId > totalResults) {
            revert ResultNotFound();
        }
        return analyticsResults[_resultId];
    }
    
    /**
     * @notice Gets trade data by ID
     * @param _tradeId The trade ID to query
     * @return trade The complete trade data structure
     */
    function getTradeData(uint256 _tradeId) external view returns (TradeData memory trade) {
        if (_tradeId == 0 || _tradeId > totalTrades) {
            revert TradeNotFound();
        }
        return tradeData[_tradeId];
    }
    
    /**
     * @notice Gets all results for a specific task
     * @param _taskId The task ID to query
     * @return resultIds Array of result IDs for the task
     */
    function getTaskResults(uint256 _taskId) external view returns (uint256[] memory resultIds) {
        return taskResults[_taskId];
    }
    
    /**
     * @notice Gets all trades for a specific user hash
     * @param _userHash The anonymized user hash to query
     * @return tradeIds Array of trade IDs for the user
     */
    function getUserTrades(bytes32 _userHash) external view returns (uint256[] memory tradeIds) {
        return userTrades[_userHash];
    }
    
    /**
     * @notice Invalidates an analytics result
     * @param _resultId The result ID to invalidate
     * @dev Only the contract owner can invalidate results
     */
    function invalidateResult(uint256 _resultId) external onlyOwner {
        if (_resultId == 0 || _resultId > totalResults) {
            revert ResultNotFound();
        }
        
        analyticsResults[_resultId].isValid = false;
    }
    
    /**
     * @notice Updates storage fees
     * @param _analyticsFee The new analytics storage fee
     * @param _tradeFee The new trade storage fee
     * @dev Only the contract owner can update fees
     */
    function updateStorageFees(uint256 _analyticsFee, uint256 _tradeFee) external onlyOwner {
        analyticsStorageFee = _analyticsFee;
        tradeStorageFee = _tradeFee;
        emit StorageFeesUpdated(_analyticsFee, _tradeFee);
    }
    
    /**
     * @notice Withdraws accumulated storage fees
     * @dev Only the contract owner can withdraw fees
     */
    function withdrawFees() external onlyOwner {
        uint256 amount = address(this).balance;
        require(amount > 0, "No fees to withdraw");
        
        (bool success, ) = payable(owner()).call{value: amount}("");
        require(success, "Fee withdrawal failed");
    }
    
    /**
     * @notice Gets statistics about stored data
     * @return _totalResults Total number of analytics results
     * @return _totalTrades Total number of trade records
     * @return _validResults Number of valid analytics results
     */
    function getDataStatistics() external view returns (
        uint256 _totalResults,
        uint256 _totalTrades,
        uint256 _validResults
    ) {
        _totalResults = totalResults;
        _totalTrades = totalTrades;
        
        for (uint256 i = 1; i <= totalResults; i++) {
            if (analyticsResults[i].isValid) {
                _validResults++;
            }
        }
    }
}