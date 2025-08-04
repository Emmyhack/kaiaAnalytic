// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";

/**
 * @title DataContract
 * @dev Contract for storing analytics results and anonymized user trade history
 */
contract DataContract is Ownable, ReentrancyGuard {
    struct AnalyticsResult {
        uint256 resultId;
        uint256 taskId;
        address provider;
        string dataType; // "yield_analysis", "governance_sentiment", "trade_optimization"
        string resultHash; // IPFS hash of detailed results
        bytes32 dataChecksum; // Integrity verification
        uint256 timestamp;
        bool isVerified;
        uint256 accessCount;
    }

    struct TradeHistory {
        uint256 recordId;
        bytes32 userHash; // Anonymized user identifier
        string tokenPair;
        uint256 tradeType; // 0: buy, 1: sell, 2: swap
        uint256 amount;
        uint256 price;
        uint256 timestamp;
        uint256 gasUsed;
        bool isSuccessful;
        bytes32 metadataHash; // Additional anonymized metadata
    }

    struct YieldOpportunity {
        uint256 opportunityId;
        string protocol;
        string poolAddress;
        uint256 apr; // Annual Percentage Rate (basis points)
        uint256 tvl; // Total Value Locked
        uint256 riskScore; // 1-100, higher is riskier
        string strategy;
        uint256 minDeposit;
        uint256 timestamp;
        bool isActive;
    }

    struct GovernanceData {
        uint256 proposalId;
        string proposalHash;
        uint256 sentimentScore; // 1-100, sentiment analysis result
        uint256 participationRate;
        uint256 votingPower;
        string analysisHash; // IPFS hash of detailed analysis
        uint256 timestamp;
    }

    // State variables
    uint256 private _resultCounter;
    uint256 private _tradeCounter;
    uint256 private _yieldCounter;
    uint256 private _governanceCounter;

    mapping(uint256 => AnalyticsResult) public analyticsResults;
    mapping(uint256 => TradeHistory) public tradeHistory;
    mapping(uint256 => YieldOpportunity) public yieldOpportunities;
    mapping(uint256 => GovernanceData) public governanceData;

    // Access control
    mapping(address => bool) public authorizedAnalytics;
    mapping(address => bool) public dataProviders;
    
    // Data categorization
    mapping(string => uint256[]) public resultsByType;
    mapping(bytes32 => uint256[]) public tradesByUser;
    mapping(string => uint256[]) public yieldByProtocol;

    // Events
    event AnalyticsResultStored(
        uint256 indexed resultId,
        uint256 indexed taskId,
        address indexed provider,
        string dataType
    );
    
    event TradeHistoryRecorded(
        uint256 indexed recordId,
        bytes32 indexed userHash,
        string tokenPair,
        uint256 tradeType
    );
    
    event YieldOpportunityUpdated(
        uint256 indexed opportunityId,
        string indexed protocol,
        uint256 apr,
        bool isActive
    );
    
    event GovernanceAnalysisStored(
        uint256 indexed proposalId,
        uint256 sentimentScore,
        uint256 participationRate
    );

    event DataProviderAuthorized(address indexed provider, bool authorized);
    event AnalyticsAuthorized(address indexed analytics, bool authorized);

    // Custom errors
    error UnauthorizedAccess();
    error InvalidDataType();
    error DataNotFound();
    error ChecksumMismatch();
    error InactiveOpportunity();

    constructor(address initialOwner) Ownable(initialOwner) {}

    /**
     * @dev Store analytics result
     * @param taskId Associated task ID
     * @param dataType Type of analytics data
     * @param resultHash IPFS hash of the result
     * @param dataChecksum Checksum for integrity verification
     */
    function storeAnalyticsResult(
        uint256 taskId,
        string memory dataType,
        string memory resultHash,
        bytes32 dataChecksum
    ) external returns (uint256) {
        if (!authorizedAnalytics[msg.sender] && msg.sender != owner()) {
            revert UnauthorizedAccess();
        }

        _resultCounter++;
        uint256 resultId = _resultCounter;

        analyticsResults[resultId] = AnalyticsResult({
            resultId: resultId,
            taskId: taskId,
            provider: msg.sender,
            dataType: dataType,
            resultHash: resultHash,
            dataChecksum: dataChecksum,
            timestamp: block.timestamp,
            isVerified: false,
            accessCount: 0
        });

        resultsByType[dataType].push(resultId);

        emit AnalyticsResultStored(resultId, taskId, msg.sender, dataType);
        return resultId;
    }

    /**
     * @dev Record anonymized trade history
     * @param userHash Anonymized user identifier
     * @param tokenPair Token pair traded
     * @param tradeType Type of trade (0: buy, 1: sell, 2: swap)
     * @param amount Trade amount
     * @param price Trade price
     * @param gasUsed Gas consumed
     * @param isSuccessful Whether trade was successful
     * @param metadataHash Additional anonymized metadata
     */
    function recordTradeHistory(
        bytes32 userHash,
        string memory tokenPair,
        uint256 tradeType,
        uint256 amount,
        uint256 price,
        uint256 gasUsed,
        bool isSuccessful,
        bytes32 metadataHash
    ) external returns (uint256) {
        if (!dataProviders[msg.sender] && msg.sender != owner()) {
            revert UnauthorizedAccess();
        }

        require(tradeType <= 2, "Invalid trade type");

        _tradeCounter++;
        uint256 recordId = _tradeCounter;

        tradeHistory[recordId] = TradeHistory({
            recordId: recordId,
            userHash: userHash,
            tokenPair: tokenPair,
            tradeType: tradeType,
            amount: amount,
            price: price,
            timestamp: block.timestamp,
            gasUsed: gasUsed,
            isSuccessful: isSuccessful,
            metadataHash: metadataHash
        });

        tradesByUser[userHash].push(recordId);

        emit TradeHistoryRecorded(recordId, userHash, tokenPair, tradeType);
        return recordId;
    }

    /**
     * @dev Update yield opportunity data
     * @param protocol Protocol name
     * @param poolAddress Pool contract address
     * @param apr Annual Percentage Rate in basis points
     * @param tvl Total Value Locked
     * @param riskScore Risk score (1-100)
     * @param strategy Yield strategy description
     * @param minDeposit Minimum deposit required
     * @param isActive Whether opportunity is currently active
     */
    function updateYieldOpportunity(
        string memory protocol,
        string memory poolAddress,
        uint256 apr,
        uint256 tvl,
        uint256 riskScore,
        string memory strategy,
        uint256 minDeposit,
        bool isActive
    ) external returns (uint256) {
        if (!dataProviders[msg.sender] && msg.sender != owner()) {
            revert UnauthorizedAccess();
        }

        require(riskScore >= 1 && riskScore <= 100, "Risk score must be 1-100");

        _yieldCounter++;
        uint256 opportunityId = _yieldCounter;

        yieldOpportunities[opportunityId] = YieldOpportunity({
            opportunityId: opportunityId,
            protocol: protocol,
            poolAddress: poolAddress,
            apr: apr,
            tvl: tvl,
            riskScore: riskScore,
            strategy: strategy,
            minDeposit: minDeposit,
            timestamp: block.timestamp,
            isActive: isActive
        });

        yieldByProtocol[protocol].push(opportunityId);

        emit YieldOpportunityUpdated(opportunityId, protocol, apr, isActive);
        return opportunityId;
    }

    /**
     * @dev Store governance analysis data
     * @param proposalId Governance proposal ID
     * @param proposalHash Hash of the proposal
     * @param sentimentScore Sentiment analysis score (1-100)
     * @param participationRate Participation rate in basis points
     * @param votingPower Total voting power involved
     * @param analysisHash IPFS hash of detailed analysis
     */
    function storeGovernanceAnalysis(
        uint256 proposalId,
        string memory proposalHash,
        uint256 sentimentScore,
        uint256 participationRate,
        uint256 votingPower,
        string memory analysisHash
    ) external {
        if (!authorizedAnalytics[msg.sender] && msg.sender != owner()) {
            revert UnauthorizedAccess();
        }

        require(sentimentScore >= 1 && sentimentScore <= 100, "Sentiment score must be 1-100");

        governanceData[proposalId] = GovernanceData({
            proposalId: proposalId,
            proposalHash: proposalHash,
            sentimentScore: sentimentScore,
            participationRate: participationRate,
            votingPower: votingPower,
            analysisHash: analysisHash,
            timestamp: block.timestamp
        });

        emit GovernanceAnalysisStored(proposalId, sentimentScore, participationRate);
    }

    /**
     * @dev Verify analytics result integrity
     * @param resultId Result ID to verify
     * @param expectedChecksum Expected checksum
     */
    function verifyResult(uint256 resultId, bytes32 expectedChecksum) external {
        if (!authorizedAnalytics[msg.sender] && msg.sender != owner()) {
            revert UnauthorizedAccess();
        }

        if (analyticsResults[resultId].resultId == 0) {
            revert DataNotFound();
        }

        if (analyticsResults[resultId].dataChecksum != expectedChecksum) {
            revert ChecksumMismatch();
        }

        analyticsResults[resultId].isVerified = true;
    }

    /**
     * @dev Increment access count for analytics result
     * @param resultId Result ID
     */
    function recordAccess(uint256 resultId) external {
        if (analyticsResults[resultId].resultId == 0) {
            revert DataNotFound();
        }

        analyticsResults[resultId].accessCount++;
    }

    /**
     * @dev Get analytics results by type
     * @param dataType Data type to filter by
     */
    function getResultsByType(string memory dataType) 
        external 
        view 
        returns (uint256[] memory) 
    {
        return resultsByType[dataType];
    }

    /**
     * @dev Get trade history by anonymized user
     * @param userHash Anonymized user identifier
     */
    function getTradesByUser(bytes32 userHash) 
        external 
        view 
        returns (uint256[] memory) 
    {
        return tradesByUser[userHash];
    }

    /**
     * @dev Get yield opportunities by protocol
     * @param protocol Protocol name
     */
    function getYieldByProtocol(string memory protocol) 
        external 
        view 
        returns (uint256[] memory) 
    {
        return yieldByProtocol[protocol];
    }

    /**
     * @dev Get active yield opportunities with minimum APR
     * @param minApr Minimum APR in basis points
     */
    function getActiveYieldOpportunities(uint256 minApr) 
        external 
        view 
        returns (uint256[] memory) 
    {
        uint256[] memory activeOpportunities = new uint256[](_yieldCounter);
        uint256 count = 0;

        for (uint256 i = 1; i <= _yieldCounter; i++) {
            if (yieldOpportunities[i].isActive && yieldOpportunities[i].apr >= minApr) {
                activeOpportunities[count] = i;
                count++;
            }
        }

        // Resize array to actual count
        uint256[] memory result = new uint256[](count);
        for (uint256 i = 0; i < count; i++) {
            result[i] = activeOpportunities[i];
        }

        return result;
    }

    /**
     * @dev Authorize analytics contracts
     * @param analytics Analytics contract address
     * @param authorized Authorization status
     */
    function setAnalyticsAuthorization(
        address analytics,
        bool authorized
    ) external onlyOwner {
        authorizedAnalytics[analytics] = authorized;
        emit AnalyticsAuthorized(analytics, authorized);
    }

    /**
     * @dev Authorize data providers
     * @param provider Data provider address
     * @param authorized Authorization status
     */
    function setDataProviderAuthorization(
        address provider,
        bool authorized
    ) external onlyOwner {
        dataProviders[provider] = authorized;
        emit DataProviderAuthorized(provider, authorized);
    }

    /**
     * @dev Get current counters
     */
    function getCounters() external view returns (
        uint256 results,
        uint256 trades,
        uint256 yields,
        uint256 governance
    ) {
        return (_resultCounter, _tradeCounter, _yieldCounter, _governanceCounter);
    }

    /**
     * @dev Emergency data cleanup (only owner)
     * @param dataType Type of data to clean
     * @param ids Array of IDs to remove
     */
    function emergencyCleanup(
        string memory dataType,
        uint256[] memory ids
    ) external onlyOwner {
        // Implementation for emergency data cleanup if needed
        // This would be used in case of corrupted or malicious data
    }
}