// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";

/**
 * @title AnalyticsRegistry
 * @dev Registers analytics tasks for the KaiaAnalyticsAI platform
 * @notice This contract manages the registration and tracking of analytics tasks
 * @author Kaia Analytics AI Team
 */
contract AnalyticsRegistry is Ownable, ReentrancyGuard {
    
    /// @notice Structure for analytics task
    struct AnalyticsTask {
        uint256 taskId;
        address requester;
        string taskType;
        string parameters;
        uint256 timestamp;
        bool isActive;
        uint256 completionTime;
        string resultHash;
    }
    
    /// @notice Mapping from task ID to task details
    mapping(uint256 => AnalyticsTask) public tasks;
    
    /// @notice Mapping from task type to task count
    mapping(string => uint256) public taskTypeCount;
    
    /// @notice Total number of tasks registered
    uint256 public totalTasks;
    
    /// @notice Task registration fee in KAIA tokens
    uint256 public registrationFee;
    
    /// @notice Emitted when a new task is registered
    /// @param taskId The unique task identifier
    /// @param requester The address that registered the task
    /// @param taskType The type of analytics task
    /// @param parameters The task parameters
    /// @param timestamp The registration timestamp
    event TaskRegistered(
        uint256 indexed taskId,
        address indexed requester,
        string taskType,
        string parameters,
        uint256 timestamp
    );
    
    /// @notice Emitted when a task is completed
    /// @param taskId The unique task identifier
    /// @param resultHash The hash of the task result
    /// @param completionTime The completion timestamp
    event TaskCompleted(
        uint256 indexed taskId,
        string resultHash,
        uint256 completionTime
    );
    
    /// @notice Emitted when registration fee is updated
    /// @param oldFee The previous fee amount
    /// @param newFee The new fee amount
    event RegistrationFeeUpdated(uint256 oldFee, uint256 newFee);
    
    /// @dev Thrown when task type is empty
    error EmptyTaskType();
    
    /// @dev Thrown when parameters are empty
    error EmptyParameters();
    
    /// @dev Thrown when task ID doesn't exist
    error TaskNotFound();
    
    /// @dev Thrown when task is already completed
    error TaskAlreadyCompleted();
    
    /// @dev Thrown when caller is not authorized to complete task
    error NotAuthorizedToComplete();
    
    /// @dev Thrown when registration fee is insufficient
    error InsufficientRegistrationFee();

    /**
     * @notice Creates a new AnalyticsRegistry contract
     * @param _registrationFee The initial registration fee for tasks
     */
    constructor(uint256 _registrationFee) {
        registrationFee = _registrationFee;
    }
    
    /**
     * @notice Registers a new analytics task
     * @param _taskType The type of analytics task (e.g., "yield_analysis", "governance_sentiment")
     * @param _parameters The task parameters in JSON format
     * @return taskId The unique identifier for the registered task
     */
    function registerTask(
        string memory _taskType,
        string memory _parameters
    ) external payable nonReentrant returns (uint256 taskId) {
        if (bytes(_taskType).length == 0) {
            revert EmptyTaskType();
        }
        
        if (bytes(_parameters).length == 0) {
            revert EmptyParameters();
        }
        
        if (msg.value < registrationFee) {
            revert InsufficientRegistrationFee();
        }
        
        taskId = totalTasks + 1;
        totalTasks = taskId;
        taskTypeCount[_taskType]++;
        
        tasks[taskId] = AnalyticsTask({
            taskId: taskId,
            requester: msg.sender,
            taskType: _taskType,
            parameters: _parameters,
            timestamp: block.timestamp,
            isActive: true,
            completionTime: 0,
            resultHash: ""
        });
        
        emit TaskRegistered(taskId, msg.sender, _taskType, _parameters, block.timestamp);
    }
    
    /**
     * @notice Completes an analytics task with results
     * @param _taskId The ID of the task to complete
     * @param _resultHash The hash of the task result
     * @dev Only authorized addresses (analytics engines) can complete tasks
     */
    function completeTask(
        uint256 _taskId,
        string memory _resultHash
    ) external onlyOwner {
        if (_taskId == 0 || _taskId > totalTasks) {
            revert TaskNotFound();
        }
        
        AnalyticsTask storage task = tasks[_taskId];
        
        if (!task.isActive) {
            revert TaskAlreadyCompleted();
        }
        
        task.isActive = false;
        task.completionTime = block.timestamp;
        task.resultHash = _resultHash;
        
        emit TaskCompleted(_taskId, _resultHash, block.timestamp);
    }
    
    /**
     * @notice Gets task details by ID
     * @param _taskId The task ID to query
     * @return task The complete task structure
     */
    function getTask(uint256 _taskId) external view returns (AnalyticsTask memory task) {
        if (_taskId == 0 || _taskId > totalTasks) {
            revert TaskNotFound();
        }
        return tasks[_taskId];
    }
    
    /**
     * @notice Gets all active tasks for a specific type
     * @param _taskType The task type to filter by
     * @return activeTaskIds Array of active task IDs
     */
    function getActiveTasksByType(string memory _taskType) external view returns (uint256[] memory activeTaskIds) {
        uint256 count = 0;
        
        // First pass: count active tasks
        for (uint256 i = 1; i <= totalTasks; i++) {
            if (tasks[i].isActive && keccak256(bytes(tasks[i].taskType)) == keccak256(bytes(_taskType))) {
                count++;
            }
        }
        
        // Second pass: collect active task IDs
        activeTaskIds = new uint256[](count);
        uint256 index = 0;
        
        for (uint256 i = 1; i <= totalTasks; i++) {
            if (tasks[i].isActive && keccak256(bytes(tasks[i].taskType)) == keccak256(bytes(_taskType))) {
                activeTaskIds[index] = i;
                index++;
            }
        }
    }
    
    /**
     * @notice Updates the registration fee
     * @param _newFee The new registration fee amount
     * @dev Only the contract owner can update the fee
     */
    function updateRegistrationFee(uint256 _newFee) external onlyOwner {
        uint256 oldFee = registrationFee;
        registrationFee = _newFee;
        emit RegistrationFeeUpdated(oldFee, _newFee);
    }
    
    /**
     * @notice Withdraws accumulated registration fees
     * @dev Only the contract owner can withdraw fees
     */
    function withdrawFees() external onlyOwner {
        uint256 amount = address(this).balance;
        require(amount > 0, "No fees to withdraw");
        
        (bool success, ) = payable(owner()).call{value: amount}("");
        require(success, "Fee withdrawal failed");
    }
    
    /**
     * @notice Gets statistics about registered tasks
     * @return _totalTasks Total number of tasks
     * @return _activeTasks Number of active tasks
     * @return _completedTasks Number of completed tasks
     */
    function getTaskStatistics() external view returns (
        uint256 _totalTasks,
        uint256 _activeTasks,
        uint256 _completedTasks
    ) {
        _totalTasks = totalTasks;
        
        for (uint256 i = 1; i <= totalTasks; i++) {
            if (tasks[i].isActive) {
                _activeTasks++;
            } else {
                _completedTasks++;
            }
        }
    }
}