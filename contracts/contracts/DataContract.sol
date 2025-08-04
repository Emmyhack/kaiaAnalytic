// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/utils/Counters.sol";

/**
 * @title DataContract
 * @dev Stores analytics results and anonymized user trade history on-chain
 */
contract DataContract is Ownable, ReentrancyGuard {
    using Counters for Counters.Counter;

    struct AnalyticsResult {
        uint256 resultId;
        uint256 taskId;
        string dataType; // "yield_analysis", "trade_history", "governance_data"
        bytes32 dataHash; // IPFS hash of the actual data
        uint256 timestamp;
        address submitter;
        bool isValidated;
        uint256 validationScore;
    }

    struct TradeHistory {
        uint256 tradeId;
        address user; // Anonymized address
        string tokenPair;
        uint256 amount;
        uint256 price;
        uint256 timestamp;
        string tradeType; // "buy", "sell", "swap"
        bool isAnonymized;
    }

    Counters.Counter private _resultIds;
    Counters.Counter private _tradeIds;
    
    mapping(uint256 => AnalyticsResult) public analyticsResults;
    mapping(uint256 => TradeHistory) public tradeHistory;
    mapping(address => uint256[]) public userTradeHistory;
    mapping(string => uint256[]) public resultsByType;
    mapping(bytes32 => bool) public dataHashes;
    
    event AnalyticsResultStored(uint256 indexed resultId, uint256 indexed taskId, string dataType, bytes32 dataHash);
    event TradeHistoryStored(uint256 indexed tradeId, address indexed user, string tokenPair, uint256 amount);
    event DataValidated(uint256 indexed resultId, uint256 validationScore);
    event DataAnonymized(uint256 indexed tradeId, address indexed originalUser, address indexed anonymizedUser);
    
    modifier resultExists(uint256 resultId) {
        require(analyticsResults[resultId].resultId != 0, "Result does not exist");
        _;
    }
    
    modifier onlyAuthorizedSubmitter() {
        require(msg.sender == owner() || isAuthorizedSubmitter(msg.sender), "Not authorized to submit data");
        _;
    }

    mapping(address => bool) private authorizedSubmitters;

    constructor() Ownable(msg.sender) {}

    /**
     * @dev Add authorized data submitter
     * @param submitter Address to authorize
     */
    function addAuthorizedSubmitter(address submitter) external onlyOwner {
        authorizedSubmitters[submitter] = true;
    }

    /**
     * @dev Remove authorized data submitter
     * @param submitter Address to remove authorization
     */
    function removeAuthorizedSubmitter(address submitter) external onlyOwner {
        authorizedSubmitters[submitter] = false;
    }

    /**
     * @dev Check if address is authorized submitter
     * @param submitter Address to check
     * @return bool True if authorized
     */
    function isAuthorizedSubmitter(address submitter) public view returns (bool) {
        return authorizedSubmitters[submitter];
    }

    /**
     * @dev Store analytics result
     * @param taskId Associated task ID
     * @param dataType Type of analytics data
     * @param dataHash IPFS hash of the data
     */
    function storeAnalyticsResult(
        uint256 taskId,
        string memory dataType,
        bytes32 dataHash
    ) external onlyAuthorizedSubmitter nonReentrant {
        require(bytes(dataType).length > 0, "Data type cannot be empty");
        require(dataHash != bytes32(0), "Data hash cannot be empty");
        require(!dataHashes[dataHash], "Data hash already exists");

        _resultIds.increment();
        uint256 resultId = _resultIds.current();

        AnalyticsResult memory newResult = AnalyticsResult({
            resultId: resultId,
            taskId: taskId,
            dataType: dataType,
            dataHash: dataHash,
            timestamp: block.timestamp,
            submitter: msg.sender,
            isValidated: false,
            validationScore: 0
        });

        analyticsResults[resultId] = newResult;
        resultsByType[dataType].push(resultId);
        dataHashes[dataHash] = true;

        emit AnalyticsResultStored(resultId, taskId, dataType, dataHash);
    }

    /**
     * @dev Store trade history entry
     * @param user User address (will be anonymized)
     * @param tokenPair Trading pair
     * @param amount Trade amount
     * @param price Trade price
     * @param tradeType Type of trade
     */
    function storeTradeHistory(
        address user,
        string memory tokenPair,
        uint256 amount,
        uint256 price,
        string memory tradeType
    ) external onlyAuthorizedSubmitter nonReentrant {
        require(user != address(0), "Invalid user address");
        require(bytes(tokenPair).length > 0, "Token pair cannot be empty");
        require(amount > 0, "Amount must be greater than 0");
        require(bytes(tradeType).length > 0, "Trade type cannot be empty");

        _tradeIds.increment();
        uint256 tradeId = _tradeIds.current();

        // Anonymize user address using simple hash
        address anonymizedUser = address(uint160(uint256(keccak256(abi.encodePacked(user, tradeId)))));

        TradeHistory memory newTrade = TradeHistory({
            tradeId: tradeId,
            user: anonymizedUser,
            tokenPair: tokenPair,
            amount: amount,
            price: price,
            timestamp: block.timestamp,
            tradeType: tradeType,
            isAnonymized: true
        });

        tradeHistory[tradeId] = newTrade;
        userTradeHistory[anonymizedUser].push(tradeId);

        emit TradeHistoryStored(tradeId, anonymizedUser, tokenPair, amount);
        emit DataAnonymized(tradeId, user, anonymizedUser);
    }

    /**
     * @dev Validate analytics result
     * @param resultId ID of the result to validate
     * @param validationScore Score from 0-100
     */
    function validateAnalyticsResult(uint256 resultId, uint256 validationScore) 
        external 
        onlyOwner 
        resultExists(resultId) 
    {
        require(validationScore <= 100, "Validation score must be between 0-100");
        
        AnalyticsResult storage result = analyticsResults[resultId];
        result.isValidated = true;
        result.validationScore = validationScore;

        emit DataValidated(resultId, validationScore);
    }

    /**
     * @dev Get analytics result
     * @param resultId ID of the result
     * @return result Analytics result details
     */
    function getAnalyticsResult(uint256 resultId) external view returns (AnalyticsResult memory result) {
        require(analyticsResults[resultId].resultId != 0, "Result does not exist");
        return analyticsResults[resultId];
    }

    /**
     * @dev Get trade history entry
     * @param tradeId ID of the trade
     * @return trade Trade history details
     */
    function getTradeHistory(uint256 tradeId) external view returns (TradeHistory memory trade) {
        require(tradeHistory[tradeId].tradeId != 0, "Trade does not exist");
        return tradeHistory[tradeId];
    }

    /**
     * @dev Get user's trade history
     * @param user Anonymized user address
     * @return tradeIds Array of trade IDs
     */
    function getUserTradeHistory(address user) external view returns (uint256[] memory tradeIds) {
        return userTradeHistory[user];
    }

    /**
     * @dev Get analytics results by type
     * @param dataType Type of analytics data
     * @return resultIds Array of result IDs
     */
    function getAnalyticsResultsByType(string memory dataType) external view returns (uint256[] memory resultIds) {
        return resultsByType[dataType];
    }

    /**
     * @dev Get total number of analytics results
     * @return Total result count
     */
    function getTotalAnalyticsResults() external view returns (uint256) {
        return _resultIds.current();
    }

    /**
     * @dev Get total number of trade history entries
     * @return Total trade count
     */
    function getTotalTradeHistory() external view returns (uint256) {
        return _tradeIds.current();
    }

    /**
     * @dev Get validated analytics results count
     * @return Validated result count
     */
    function getValidatedResultsCount() external view returns (uint256) {
        uint256 validatedCount = 0;
        for (uint256 i = 1; i <= _resultIds.current(); i++) {
            if (analyticsResults[i].isValidated) {
                validatedCount++;
            }
        }
        return validatedCount;
    }
}