// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/utils/Counters.sol";

/**
 * @title ActionContract
 * @dev Executes on-chain actions triggered by the chat feature
 */
contract ActionContract is Ownable, ReentrancyGuard {
    using Counters for Counters.Counter;

    struct Action {
        uint256 actionId;
        address user;
        string actionType; // "stake", "vote", "swap", "yield_farm"
        bytes actionData; // Encoded action parameters
        uint256 timestamp;
        bool isExecuted;
        bool isSuccessful;
        string result; // Execution result or error message
        uint256 gasUsed;
    }

    struct StakingAction {
        address token;
        uint256 amount;
        uint256 lockPeriod;
    }

    struct VotingAction {
        uint256 proposalId;
        bool support;
        uint256 weight;
    }

    struct SwapAction {
        address tokenIn;
        address tokenOut;
        uint256 amountIn;
        uint256 minAmountOut;
    }

    struct YieldFarmAction {
        address pool;
        uint256 amount;
        uint256 lockPeriod;
    }

    Counters.Counter private _actionIds;
    
    mapping(uint256 => Action) public actions;
    mapping(address => uint256[]) public userActions;
    mapping(string => uint256[]) public actionsByType;
    mapping(address => bool) public authorizedExecutors;
    
    // Mock token addresses for testing (replace with actual addresses)
    address public constant MOCK_KAIA_TOKEN = 0x1234567890123456789012345678901234567890;
    address public constant MOCK_STAKING_POOL = 0x2345678901234567890123456789012345678901;
    address public constant MOCK_GOVERNANCE = 0x3456789012345678901234567890123456789012;
    address public constant MOCK_DEX = 0x4567890123456789012345678901234567890123;
    
    event ActionCreated(uint256 indexed actionId, address indexed user, string actionType);
    event ActionExecuted(uint256 indexed actionId, bool success, string result, uint256 gasUsed);
    event ExecutorAuthorized(address indexed executor);
    event ExecutorRevoked(address indexed executor);
    
    modifier actionExists(uint256 actionId) {
        require(actions[actionId].actionId != 0, "Action does not exist");
        _;
    }
    
    modifier onlyAuthorizedExecutor() {
        require(msg.sender == owner() || authorizedExecutors[msg.sender], "Not authorized executor");
        _;
    }

    constructor() Ownable(msg.sender) {}

    /**
     * @dev Authorize an executor
     * @param executor Address to authorize
     */
    function authorizeExecutor(address executor) external onlyOwner {
        authorizedExecutors[executor] = true;
        emit ExecutorAuthorized(executor);
    }

    /**
     * @dev Revoke executor authorization
     * @param executor Address to revoke
     */
    function revokeExecutor(address executor) external onlyOwner {
        authorizedExecutors[executor] = false;
        emit ExecutorRevoked(executor);
    }

    /**
     * @dev Create a staking action
     * @param token Token address to stake
     * @param amount Amount to stake
     * @param lockPeriod Lock period in seconds
     */
    function createStakingAction(
        address token,
        uint256 amount,
        uint256 lockPeriod
    ) external nonReentrant {
        require(token != address(0), "Invalid token address");
        require(amount > 0, "Amount must be greater than 0");
        require(lockPeriod > 0, "Lock period must be greater than 0");

        StakingAction memory stakingData = StakingAction({
            token: token,
            amount: amount,
            lockPeriod: lockPeriod
        });

        bytes memory actionData = abi.encode(stakingData);
        _createAction("stake", actionData);
    }

    /**
     * @dev Create a voting action
     * @param proposalId ID of the governance proposal
     * @param support Whether to support the proposal
     * @param weight Voting weight
     */
    function createVotingAction(
        uint256 proposalId,
        bool support,
        uint256 weight
    ) external nonReentrant {
        require(proposalId > 0, "Invalid proposal ID");
        require(weight > 0, "Voting weight must be greater than 0");

        VotingAction memory votingData = VotingAction({
            proposalId: proposalId,
            support: support,
            weight: weight
        });

        bytes memory actionData = abi.encode(votingData);
        _createAction("vote", actionData);
    }

    /**
     * @dev Create a swap action
     * @param tokenIn Input token address
     * @param tokenOut Output token address
     * @param amountIn Amount to swap
     * @param minAmountOut Minimum amount to receive
     */
    function createSwapAction(
        address tokenIn,
        address tokenOut,
        uint256 amountIn,
        uint256 minAmountOut
    ) external nonReentrant {
        require(tokenIn != address(0), "Invalid input token");
        require(tokenOut != address(0), "Invalid output token");
        require(amountIn > 0, "Input amount must be greater than 0");
        require(minAmountOut > 0, "Minimum output must be greater than 0");

        SwapAction memory swapData = SwapAction({
            tokenIn: tokenIn,
            tokenOut: tokenOut,
            amountIn: amountIn,
            minAmountOut: minAmountOut
        });

        bytes memory actionData = abi.encode(swapData);
        _createAction("swap", actionData);
    }

    /**
     * @dev Create a yield farming action
     * @param pool Pool address
     * @param amount Amount to deposit
     * @param lockPeriod Lock period in seconds
     */
    function createYieldFarmAction(
        address pool,
        uint256 amount,
        uint256 lockPeriod
    ) external nonReentrant {
        require(pool != address(0), "Invalid pool address");
        require(amount > 0, "Amount must be greater than 0");
        require(lockPeriod > 0, "Lock period must be greater than 0");

        YieldFarmAction memory farmData = YieldFarmAction({
            pool: pool,
            amount: amount,
            lockPeriod: lockPeriod
        });

        bytes memory actionData = abi.encode(farmData);
        _createAction("yield_farm", actionData);
    }

    /**
     * @dev Execute an action
     * @param actionId ID of the action to execute
     */
    function executeAction(uint256 actionId) 
        external 
        onlyAuthorizedExecutor 
        actionExists(actionId) 
        nonReentrant 
    {
        Action storage action = actions[actionId];
        require(!action.isExecuted, "Action already executed");

        uint256 gasStart = gasleft();
        bool success = false;
        string memory result = "";

        try this._executeActionInternal(actionId) {
            success = true;
            result = "Action executed successfully";
        } catch Error(string memory reason) {
            success = false;
            result = reason;
        } catch {
            success = false;
            result = "Action execution failed";
        }

        action.isExecuted = true;
        action.isSuccessful = success;
        action.result = result;
        action.gasUsed = gasStart - gasleft();

        emit ActionExecuted(actionId, success, result, action.gasUsed);
    }

    /**
     * @dev Internal function to execute action (called via try-catch)
     * @param actionId ID of the action to execute
     */
    function _executeActionInternal(uint256 actionId) external {
        require(msg.sender == address(this), "Only self-call allowed");
        
        Action storage action = actions[actionId];
        
        if (keccak256(bytes(action.actionType)) == keccak256(bytes("stake"))) {
            _executeStakingAction(actionId);
        } else if (keccak256(bytes(action.actionType)) == keccak256(bytes("vote"))) {
            _executeVotingAction(actionId);
        } else if (keccak256(bytes(action.actionType)) == keccak256(bytes("swap"))) {
            _executeSwapAction(actionId);
        } else if (keccak256(bytes(action.actionType)) == keccak256(bytes("yield_farm"))) {
            _executeYieldFarmAction(actionId);
        } else {
            revert("Unknown action type");
        }
    }

    /**
     * @dev Execute staking action
     * @param actionId ID of the action
     */
    function _executeStakingAction(uint256 actionId) internal {
        Action storage action = actions[actionId];
        StakingAction memory stakingData = abi.decode(action.actionData, (StakingAction));
        
        // Mock staking execution - in real implementation, call actual staking contract
        IERC20 token = IERC20(stakingData.token);
        require(token.transferFrom(action.user, MOCK_STAKING_POOL, stakingData.amount), "Staking failed");
    }

    /**
     * @dev Execute voting action
     * @param actionId ID of the action
     */
    function _executeVotingAction(uint256 actionId) internal {
        Action storage action = actions[actionId];
        VotingAction memory votingData = abi.decode(action.actionData, (VotingAction));
        
        // Mock voting execution - in real implementation, call actual governance contract
        // This would typically involve calling a governance contract's vote function
    }

    /**
     * @dev Execute swap action
     * @param actionId ID of the action
     */
    function _executeSwapAction(uint256 actionId) internal {
        Action storage action = actions[actionId];
        SwapAction memory swapData = abi.decode(action.actionData, (SwapAction));
        
        // Mock swap execution - in real implementation, call actual DEX contract
        IERC20 tokenIn = IERC20(swapData.tokenIn);
        require(tokenIn.transferFrom(action.user, MOCK_DEX, swapData.amountIn), "Swap failed");
    }

    /**
     * @dev Execute yield farming action
     * @param actionId ID of the action
     */
    function _executeYieldFarmAction(uint256 actionId) internal {
        Action storage action = actions[actionId];
        YieldFarmAction memory farmData = abi.decode(action.actionData, (YieldFarmAction));
        
        // Mock yield farming execution - in real implementation, call actual farming contract
        IERC20 token = IERC20(MOCK_KAIA_TOKEN);
        require(token.transferFrom(action.user, farmData.pool, farmData.amount), "Yield farming failed");
    }

    /**
     * @dev Create an action
     * @param actionType Type of action
     * @param actionData Encoded action data
     */
    function _createAction(string memory actionType, bytes memory actionData) internal {
        _actionIds.increment();
        uint256 actionId = _actionIds.current();

        Action memory newAction = Action({
            actionId: actionId,
            user: msg.sender,
            actionType: actionType,
            actionData: actionData,
            timestamp: block.timestamp,
            isExecuted: false,
            isSuccessful: false,
            result: "",
            gasUsed: 0
        });

        actions[actionId] = newAction;
        userActions[msg.sender].push(actionId);
        actionsByType[actionType].push(actionId);

        emit ActionCreated(actionId, msg.sender, actionType);
    }

    /**
     * @dev Get action details
     * @param actionId ID of the action
     * @return action Action details
     */
    function getAction(uint256 actionId) external view returns (Action memory action) {
        require(actions[actionId].actionId != 0, "Action does not exist");
        return actions[actionId];
    }

    /**
     * @dev Get user's actions
     * @param user User address
     * @return actionIds Array of action IDs
     */
    function getUserActions(address user) external view returns (uint256[] memory actionIds) {
        return userActions[user];
    }

    /**
     * @dev Get actions by type
     * @param actionType Type of actions to retrieve
     * @return actionIds Array of action IDs
     */
    function getActionsByType(string memory actionType) external view returns (uint256[] memory actionIds) {
        return actionsByType[actionType];
    }

    /**
     * @dev Get total number of actions
     * @return Total action count
     */
    function getTotalActions() external view returns (uint256) {
        return _actionIds.current();
    }

    /**
     * @dev Get executed actions count
     * @return Executed action count
     */
    function getExecutedActionsCount() external view returns (uint256) {
        uint256 executedCount = 0;
        for (uint256 i = 1; i <= _actionIds.current(); i++) {
            if (actions[i].isExecuted) {
                executedCount++;
            }
        }
        return executedCount;
    }
}