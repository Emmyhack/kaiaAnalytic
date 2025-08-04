// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/utils/Counters.sol";

interface ISubscriptionContract {
    function canPerformAction(address user) external view returns (bool);
    function updateUsage(address user, uint256 queryIncrement, uint256 actionIncrement) external;
}

/**
 * @title ActionContract
 * @dev Contract for executing on-chain actions triggered by chat feature
 */
contract ActionContract is Ownable, ReentrancyGuard {
    using SafeERC20 for IERC20;
    using Counters for Counters.Counter;

    struct Action {
        uint256 id;
        address user;
        ActionType actionType;
        string parameters; // JSON parameters for the action
        ActionStatus status;
        uint256 createdAt;
        uint256 executedAt;
        string txHash; // Transaction hash if external call
        string result; // Result data or error message
        uint256 gasCost;
    }

    struct StakingAction {
        address token;
        address stakingContract;
        uint256 amount;
        uint256 duration;
        string actionType; // "stake", "unstake", "claim"
    }

    struct SwapAction {
        address tokenIn;
        address tokenOut;
        uint256 amountIn;
        uint256 minAmountOut;
        address dexRouter;
        uint256 deadline;
    }

    struct GovernanceAction {
        address governanceContract;
        uint256 proposalId;
        bool support; // true for yes, false for no
        string reason;
    }

    struct LiquidityAction {
        address tokenA;
        address tokenB;
        uint256 amountA;
        uint256 amountB;
        address liquidityPool;
        string actionType; // "add", "remove"
    }

    enum ActionType {
        Stake,
        Unstake,
        ClaimRewards,
        Swap,
        Vote,
        AddLiquidity,
        RemoveLiquidity,
        Custom
    }

    enum ActionStatus {
        Pending,
        Approved,
        Executing,
        Completed,
        Failed,
        Cancelled
    }

    // State variables
    ISubscriptionContract public subscriptionContract;
    Counters.Counter private _actionIdCounter;

    mapping(uint256 => Action) public actions;
    mapping(address => uint256[]) public userActions;
    mapping(ActionType => bool) public enabledActions;
    mapping(address => bool) public authorizedExecutors;

    // Supported protocols and contracts
    mapping(address => bool) public supportedTokens;
    mapping(address => bool) public supportedStakingContracts;
    mapping(address => bool) public supportedDexRouters;
    mapping(address => bool) public supportedGovernanceContracts;
    mapping(address => bool) public supportedLiquidityPools;

    // Action limits and fees
    mapping(ActionType => uint256) public actionFees; // Fees in KAIA
    mapping(ActionType => uint256) public actionLimits; // Max amount per action
    uint256 public dailyActionLimit = 10; // Max actions per user per day
    mapping(address => mapping(uint256 => uint256)) public dailyActionCount; // user => day => count

    // Emergency controls
    bool public emergencyPause = false;
    mapping(ActionType => bool) public actionPaused;

    // Events
    event ActionCreated(
        uint256 indexed actionId,
        address indexed user,
        ActionType actionType,
        string parameters
    );

    event ActionExecuted(
        uint256 indexed actionId,
        address indexed user,
        ActionStatus status,
        string result
    );

    event ActionApproved(
        uint256 indexed actionId,
        address indexed approver
    );

    event ActionCancelled(
        uint256 indexed actionId,
        address indexed user,
        string reason
    );

    event ProtocolSupported(
        address indexed protocol,
        string protocolType
    );

    // Modifiers
    modifier notPaused() {
        require(!emergencyPause, "Contract is paused");
        _;
    }

    modifier actionNotPaused(ActionType actionType) {
        require(!actionPaused[actionType], "Action type is paused");
        _;
    }

    modifier onlyAuthorizedExecutor() {
        require(authorizedExecutors[msg.sender] || msg.sender == owner(), "Not authorized executor");
        _;
    }

    modifier validActionId(uint256 actionId) {
        require(actionId > 0 && actionId <= _actionIdCounter.current(), "Invalid action ID");
        _;
    }

    constructor(address _subscriptionContract) {
        require(_subscriptionContract != address(0), "Invalid subscription contract");
        subscriptionContract = ISubscriptionContract(_subscriptionContract);
        
        // Enable all action types by default
        _enableAllActions();
        
        // Set default action fees (in KAIA tokens)
        _setDefaultFees();
    }

    /**
     * @dev Create a new action request
     * @param actionType Type of action to execute
     * @param parameters JSON parameters for the action
     */
    function createAction(
        ActionType actionType,
        string memory parameters
    ) external notPaused actionNotPaused(actionType) nonReentrant returns (uint256) {
        require(enabledActions[actionType], "Action type not enabled");
        require(bytes(parameters).length > 0, "Parameters cannot be empty");
        
        // Check subscription and daily limits
        require(subscriptionContract.canPerformAction(msg.sender), "Subscription limit exceeded");
        _checkDailyLimit(msg.sender);

        // Create action
        _actionIdCounter.increment();
        uint256 actionId = _actionIdCounter.current();

        actions[actionId] = Action({
            id: actionId,
            user: msg.sender,
            actionType: actionType,
            parameters: parameters,
            status: ActionStatus.Pending,
            createdAt: block.timestamp,
            executedAt: 0,
            txHash: "",
            result: "",
            gasCost: 0
        });

        userActions[msg.sender].push(actionId);
        _incrementDailyCount(msg.sender);

        emit ActionCreated(actionId, msg.sender, actionType, parameters);
        
        return actionId;
    }

    /**
     * @dev Execute a staking action
     * @param actionId ID of the action
     * @param stakingData Staking parameters
     */
    function executeStakingAction(
        uint256 actionId,
        StakingAction memory stakingData
    ) external onlyAuthorizedExecutor validActionId(actionId) nonReentrant {
        Action storage action = actions[actionId];
        require(action.status == ActionStatus.Approved, "Action not approved");
        require(action.actionType == ActionType.Stake || action.actionType == ActionType.Unstake, "Invalid action type");
        
        // Validate staking parameters
        require(supportedTokens[stakingData.token], "Token not supported");
        require(supportedStakingContracts[stakingData.stakingContract], "Staking contract not supported");
        
        action.status = ActionStatus.Executing;
        
        try this._executeStaking(action.user, stakingData) {
            action.status = ActionStatus.Completed;
            action.executedAt = block.timestamp;
            action.result = "Staking action completed successfully";
            
            // Update subscription usage
            subscriptionContract.updateUsage(action.user, 0, 1);
            
        } catch Error(string memory reason) {
            action.status = ActionStatus.Failed;
            action.result = reason;
        } catch {
            action.status = ActionStatus.Failed;
            action.result = "Unknown error occurred";
        }

        emit ActionExecuted(actionId, action.user, action.status, action.result);
    }

    /**
     * @dev Execute a swap action
     * @param actionId ID of the action
     * @param swapData Swap parameters
     */
    function executeSwapAction(
        uint256 actionId,
        SwapAction memory swapData
    ) external onlyAuthorizedExecutor validActionId(actionId) nonReentrant {
        Action storage action = actions[actionId];
        require(action.status == ActionStatus.Approved, "Action not approved");
        require(action.actionType == ActionType.Swap, "Invalid action type");
        
        // Validate swap parameters
        require(supportedTokens[swapData.tokenIn], "Input token not supported");
        require(supportedTokens[swapData.tokenOut], "Output token not supported");
        require(supportedDexRouters[swapData.dexRouter], "DEX router not supported");
        require(swapData.deadline > block.timestamp, "Deadline passed");
        
        action.status = ActionStatus.Executing;
        
        try this._executeSwap(action.user, swapData) {
            action.status = ActionStatus.Completed;
            action.executedAt = block.timestamp;
            action.result = "Swap completed successfully";
            
            // Update subscription usage
            subscriptionContract.updateUsage(action.user, 0, 1);
            
        } catch Error(string memory reason) {
            action.status = ActionStatus.Failed;
            action.result = reason;
        } catch {
            action.status = ActionStatus.Failed;
            action.result = "Unknown error occurred";
        }

        emit ActionExecuted(actionId, action.user, action.status, action.result);
    }

    /**
     * @dev Execute a governance action
     * @param actionId ID of the action
     * @param governanceData Governance parameters
     */
    function executeGovernanceAction(
        uint256 actionId,
        GovernanceAction memory governanceData
    ) external onlyAuthorizedExecutor validActionId(actionId) nonReentrant {
        Action storage action = actions[actionId];
        require(action.status == ActionStatus.Approved, "Action not approved");
        require(action.actionType == ActionType.Vote, "Invalid action type");
        
        // Validate governance parameters
        require(supportedGovernanceContracts[governanceData.governanceContract], "Governance contract not supported");
        
        action.status = ActionStatus.Executing;
        
        try this._executeVote(action.user, governanceData) {
            action.status = ActionStatus.Completed;
            action.executedAt = block.timestamp;
            action.result = "Vote cast successfully";
            
            // Update subscription usage
            subscriptionContract.updateUsage(action.user, 0, 1);
            
        } catch Error(string memory reason) {
            action.status = ActionStatus.Failed;
            action.result = reason;
        } catch {
            action.status = ActionStatus.Failed;
            action.result = "Unknown error occurred";
        }

        emit ActionExecuted(actionId, action.user, action.status, action.result);
    }

    /**
     * @dev Execute a liquidity action
     * @param actionId ID of the action
     * @param liquidityData Liquidity parameters
     */
    function executeLiquidityAction(
        uint256 actionId,
        LiquidityAction memory liquidityData
    ) external onlyAuthorizedExecutor validActionId(actionId) nonReentrant {
        Action storage action = actions[actionId];
        require(action.status == ActionStatus.Approved, "Action not approved");
        require(
            action.actionType == ActionType.AddLiquidity || action.actionType == ActionType.RemoveLiquidity,
            "Invalid action type"
        );
        
        // Validate liquidity parameters
        require(supportedTokens[liquidityData.tokenA], "Token A not supported");
        require(supportedTokens[liquidityData.tokenB], "Token B not supported");
        require(supportedLiquidityPools[liquidityData.liquidityPool], "Liquidity pool not supported");
        
        action.status = ActionStatus.Executing;
        
        try this._executeLiquidity(action.user, liquidityData) {
            action.status = ActionStatus.Completed;
            action.executedAt = block.timestamp;
            action.result = "Liquidity action completed successfully";
            
            // Update subscription usage
            subscriptionContract.updateUsage(action.user, 0, 1);
            
        } catch Error(string memory reason) {
            action.status = ActionStatus.Failed;
            action.result = reason;
        } catch {
            action.status = ActionStatus.Failed;
            action.result = "Unknown error occurred";
        }

        emit ActionExecuted(actionId, action.user, action.status, action.result);
    }

    /**
     * @dev Approve an action for execution
     * @param actionId ID of the action
     */
    function approveAction(uint256 actionId) external onlyAuthorizedExecutor validActionId(actionId) {
        Action storage action = actions[actionId];
        require(action.status == ActionStatus.Pending, "Action not pending");
        
        action.status = ActionStatus.Approved;
        
        emit ActionApproved(actionId, msg.sender);
    }

    /**
     * @dev Cancel an action
     * @param actionId ID of the action
     * @param reason Reason for cancellation
     */
    function cancelAction(uint256 actionId, string memory reason) external validActionId(actionId) {
        Action storage action = actions[actionId];
        require(
            msg.sender == action.user || msg.sender == owner() || authorizedExecutors[msg.sender],
            "Not authorized to cancel"
        );
        require(
            action.status == ActionStatus.Pending || action.status == ActionStatus.Approved,
            "Cannot cancel this action"
        );
        
        action.status = ActionStatus.Cancelled;
        action.result = reason;
        
        emit ActionCancelled(actionId, action.user, reason);
    }

    /**
     * @dev Get user's actions
     * @param user Address of the user
     */
    function getUserActions(address user) external view returns (uint256[] memory) {
        return userActions[user];
    }

    /**
     * @dev Get action details
     * @param actionId ID of the action
     */
    function getAction(uint256 actionId) external view validActionId(actionId) returns (Action memory) {
        return actions[actionId];
    }

    /**
     * @dev Get user's daily action count
     * @param user Address of the user
     */
    function getDailyActionCount(address user) external view returns (uint256) {
        uint256 today = block.timestamp / 86400; // Current day
        return dailyActionCount[user][today];
    }

    // Admin functions
    /**
     * @dev Add supported token
     * @param token Address of the token
     */
    function addSupportedToken(address token) external onlyOwner {
        require(token != address(0), "Invalid token address");
        supportedTokens[token] = true;
        emit ProtocolSupported(token, "token");
    }

    /**
     * @dev Add supported staking contract
     * @param stakingContract Address of the staking contract
     */
    function addSupportedStakingContract(address stakingContract) external onlyOwner {
        require(stakingContract != address(0), "Invalid staking contract");
        supportedStakingContracts[stakingContract] = true;
        emit ProtocolSupported(stakingContract, "staking");
    }

    /**
     * @dev Add supported DEX router
     * @param dexRouter Address of the DEX router
     */
    function addSupportedDexRouter(address dexRouter) external onlyOwner {
        require(dexRouter != address(0), "Invalid DEX router");
        supportedDexRouters[dexRouter] = true;
        emit ProtocolSupported(dexRouter, "dex");
    }

    /**
     * @dev Add supported governance contract
     * @param governanceContract Address of the governance contract
     */
    function addSupportedGovernanceContract(address governanceContract) external onlyOwner {
        require(governanceContract != address(0), "Invalid governance contract");
        supportedGovernanceContracts[governanceContract] = true;
        emit ProtocolSupported(governanceContract, "governance");
    }

    /**
     * @dev Authorize an executor
     * @param executor Address to authorize
     */
    function authorizeExecutor(address executor) external onlyOwner {
        require(executor != address(0), "Invalid executor address");
        authorizedExecutors[executor] = true;
    }

    /**
     * @dev Set action fee
     * @param actionType Type of action
     * @param fee Fee amount in KAIA tokens
     */
    function setActionFee(ActionType actionType, uint256 fee) external onlyOwner {
        actionFees[actionType] = fee;
    }

    /**
     * @dev Set daily action limit
     * @param limit New daily limit
     */
    function setDailyActionLimit(uint256 limit) external onlyOwner {
        dailyActionLimit = limit;
    }

    /**
     * @dev Emergency pause
     * @param paused Whether to pause the contract
     */
    function setEmergencyPause(bool paused) external onlyOwner {
        emergencyPause = paused;
    }

    /**
     * @dev Pause specific action type
     * @param actionType Type of action to pause
     * @param paused Whether to pause
     */
    function setActionPaused(ActionType actionType, bool paused) external onlyOwner {
        actionPaused[actionType] = paused;
    }

    // Internal functions
    function _checkDailyLimit(address user) private view {
        uint256 today = block.timestamp / 86400;
        require(dailyActionCount[user][today] < dailyActionLimit, "Daily action limit exceeded");
    }

    function _incrementDailyCount(address user) private {
        uint256 today = block.timestamp / 86400;
        dailyActionCount[user][today]++;
    }

    function _enableAllActions() private {
        enabledActions[ActionType.Stake] = true;
        enabledActions[ActionType.Unstake] = true;
        enabledActions[ActionType.ClaimRewards] = true;
        enabledActions[ActionType.Swap] = true;
        enabledActions[ActionType.Vote] = true;
        enabledActions[ActionType.AddLiquidity] = true;
        enabledActions[ActionType.RemoveLiquidity] = true;
        enabledActions[ActionType.Custom] = true;
    }

    function _setDefaultFees() private {
        actionFees[ActionType.Stake] = 1 * 10**18; // 1 KAIA
        actionFees[ActionType.Unstake] = 1 * 10**18; // 1 KAIA
        actionFees[ActionType.ClaimRewards] = 0.5 * 10**18; // 0.5 KAIA
        actionFees[ActionType.Swap] = 2 * 10**18; // 2 KAIA
        actionFees[ActionType.Vote] = 0.1 * 10**18; // 0.1 KAIA
        actionFees[ActionType.AddLiquidity] = 2 * 10**18; // 2 KAIA
        actionFees[ActionType.RemoveLiquidity] = 2 * 10**18; // 2 KAIA
        actionFees[ActionType.Custom] = 5 * 10**18; // 5 KAIA
    }

    // External functions for execution (called by try-catch)
    function _executeStaking(address user, StakingAction memory stakingData) external {
        // This would contain the actual staking logic
        // For now, we'll just emit an event
        require(msg.sender == address(this), "Only self-call allowed");
        // Implementation would go here
    }

    function _executeSwap(address user, SwapAction memory swapData) external {
        // This would contain the actual swap logic
        require(msg.sender == address(this), "Only self-call allowed");
        // Implementation would go here
    }

    function _executeVote(address user, GovernanceAction memory governanceData) external {
        // This would contain the actual voting logic
        require(msg.sender == address(this), "Only self-call allowed");
        // Implementation would go here
    }

    function _executeLiquidity(address user, LiquidityAction memory liquidityData) external {
        // This would contain the actual liquidity logic
        require(msg.sender == address(this), "Only self-call allowed");
        // Implementation would go here
    }
}