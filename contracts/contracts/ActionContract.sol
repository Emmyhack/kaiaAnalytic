// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "./SubscriptionContract.sol";

/**
 * @title ActionContract
 * @dev Contract for executing on-chain actions triggered by the chat feature
 */
contract ActionContract is Ownable, ReentrancyGuard {
    using SafeERC20 for IERC20;

    struct PendingAction {
        uint256 actionId;
        address user;
        ActionType actionType;
        bytes actionData;
        uint256 timestamp;
        ActionStatus status;
        string chatContext; // Context from chat interaction
        uint256 gasLimit;
        address targetContract;
        uint256 value; // ETH value if needed
    }

    struct StakingAction {
        address stakingContract;
        address token;
        uint256 amount;
        uint256 duration;
    }

    struct SwapAction {
        address dexRouter;
        address tokenIn;
        address tokenOut;
        uint256 amountIn;
        uint256 minAmountOut;
        uint256 deadline;
    }

    struct GovernanceAction {
        address governanceContract;
        uint256 proposalId;
        bool support; // true for yes, false for no
        string reason;
    }

    struct YieldAction {
        address yieldProtocol;
        address lpToken;
        uint256 amount;
        string strategy; // "add_liquidity", "remove_liquidity", "compound"
    }

    enum ActionType {
        Stake,
        Unstake,
        Swap,
        Vote,
        YieldFarm,
        Transfer,
        Custom
    }

    enum ActionStatus {
        Pending,
        Approved,
        Executed,
        Failed,
        Cancelled,
        Expired
    }

    // State variables
    SubscriptionContract public subscriptionContract;
    uint256 private _actionCounter;
    uint256 public actionExpiryTime = 1 hours; // Actions expire after 1 hour
    
    mapping(uint256 => PendingAction) public pendingActions;
    mapping(address => uint256[]) public userActions;
    mapping(ActionType => bool) public enabledActions;
    mapping(address => bool) public authorizedBots; // Chat bots that can create actions
    mapping(address => uint256) public lastActionTime;
    
    // Rate limiting
    uint256 public constant ACTION_COOLDOWN = 30 seconds;
    mapping(address => uint256) public dailyActionCount;
    mapping(address => uint256) public lastActionDay;
    
    // Supported protocols
    mapping(address => bool) public supportedStakingContracts;
    mapping(address => bool) public supportedDexRouters;
    mapping(address => bool) public supportedGovernanceContracts;
    mapping(address => bool) public supportedYieldProtocols;

    // Events
    event ActionRequested(
        uint256 indexed actionId,
        address indexed user,
        ActionType indexed actionType,
        string chatContext
    );
    
    event ActionApproved(
        uint256 indexed actionId,
        address indexed user,
        address indexed approver
    );
    
    event ActionExecuted(
        uint256 indexed actionId,
        address indexed user,
        ActionType indexed actionType,
        bool success,
        bytes result
    );
    
    event ActionCancelled(
        uint256 indexed actionId,
        address indexed user,
        string reason
    );
    
    event BotAuthorized(address indexed bot, bool authorized);
    event ProtocolSupported(address indexed protocol, string protocolType, bool supported);

    // Custom errors
    error UnauthorizedBot();
    error ActionNotFound();
    error ActionNotPending();
    error ActionExpired();
    error InsufficientSubscription();
    error ActionCooldownActive();
    error DailyLimitExceeded();
    error UnsupportedProtocol();
    error InvalidActionData();
    error ExecutionFailed();

    constructor(address _subscriptionContract, address initialOwner) Ownable(initialOwner) {
        subscriptionContract = SubscriptionContract(_subscriptionContract);
        
        // Enable all action types by default
        enabledActions[ActionType.Stake] = true;
        enabledActions[ActionType.Unstake] = true;
        enabledActions[ActionType.Swap] = true;
        enabledActions[ActionType.Vote] = true;
        enabledActions[ActionType.YieldFarm] = true;
        enabledActions[ActionType.Transfer] = true;
        enabledActions[ActionType.Custom] = true;
    }

    /**
     * @dev Request a new action (called by authorized chat bots)
     * @param user User requesting the action
     * @param actionType Type of action to execute
     * @param actionData Encoded action parameters
     * @param chatContext Context from chat interaction
     * @param targetContract Target contract address
     * @param value ETH value if needed
     * @param gasLimit Gas limit for execution
     */
    function requestAction(
        address user,
        ActionType actionType,
        bytes memory actionData,
        string memory chatContext,
        address targetContract,
        uint256 value,
        uint256 gasLimit
    ) external returns (uint256) {
        if (!authorizedBots[msg.sender]) {
            revert UnauthorizedBot();
        }
        
        // Check subscription access
        if (!subscriptionContract.hasFeatureAccess(user, "chat_actions")) {
            revert InsufficientSubscription();
        }
        
        // Check rate limiting
        _checkRateLimit(user);
        
        // Validate action type
        require(enabledActions[actionType], "Action type not enabled");
        
        // Validate target contract for specific action types
        _validateTargetContract(actionType, targetContract);

        _actionCounter++;
        uint256 actionId = _actionCounter;

        pendingActions[actionId] = PendingAction({
            actionId: actionId,
            user: user,
            actionType: actionType,
            actionData: actionData,
            timestamp: block.timestamp,
            status: ActionStatus.Pending,
            chatContext: chatContext,
            gasLimit: gasLimit,
            targetContract: targetContract,
            value: value
        });

        userActions[user].push(actionId);
        lastActionTime[user] = block.timestamp;
        
        // Update daily action count
        uint256 currentDay = block.timestamp / 1 days;
        if (lastActionDay[user] < currentDay) {
            dailyActionCount[user] = 0;
            lastActionDay[user] = currentDay;
        }
        dailyActionCount[user]++;

        emit ActionRequested(actionId, user, actionType, chatContext);
        return actionId;
    }

    /**
     * @dev Approve and execute an action
     * @param actionId Action ID to execute
     */
    function executeAction(uint256 actionId) external nonReentrant {
        PendingAction storage action = pendingActions[actionId];
        
        if (action.actionId == 0) {
            revert ActionNotFound();
        }
        
        // Only the user or owner can execute
        require(
            action.user == msg.sender || msg.sender == owner(),
            "Not authorized to execute"
        );
        
        if (action.status != ActionStatus.Pending) {
            revert ActionNotPending();
        }
        
        // Check if action has expired
        if (block.timestamp > action.timestamp + actionExpiryTime) {
            action.status = ActionStatus.Expired;
            revert ActionExpired();
        }

        action.status = ActionStatus.Approved;
        emit ActionApproved(actionId, action.user, msg.sender);

        // Execute the action
        bool success;
        bytes memory result;
        
        try this._executeActionInternal(action) returns (bytes memory _result) {
            success = true;
            result = _result;
            action.status = ActionStatus.Executed;
        } catch Error(string memory reason) {
            success = false;
            result = bytes(reason);
            action.status = ActionStatus.Failed;
        } catch (bytes memory lowLevelData) {
            success = false;
            result = lowLevelData;
            action.status = ActionStatus.Failed;
        }

        emit ActionExecuted(actionId, action.user, action.actionType, success, result);
    }

    /**
     * @dev Internal function to execute actions
     * @param action Action to execute
     */
    function _executeActionInternal(PendingAction memory action) 
        external 
        returns (bytes memory) 
    {
        require(msg.sender == address(this), "Internal function");

        if (action.actionType == ActionType.Stake) {
            return _executeStakeAction(action);
        } else if (action.actionType == ActionType.Unstake) {
            return _executeUnstakeAction(action);
        } else if (action.actionType == ActionType.Swap) {
            return _executeSwapAction(action);
        } else if (action.actionType == ActionType.Vote) {
            return _executeVoteAction(action);
        } else if (action.actionType == ActionType.YieldFarm) {
            return _executeYieldAction(action);
        } else if (action.actionType == ActionType.Transfer) {
            return _executeTransferAction(action);
        } else if (action.actionType == ActionType.Custom) {
            return _executeCustomAction(action);
        }
        
        revert InvalidActionData();
    }

    /**
     * @dev Execute staking action
     */
    function _executeStakeAction(PendingAction memory action) 
        internal 
        returns (bytes memory) 
    {
        StakingAction memory stakingData = abi.decode(action.actionData, (StakingAction));
        
        // Transfer tokens to this contract first
        IERC20(stakingData.token).safeTransferFrom(
            action.user,
            address(this),
            stakingData.amount
        );
        
        // Approve staking contract
        IERC20(stakingData.token).approve(stakingData.stakingContract, stakingData.amount);
        
        // Call staking contract
        (bool success, bytes memory result) = action.targetContract.call{
            gas: action.gasLimit,
            value: action.value
        }(
            abi.encodeWithSignature(
                "stake(uint256,uint256)",
                stakingData.amount,
                stakingData.duration
            )
        );
        
        if (!success) {
            revert ExecutionFailed();
        }
        
        return result;
    }

    /**
     * @dev Execute unstaking action
     */
    function _executeUnstakeAction(PendingAction memory action) 
        internal 
        returns (bytes memory) 
    {
        // Decode unstake parameters
        (uint256 amount) = abi.decode(action.actionData, (uint256));
        
        // Call unstaking contract
        (bool success, bytes memory result) = action.targetContract.call{
            gas: action.gasLimit,
            value: action.value
        }(
            abi.encodeWithSignature("unstake(uint256)", amount)
        );
        
        if (!success) {
            revert ExecutionFailed();
        }
        
        return result;
    }

    /**
     * @dev Execute swap action
     */
    function _executeSwapAction(PendingAction memory action) 
        internal 
        returns (bytes memory) 
    {
        SwapAction memory swapData = abi.decode(action.actionData, (SwapAction));
        
        // Transfer input tokens
        IERC20(swapData.tokenIn).safeTransferFrom(
            action.user,
            address(this),
            swapData.amountIn
        );
        
        // Approve DEX router
        IERC20(swapData.tokenIn).approve(swapData.dexRouter, swapData.amountIn);
        
        // Execute swap
        (bool success, bytes memory result) = action.targetContract.call{
            gas: action.gasLimit,
            value: action.value
        }(
            abi.encodeWithSignature(
                "swapExactTokensForTokens(uint256,uint256,address[],address,uint256)",
                swapData.amountIn,
                swapData.minAmountOut,
                _getSwapPath(swapData.tokenIn, swapData.tokenOut),
                action.user,
                swapData.deadline
            )
        );
        
        if (!success) {
            revert ExecutionFailed();
        }
        
        return result;
    }

    /**
     * @dev Execute governance vote action
     */
    function _executeVoteAction(PendingAction memory action) 
        internal 
        returns (bytes memory) 
    {
        GovernanceAction memory voteData = abi.decode(action.actionData, (GovernanceAction));
        
        // Call governance contract
        (bool success, bytes memory result) = action.targetContract.call{
            gas: action.gasLimit,
            value: action.value
        }(
            abi.encodeWithSignature(
                "castVote(uint256,bool,string)",
                voteData.proposalId,
                voteData.support,
                voteData.reason
            )
        );
        
        if (!success) {
            revert ExecutionFailed();
        }
        
        return result;
    }

    /**
     * @dev Execute yield farming action
     */
    function _executeYieldAction(PendingAction memory action) 
        internal 
        returns (bytes memory) 
    {
        YieldAction memory yieldData = abi.decode(action.actionData, (YieldAction));
        
        // Handle different yield strategies
        bytes memory callData;
        if (keccak256(bytes(yieldData.strategy)) == keccak256("add_liquidity")) {
            callData = abi.encodeWithSignature("addLiquidity(uint256)", yieldData.amount);
        } else if (keccak256(bytes(yieldData.strategy)) == keccak256("remove_liquidity")) {
            callData = abi.encodeWithSignature("removeLiquidity(uint256)", yieldData.amount);
        } else if (keccak256(bytes(yieldData.strategy)) == keccak256("compound")) {
            callData = abi.encodeWithSignature("compound()");
        }
        
        (bool success, bytes memory result) = action.targetContract.call{
            gas: action.gasLimit,
            value: action.value
        }(callData);
        
        if (!success) {
            revert ExecutionFailed();
        }
        
        return result;
    }

    /**
     * @dev Execute token transfer action
     */
    function _executeTransferAction(PendingAction memory action) 
        internal 
        returns (bytes memory) 
    {
        (address token, address to, uint256 amount) = abi.decode(
            action.actionData, 
            (address, address, uint256)
        );
        
        IERC20(token).safeTransferFrom(action.user, to, amount);
        return abi.encode(true);
    }

    /**
     * @dev Execute custom action
     */
    function _executeCustomAction(PendingAction memory action) 
        internal 
        returns (bytes memory) 
    {
        // For custom actions, the actionData contains the raw call data
        (bool success, bytes memory result) = action.targetContract.call{
            gas: action.gasLimit,
            value: action.value
        }(action.actionData);
        
        if (!success) {
            revert ExecutionFailed();
        }
        
        return result;
    }

    /**
     * @dev Cancel a pending action
     * @param actionId Action ID to cancel
     * @param reason Cancellation reason
     */
    function cancelAction(uint256 actionId, string memory reason) external {
        PendingAction storage action = pendingActions[actionId];
        
        if (action.actionId == 0) {
            revert ActionNotFound();
        }
        
        require(
            action.user == msg.sender || msg.sender == owner(),
            "Not authorized to cancel"
        );
        
        require(
            action.status == ActionStatus.Pending,
            "Can only cancel pending actions"
        );

        action.status = ActionStatus.Cancelled;
        emit ActionCancelled(actionId, action.user, reason);
    }

    /**
     * @dev Get user's actions
     * @param user User address
     */
    function getUserActions(address user) external view returns (uint256[] memory) {
        return userActions[user];
    }

    /**
     * @dev Check rate limiting
     * @param user User address
     */
    function _checkRateLimit(address user) internal view {
        if (block.timestamp < lastActionTime[user] + ACTION_COOLDOWN) {
            revert ActionCooldownActive();
        }
        
        // Check daily limits based on subscription tier
        SubscriptionContract.SubscriptionTier tier = subscriptionContract.getUserSubscriptionTier(user);
        uint256 dailyLimit;
        
        if (tier == SubscriptionContract.SubscriptionTier.Free) {
            dailyLimit = 5;
        } else if (tier == SubscriptionContract.SubscriptionTier.Basic) {
            dailyLimit = 20;
        } else if (tier == SubscriptionContract.SubscriptionTier.Pro) {
            dailyLimit = 100;
        } else {
            dailyLimit = 1000; // Enterprise
        }
        
        if (dailyActionCount[user] >= dailyLimit) {
            revert DailyLimitExceeded();
        }
    }

    /**
     * @dev Validate target contract for action type
     */
    function _validateTargetContract(ActionType actionType, address targetContract) internal view {
        if (actionType == ActionType.Stake || actionType == ActionType.Unstake) {
            require(supportedStakingContracts[targetContract], "Unsupported staking contract");
        } else if (actionType == ActionType.Swap) {
            require(supportedDexRouters[targetContract], "Unsupported DEX router");
        } else if (actionType == ActionType.Vote) {
            require(supportedGovernanceContracts[targetContract], "Unsupported governance contract");
        } else if (actionType == ActionType.YieldFarm) {
            require(supportedYieldProtocols[targetContract], "Unsupported yield protocol");
        }
    }

    /**
     * @dev Get swap path for token swap
     */
    function _getSwapPath(address tokenIn, address tokenOut) internal pure returns (address[] memory) {
        address[] memory path = new address[](2);
        path[0] = tokenIn;
        path[1] = tokenOut;
        return path;
    }

    /**
     * @dev Authorize chat bot
     * @param bot Bot address
     * @param authorized Authorization status
     */
    function setBotAuthorization(address bot, bool authorized) external onlyOwner {
        authorizedBots[bot] = authorized;
        emit BotAuthorized(bot, authorized);
    }

    /**
     * @dev Set protocol support
     * @param protocol Protocol contract address
     * @param protocolType Type of protocol
     * @param supported Support status
     */
    function setProtocolSupport(
        address protocol,
        string memory protocolType,
        bool supported
    ) external onlyOwner {
        bytes32 typeHash = keccak256(abi.encodePacked(protocolType));
        
        if (typeHash == keccak256("staking")) {
            supportedStakingContracts[protocol] = supported;
        } else if (typeHash == keccak256("dex")) {
            supportedDexRouters[protocol] = supported;
        } else if (typeHash == keccak256("governance")) {
            supportedGovernanceContracts[protocol] = supported;
        } else if (typeHash == keccak256("yield")) {
            supportedYieldProtocols[protocol] = supported;
        }
        
        emit ProtocolSupported(protocol, protocolType, supported);
    }

    /**
     * @dev Set action type enabled status
     * @param actionType Action type
     * @param enabled Enabled status
     */
    function setActionTypeEnabled(ActionType actionType, bool enabled) external onlyOwner {
        enabledActions[actionType] = enabled;
    }

    /**
     * @dev Set action expiry time
     * @param _actionExpiryTime New expiry time in seconds
     */
    function setActionExpiryTime(uint256 _actionExpiryTime) external onlyOwner {
        actionExpiryTime = _actionExpiryTime;
    }

    /**
     * @dev Emergency stop all actions
     */
    function emergencyStop() external onlyOwner {
        // Disable all action types
        enabledActions[ActionType.Stake] = false;
        enabledActions[ActionType.Unstake] = false;
        enabledActions[ActionType.Swap] = false;
        enabledActions[ActionType.Vote] = false;
        enabledActions[ActionType.YieldFarm] = false;
        enabledActions[ActionType.Transfer] = false;
        enabledActions[ActionType.Custom] = false;
    }
}