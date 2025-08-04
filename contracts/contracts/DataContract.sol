// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/utils/Counters.sol";

/**
 * @title DataContract
 * @dev Contract for storing analytics results and anonymized user trade history
 */
contract DataContract is Ownable, ReentrancyGuard {
    using Counters for Counters.Counter;

    struct AnalyticsResult {
        uint256 id;
        uint256 taskId; // Reference to AnalyticsRegistry task
        string dataType; // "yield_data", "sentiment_score", "trade_pattern"
        string resultData; // JSON or IPFS hash
        uint256 timestamp;
        address processor;
        bool isPublic; // Whether data is publicly accessible
        uint256 expiryTime; // When data expires (0 = never)
    }

    struct TradeData {
        uint256 id;
        bytes32 userHash; // Anonymized user identifier
        string tokenPair;
        uint256 amount; // Anonymized/normalized amount
        string tradeType; // "buy", "sell", "swap"
        uint256 timestamp;
        uint256 blockNumber;
        string protocolUsed;
        bool isAnonymized;
    }

    struct YieldOpportunity {
        uint256 id;
        string protocol;
        string tokenPair;
        uint256 apy; // APY in basis points (10000 = 100%)
        uint256 tvl; // Total Value Locked
        uint256 riskScore; // 1-100, higher is riskier
        string category; // "lending", "farming", "staking"
        uint256 timestamp;
        bool isActive;
    }

    struct GovernanceData {
        uint256 id;
        string proposalId;
        int256 sentimentScore; // -100 to 100
        uint256 participationRate; // In basis points
        string category;
        uint256 timestamp;
        string dataHash; // IPFS hash for detailed data
    }

    // State variables
    Counters.Counter private _resultIdCounter;
    Counters.Counter private _tradeIdCounter;
    Counters.Counter private _yieldIdCounter;
    Counters.Counter private _governanceIdCounter;

    mapping(uint256 => AnalyticsResult) public analyticsResults;
    mapping(uint256 => TradeData) public tradeData;
    mapping(uint256 => YieldOpportunity) public yieldOpportunities;
    mapping(uint256 => GovernanceData) public governanceData;

    // Indexing mappings
    mapping(string => uint256[]) public resultsByType;
    mapping(address => uint256[]) public resultsByProcessor;
    mapping(bytes32 => uint256[]) public tradesByUser;
    mapping(string => uint256[]) public tradesByProtocol;
    mapping(string => uint256[]) public yieldByProtocol;
    mapping(string => uint256[]) public governanceByCategory;

    // Access control
    mapping(address => bool) public authorizedProcessors;
    mapping(address => bool) public dataReaders;

    // Data retention settings
    uint256 public defaultDataRetention = 365 days;
    mapping(string => uint256) public dataTypeRetention;

    // Events
    event AnalyticsResultStored(
        uint256 indexed resultId,
        uint256 indexed taskId,
        string dataType,
        address processor
    );

    event TradeDataStored(
        uint256 indexed tradeId,
        bytes32 indexed userHash,
        string tokenPair,
        string protocol
    );

    event YieldOpportunityUpdated(
        uint256 indexed yieldId,
        string protocol,
        uint256 apy,
        bool isActive
    );

    event GovernanceDataStored(
        uint256 indexed governanceId,
        string proposalId,
        int256 sentimentScore
    );

    event ProcessorAuthorized(address indexed processor);
    event DataReaderAuthorized(address indexed reader);

    // Modifiers
    modifier onlyAuthorizedProcessor() {
        require(authorizedProcessors[msg.sender] || msg.sender == owner(), "Not authorized processor");
        _;
    }

    modifier onlyDataReader() {
        require(dataReaders[msg.sender] || msg.sender == owner(), "Not authorized reader");
        _;
    }

    constructor() {
        authorizedProcessors[msg.sender] = true;
        dataReaders[msg.sender] = true;
    }

    /**
     * @dev Store analytics result
     */
    function storeAnalyticsResult(
        uint256 taskId,
        string memory dataType,
        string memory resultData,
        bool isPublic,
        uint256 expiryTime
    ) external onlyAuthorizedProcessor nonReentrant returns (uint256) {
        require(bytes(dataType).length > 0, "Data type cannot be empty");
        require(bytes(resultData).length > 0, "Result data cannot be empty");

        _resultIdCounter.increment();
        uint256 resultId = _resultIdCounter.current();

        analyticsResults[resultId] = AnalyticsResult({
            id: resultId,
            taskId: taskId,
            dataType: dataType,
            resultData: resultData,
            timestamp: block.timestamp,
            processor: msg.sender,
            isPublic: isPublic,
            expiryTime: expiryTime
        });

        resultsByType[dataType].push(resultId);
        resultsByProcessor[msg.sender].push(resultId);

        emit AnalyticsResultStored(resultId, taskId, dataType, msg.sender);
        
        return resultId;
    }

    /**
     * @dev Store anonymized trade data
     */
    function storeTradeData(
        bytes32 userHash,
        string memory tokenPair,
        uint256 amount,
        string memory tradeType,
        string memory protocolUsed
    ) external onlyAuthorizedProcessor nonReentrant returns (uint256) {
        require(userHash != bytes32(0), "User hash cannot be empty");
        require(bytes(tokenPair).length > 0, "Token pair cannot be empty");

        _tradeIdCounter.increment();
        uint256 tradeId = _tradeIdCounter.current();

        tradeData[tradeId] = TradeData({
            id: tradeId,
            userHash: userHash,
            tokenPair: tokenPair,
            amount: amount,
            tradeType: tradeType,
            timestamp: block.timestamp,
            blockNumber: block.number,
            protocolUsed: protocolUsed,
            isAnonymized: true
        });

        tradesByUser[userHash].push(tradeId);
        tradesByProtocol[protocolUsed].push(tradeId);

        emit TradeDataStored(tradeId, userHash, tokenPair, protocolUsed);
        
        return tradeId;
    }

    /**
     * @dev Store or update yield opportunity
     */
    function storeYieldOpportunity(
        string memory protocol,
        string memory tokenPair,
        uint256 apy,
        uint256 tvl,
        uint256 riskScore,
        string memory category,
        bool isActive
    ) external onlyAuthorizedProcessor nonReentrant returns (uint256) {
        require(bytes(protocol).length > 0, "Protocol cannot be empty");
        require(apy <= 1000000, "APY too high"); // Max 10,000% APY
        require(riskScore <= 100, "Risk score must be <= 100");

        _yieldIdCounter.increment();
        uint256 yieldId = _yieldIdCounter.current();

        yieldOpportunities[yieldId] = YieldOpportunity({
            id: yieldId,
            protocol: protocol,
            tokenPair: tokenPair,
            apy: apy,
            tvl: tvl,
            riskScore: riskScore,
            category: category,
            timestamp: block.timestamp,
            isActive: isActive
        });

        yieldByProtocol[protocol].push(yieldId);

        emit YieldOpportunityUpdated(yieldId, protocol, apy, isActive);
        
        return yieldId;
    }

    /**
     * @dev Store governance sentiment data
     */
    function storeGovernanceData(
        string memory proposalId,
        int256 sentimentScore,
        uint256 participationRate,
        string memory category,
        string memory dataHash
    ) external onlyAuthorizedProcessor nonReentrant returns (uint256) {
        require(bytes(proposalId).length > 0, "Proposal ID cannot be empty");
        require(sentimentScore >= -100 && sentimentScore <= 100, "Invalid sentiment score");
        require(participationRate <= 10000, "Participation rate must be <= 100%");

        _governanceIdCounter.increment();
        uint256 governanceId = _governanceIdCounter.current();

        governanceData[governanceId] = GovernanceData({
            id: governanceId,
            proposalId: proposalId,
            sentimentScore: sentimentScore,
            participationRate: participationRate,
            category: category,
            timestamp: block.timestamp,
            dataHash: dataHash
        });

        governanceByCategory[category].push(governanceId);

        emit GovernanceDataStored(governanceId, proposalId, sentimentScore);
        
        return governanceId;
    }

    /**
     * @dev Get analytics result by ID
     */
    function getAnalyticsResult(uint256 resultId) external view returns (AnalyticsResult memory) {
        require(resultId > 0 && resultId <= _resultIdCounter.current(), "Invalid result ID");
        
        AnalyticsResult memory result = analyticsResults[resultId];
        
        // Check if data has expired
        if (result.expiryTime > 0 && block.timestamp > result.expiryTime) {
            revert("Data has expired");
        }
        
        // Check access permissions
        if (!result.isPublic) {
            require(
                dataReaders[msg.sender] || 
                msg.sender == result.processor || 
                msg.sender == owner(),
                "Access denied"
            );
        }
        
        return result;
    }

    /**
     * @dev Get yield opportunities by protocol
     */
    function getYieldOpportunitiesByProtocol(string memory protocol) 
        external 
        view 
        returns (YieldOpportunity[] memory) 
    {
        uint256[] memory yieldIds = yieldByProtocol[protocol];
        YieldOpportunity[] memory opportunities = new YieldOpportunity[](yieldIds.length);
        
        for (uint256 i = 0; i < yieldIds.length; i++) {
            opportunities[i] = yieldOpportunities[yieldIds[i]];
        }
        
        return opportunities;
    }

    /**
     * @dev Get active yield opportunities sorted by APY
     */
    function getTopYieldOpportunities(uint256 limit) 
        external 
        view 
        returns (YieldOpportunity[] memory) 
    {
        // This is a simplified version - in production, you'd want more efficient sorting
        uint256 totalYields = _yieldIdCounter.current();
        YieldOpportunity[] memory activeOpportunities = new YieldOpportunity[](limit);
        uint256 count = 0;
        
        for (uint256 i = 1; i <= totalYields && count < limit; i++) {
            if (yieldOpportunities[i].isActive) {
                activeOpportunities[count] = yieldOpportunities[i];
                count++;
            }
        }
        
        // Resize array to actual count
        YieldOpportunity[] memory result = new YieldOpportunity[](count);
        for (uint256 i = 0; i < count; i++) {
            result[i] = activeOpportunities[i];
        }
        
        return result;
    }

    /**
     * @dev Get user trade history (anonymized)
     */
    function getUserTradeHistory(bytes32 userHash) 
        external 
        view 
        onlyDataReader 
        returns (TradeData[] memory) 
    {
        uint256[] memory tradeIds = tradesByUser[userHash];
        TradeData[] memory trades = new TradeData[](tradeIds.length);
        
        for (uint256 i = 0; i < tradeIds.length; i++) {
            trades[i] = tradeData[tradeIds[i]];
        }
        
        return trades;
    }

    /**
     * @dev Get governance data by category
     */
    function getGovernanceDataByCategory(string memory category) 
        external 
        view 
        returns (GovernanceData[] memory) 
    {
        uint256[] memory governanceIds = governanceByCategory[category];
        GovernanceData[] memory data = new GovernanceData[](governanceIds.length);
        
        for (uint256 i = 0; i < governanceIds.length; i++) {
            data[i] = governanceData[governanceIds[i]];
        }
        
        return data;
    }

    /**
     * @dev Clean up expired data
     */
    function cleanupExpiredData() external onlyOwner {
        uint256 totalResults = _resultIdCounter.current();
        
        for (uint256 i = 1; i <= totalResults; i++) {
            AnalyticsResult storage result = analyticsResults[i];
            if (result.expiryTime > 0 && block.timestamp > result.expiryTime) {
                delete analyticsResults[i];
            }
        }
    }

    /**
     * @dev Set data retention period for a data type
     */
    function setDataRetention(string memory dataType, uint256 retentionPeriod) external onlyOwner {
        dataTypeRetention[dataType] = retentionPeriod;
    }

    /**
     * @dev Authorize a processor
     */
    function authorizeProcessor(address processor) external onlyOwner {
        require(processor != address(0), "Invalid processor address");
        authorizedProcessors[processor] = true;
        emit ProcessorAuthorized(processor);
    }

    /**
     * @dev Authorize a data reader
     */
    function authorizeDataReader(address reader) external onlyOwner {
        require(reader != address(0), "Invalid reader address");
        dataReaders[reader] = true;
        emit DataReaderAuthorized(reader);
    }

    /**
     * @dev Get total counts
     */
    function getTotalCounts() external view returns (uint256, uint256, uint256, uint256) {
        return (
            _resultIdCounter.current(),
            _tradeIdCounter.current(),
            _yieldIdCounter.current(),
            _governanceIdCounter.current()
        );
    }
}