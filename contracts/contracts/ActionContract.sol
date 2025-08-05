// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

/**
 * @title ActionContract
 * @dev Executes on-chain actions triggered by the chat feature
 * @notice This contract handles automated actions like staking, voting, and trading
 * @author Kaia Analytics AI Team
 */
contract ActionContract is Ownable, ReentrancyGuard {
    
    /// @notice Structure for action request
    struct ActionRequest {
        uint256 actionId;
        address user;
        string actionType;
        string parameters;
        uint256 timestamp;
        bool isExecuted;
        bool isSuccessful;
        string result;
        uint256 gasUsed;
    }
    
    /// @notice Structure for supported action type
    struct ActionType {
        string name;
        bool isEnabled;
        uint256 gasLimit;
        uint256 fee;
        string description;
    }
    
    /// @notice Mapping from action ID to action request
    mapping(uint256 => ActionRequest) public actionRequests;
    
    /// @notice Mapping from action type to action type configuration
    mapping(string => ActionType) public actionTypes;
    
    /// @notice Mapping from user to their action requests
    mapping(address => uint256[]) public userActions;
    
    /// @notice Total number of action requests
    uint256 public totalActions;
    
    /// @notice Total fees collected
    uint256 public totalFees;
    
    /// @notice Emitted when a new action is requested
    /// @param actionId The unique action identifier
    /// @param user The user requesting the action
    /// @param actionType The type of action
    /// @param parameters The action parameters
    /// @param timestamp The request timestamp
    event ActionRequested(
        uint256 indexed actionId,
        address indexed user,
        string actionType,
        string parameters,
        uint256 timestamp
    );
    
    /// @notice Emitted when an action is executed
    /// @param actionId The action ID
    /// @param user The user who requested the action
    /// @param actionType The type of action
    /// @param isSuccessful Whether the action was successful
    /// @param result The action result
    /// @param gasUsed The gas used for execution
    event ActionExecuted(
        uint256 indexed actionId,
        address indexed user,
        string actionType,
        bool isSuccessful,
        string result,
        uint256 gasUsed
    );
    
    /// @notice Emitted when an action type is registered
    /// @param actionType The action type name
    /// @param gasLimit The gas limit for the action
    /// @param fee The fee for the action
    event ActionTypeRegistered(string actionType, uint256 gasLimit, uint256 fee);
    
    /// @dev Thrown when action type is not supported
    error ActionTypeNotSupported();
    
    /// @dev Thrown when action type is disabled
    error ActionTypeDisabled();
    
    /// @dev Thrown when action ID doesn't exist
    error ActionNotFound();
    
    /// @dev Thrown when action is already executed
    error ActionAlreadyExecuted();
    
    /// @dev Thrown when insufficient fee is provided
    error InsufficientFee();
    
    /// @dev Thrown when gas limit is exceeded
    error GasLimitExceeded();
    
    /// @dev Thrown when action execution fails
    error ActionExecutionFailed();

    /**
     * @notice Creates a new ActionContract
     */
    constructor() {
        // Register default action types
        _registerActionType("stake", 100000, 0.01 ether, "Stake tokens in a protocol");
        _registerActionType("unstake", 100000, 0.01 ether, "Unstake tokens from a protocol");
        _registerActionType("vote", 50000, 0.005 ether, "Vote on a governance proposal");
        _registerActionType("swap", 150000, 0.02 ether, "Swap tokens on a DEX");
        _registerActionType("yield_farm", 120000, 0.015 ether, "Deposit into yield farming");
        _registerActionType("withdraw_yield", 80000, 0.01 ether, "Withdraw from yield farming");
    }
    
    /**
     * @notice Requests an on-chain action
     * @param _actionType The type of action to execute
     * @param _parameters The action parameters in JSON format
     * @return actionId The unique identifier for the action request
     */
    function requestAction(
        string memory _actionType,
        string memory _parameters
    ) external payable nonReentrant returns (uint256 actionId) {
        if (!actionTypes[_actionType].isEnabled) {
            revert ActionTypeNotSupported();
        }
        
        ActionType storage actionType = actionTypes[_actionType];
        if (!actionType.isEnabled) {
            revert ActionTypeDisabled();
        }
        
        if (msg.value < actionType.fee) {
            revert InsufficientFee();
        }
        
        actionId = totalActions + 1;
        totalActions = actionId;
        totalFees += msg.value;
        
        actionRequests[actionId] = ActionRequest({
            actionId: actionId,
            user: msg.sender,
            actionType: _actionType,
            parameters: _parameters,
            timestamp: block.timestamp,
            isExecuted: false,
            isSuccessful: false,
            result: "",
            gasUsed: 0
        });
        
        userActions[msg.sender].push(actionId);
        
        emit ActionRequested(actionId, msg.sender, _actionType, _parameters, block.timestamp);
    }
    
    /**
     * @notice Executes a requested action
     * @param _actionId The action ID to execute
     * @dev Only the contract owner or authorized executors can execute actions
     */
    function executeAction(uint256 _actionId) external onlyOwner {
        if (_actionId == 0 || _actionId > totalActions) {
            revert ActionNotFound();
        }
        
        ActionRequest storage action = actionRequests[_actionId];
        if (action.isExecuted) {
            revert ActionAlreadyExecuted();
        }
        
        ActionType storage actionType = actionTypes[action.actionType];
        uint256 gasStart = gasleft();
        
        try this._executeActionInternal(_actionId) {
            action.isExecuted = true;
            action.isSuccessful = true;
            action.result = "Success";
        } catch {
            action.isExecuted = true;
            action.isSuccessful = false;
            action.result = "Failed";
        }
        
        action.gasUsed = gasStart - gasleft();
        
        if (action.gasUsed > actionType.gasLimit) {
            revert GasLimitExceeded();
        }
        
        emit ActionExecuted(
            _actionId,
            action.user,
            action.actionType,
            action.isSuccessful,
            action.result,
            action.gasUsed
        );
    }
    
    /**
     * @notice Internal function to execute different action types
     * @param _actionId The action ID to execute
     * @dev This function is called by executeAction and handles different action types
     */
    function _executeActionInternal(uint256 _actionId) external {
        // This function is called by executeAction
        // In a real implementation, this would contain the actual action logic
        // For now, we'll simulate successful execution
        require(msg.sender == address(this), "Only self-call allowed");
        
        ActionRequest storage action = actionRequests[_actionId];
        
        // Parse parameters and execute based on action type
        if (keccak256(bytes(action.actionType)) == keccak256(bytes("stake"))) {
            _executeStakeAction(action);
        } else if (keccak256(bytes(action.actionType)) == keccak256(bytes("unstake"))) {
            _executeUnstakeAction(action);
        } else if (keccak256(bytes(action.actionType)) == keccak256(bytes("vote"))) {
            _executeVoteAction(action);
        } else if (keccak256(bytes(action.actionType)) == keccak256(bytes("swap"))) {
            _executeSwapAction(action);
        } else if (keccak256(bytes(action.actionType)) == keccak256(bytes("yield_farm"))) {
            _executeYieldFarmAction(action);
        } else if (keccak256(bytes(action.actionType)) == keccak256(bytes("withdraw_yield"))) {
            _executeWithdrawYieldAction(action);
        } else {
            revert ActionTypeNotSupported();
        }
    }
    
    /**
     * @notice Executes a staking action
     * @param _action The action request
     */
    function _executeStakeAction(ActionRequest storage _action) internal {
        // Simulate staking logic
        // In a real implementation, this would interact with staking contracts
        _action.result = "Staking action executed successfully";
    }
    
    /**
     * @notice Executes an unstaking action
     * @param _action The action request
     */
    function _executeUnstakeAction(ActionRequest storage _action) internal {
        // Simulate unstaking logic
        _action.result = "Unstaking action executed successfully";
    }
    
    /**
     * @notice Executes a voting action
     * @param _action The action request
     */
    function _executeVoteAction(ActionRequest storage _action) internal {
        // Simulate voting logic
        _action.result = "Voting action executed successfully";
    }
    
    /**
     * @notice Executes a swap action
     * @param _action The action request
     */
    function _executeSwapAction(ActionRequest storage _action) internal {
        // Simulate swap logic
        _action.result = "Swap action executed successfully";
    }
    
    /**
     * @notice Executes a yield farming action
     * @param _action The action request
     */
    function _executeYieldFarmAction(ActionRequest storage _action) internal {
        // Simulate yield farming logic
        _action.result = "Yield farming action executed successfully";
    }
    
    /**
     * @notice Executes a yield withdrawal action
     * @param _action The action request
     */
    function _executeWithdrawYieldAction(ActionRequest storage _action) internal {
        // Simulate yield withdrawal logic
        _action.result = "Yield withdrawal action executed successfully";
    }
    
    /**
     * @notice Registers a new action type
     * @param _actionType The action type name
     * @param _gasLimit The gas limit for the action
     * @param _fee The fee for the action
     * @param _description The description of the action
     * @dev Only the contract owner can register action types
     */
    function registerActionType(
        string memory _actionType,
        uint256 _gasLimit,
        uint256 _fee,
        string memory _description
    ) external onlyOwner {
        _registerActionType(_actionType, _gasLimit, _fee, _description);
    }
    
    /**
     * @notice Internal function to register action types
     * @param _actionType The action type name
     * @param _gasLimit The gas limit for the action
     * @param _fee The fee for the action
     * @param _description The description of the action
     */
    function _registerActionType(
        string memory _actionType,
        uint256 _gasLimit,
        uint256 _fee,
        string memory _description
    ) internal {
        actionTypes[_actionType] = ActionType({
            name: _actionType,
            isEnabled: true,
            gasLimit: _gasLimit,
            fee: _fee,
            description: _description
        });
        
        emit ActionTypeRegistered(_actionType, _gasLimit, _fee);
    }
    
    /**
     * @notice Gets action request details by ID
     * @param _actionId The action ID to query
     * @return action The complete action request structure
     */
    function getActionRequest(uint256 _actionId) external view returns (ActionRequest memory action) {
        if (_actionId == 0 || _actionId > totalActions) {
            revert ActionNotFound();
        }
        return actionRequests[_actionId];
    }
    
    /**
     * @notice Gets all actions for a specific user
     * @param _user The user address to query
     * @return actionIds Array of action IDs for the user
     */
    function getUserActions(address _user) external view returns (uint256[] memory actionIds) {
        return userActions[_user];
    }
    
    /**
     * @notice Gets action type details
     * @param _actionType The action type to query
     * @return actionType The complete action type structure
     */
    function getActionType(string memory _actionType) external view returns (ActionType memory actionType) {
        return actionTypes[_actionType];
    }
    
    /**
     * @notice Enables or disables an action type
     * @param _actionType The action type to toggle
     * @param _isEnabled The new enabled status
     * @dev Only the contract owner can toggle action types
     */
    function toggleActionType(string memory _actionType, bool _isEnabled) external onlyOwner {
        if (!actionTypes[_actionType].isEnabled) {
            revert ActionTypeNotSupported();
        }
        
        actionTypes[_actionType].isEnabled = _isEnabled;
    }
    
    /**
     * @notice Updates action type configuration
     * @param _actionType The action type to update
     * @param _gasLimit The new gas limit
     * @param _fee The new fee
     * @param _description The new description
     * @dev Only the contract owner can update action types
     */
    function updateActionType(
        string memory _actionType,
        uint256 _gasLimit,
        uint256 _fee,
        string memory _description
    ) external onlyOwner {
        if (!actionTypes[_actionType].isEnabled) {
            revert ActionTypeNotSupported();
        }
        
        actionTypes[_actionType].gasLimit = _gasLimit;
        actionTypes[_actionType].fee = _fee;
        actionTypes[_actionType].description = _description;
    }
    
    /**
     * @notice Withdraws accumulated fees
     * @dev Only the contract owner can withdraw fees
     */
    function withdrawFees() external onlyOwner {
        uint256 amount = address(this).balance;
        require(amount > 0, "No fees to withdraw");
        
        (bool success, ) = payable(owner()).call{value: amount}("");
        require(success, "Fee withdrawal failed");
    }
    
    /**
     * @notice Gets action statistics
     * @return _totalActions Total number of action requests
     * @return _executedActions Number of executed actions
     * @return _successfulActions Number of successful actions
     * @return _totalFees Total fees collected
     */
    function getActionStatistics() external view returns (
        uint256 _totalActions,
        uint256 _executedActions,
        uint256 _successfulActions,
        uint256 _totalFees
    ) {
        _totalActions = totalActions;
        _totalFees = totalFees;
        
        for (uint256 i = 1; i <= totalActions; i++) {
            if (actionRequests[i].isExecuted) {
                _executedActions++;
                if (actionRequests[i].isSuccessful) {
                    _successfulActions++;
                }
            }
        }
    }
}