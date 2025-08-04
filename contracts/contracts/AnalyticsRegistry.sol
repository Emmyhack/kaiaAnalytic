// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/utils/Counters.sol";

/**
 * @title AnalyticsRegistry
 * @dev Contract for registering and managing analytics tasks on Kaia blockchain
 */
contract AnalyticsRegistry is Ownable, ReentrancyGuard {
    using Counters for Counters.Counter;

    struct AnalyticsTask {
        uint256 id;
        address requester;
        string taskType; // "yield_analysis", "governance_sentiment", "trade_optimization"
        string parameters; // JSON string with task parameters
        uint256 priority; // 1-5, higher is more urgent
        uint256 createdAt;
        uint256 completedAt;
        TaskStatus status;
        string resultHash; // IPFS hash or data reference
    }

    enum TaskStatus {
        Pending,
        InProgress,
        Completed,
        Failed,
        Cancelled
    }

    // State variables
    Counters.Counter private _taskIdCounter;
    mapping(uint256 => AnalyticsTask) public tasks;
    mapping(address => uint256[]) public userTasks;
    mapping(string => uint256) public taskTypeCounts;
    
    // Task queue for processing
    uint256[] public pendingTasks;
    mapping(uint256 => uint256) private _pendingTaskIndex;

    // Access control
    mapping(address => bool) public authorizedProcessors;
    
    // Events
    event TaskRegistered(
        uint256 indexed taskId,
        address indexed requester,
        string taskType,
        uint256 priority
    );
    
    event TaskStatusUpdated(
        uint256 indexed taskId,
        TaskStatus oldStatus,
        TaskStatus newStatus
    );
    
    event TaskCompleted(
        uint256 indexed taskId,
        string resultHash
    );
    
    event ProcessorAuthorized(address indexed processor);
    event ProcessorRevoked(address indexed processor);

    // Modifiers
    modifier onlyAuthorizedProcessor() {
        require(authorizedProcessors[msg.sender] || msg.sender == owner(), "Not authorized processor");
        _;
    }

    modifier validTaskId(uint256 taskId) {
        require(taskId > 0 && taskId <= _taskIdCounter.current(), "Invalid task ID");
        _;
    }

    constructor() {
        authorizedProcessors[msg.sender] = true;
    }

    /**
     * @dev Register a new analytics task
     * @param taskType Type of analytics task
     * @param parameters JSON parameters for the task
     * @param priority Task priority (1-5)
     */
    function registerTask(
        string memory taskType,
        string memory parameters,
        uint256 priority
    ) external nonReentrant returns (uint256) {
        require(bytes(taskType).length > 0, "Task type cannot be empty");
        require(priority >= 1 && priority <= 5, "Priority must be between 1 and 5");

        _taskIdCounter.increment();
        uint256 taskId = _taskIdCounter.current();

        tasks[taskId] = AnalyticsTask({
            id: taskId,
            requester: msg.sender,
            taskType: taskType,
            parameters: parameters,
            priority: priority,
            createdAt: block.timestamp,
            completedAt: 0,
            status: TaskStatus.Pending,
            resultHash: ""
        });

        userTasks[msg.sender].push(taskId);
        taskTypeCounts[taskType]++;
        
        // Add to pending queue
        _addToPendingQueue(taskId);

        emit TaskRegistered(taskId, msg.sender, taskType, priority);
        
        return taskId;
    }

    /**
     * @dev Update task status (only by authorized processors)
     * @param taskId ID of the task
     * @param newStatus New status for the task
     */
    function updateTaskStatus(
        uint256 taskId,
        TaskStatus newStatus
    ) external onlyAuthorizedProcessor validTaskId(taskId) {
        AnalyticsTask storage task = tasks[taskId];
        TaskStatus oldStatus = task.status;
        
        require(oldStatus != TaskStatus.Completed, "Cannot update completed task");
        require(oldStatus != TaskStatus.Cancelled, "Cannot update cancelled task");

        task.status = newStatus;

        if (newStatus == TaskStatus.InProgress) {
            _removeFromPendingQueue(taskId);
        } else if (newStatus == TaskStatus.Completed || newStatus == TaskStatus.Failed) {
            task.completedAt = block.timestamp;
            _removeFromPendingQueue(taskId);
        }

        emit TaskStatusUpdated(taskId, oldStatus, newStatus);
    }

    /**
     * @dev Complete a task with results
     * @param taskId ID of the task
     * @param resultHash Hash of the result data
     */
    function completeTask(
        uint256 taskId,
        string memory resultHash
    ) external onlyAuthorizedProcessor validTaskId(taskId) {
        AnalyticsTask storage task = tasks[taskId];
        
        require(task.status == TaskStatus.InProgress, "Task must be in progress");
        require(bytes(resultHash).length > 0, "Result hash cannot be empty");

        task.status = TaskStatus.Completed;
        task.completedAt = block.timestamp;
        task.resultHash = resultHash;

        _removeFromPendingQueue(taskId);

        emit TaskCompleted(taskId, resultHash);
        emit TaskStatusUpdated(taskId, TaskStatus.InProgress, TaskStatus.Completed);
    }

    /**
     * @dev Cancel a task (only by requester or owner)
     * @param taskId ID of the task to cancel
     */
    function cancelTask(uint256 taskId) external validTaskId(taskId) {
        AnalyticsTask storage task = tasks[taskId];
        
        require(
            msg.sender == task.requester || msg.sender == owner(),
            "Only requester or owner can cancel"
        );
        require(
            task.status == TaskStatus.Pending || task.status == TaskStatus.InProgress,
            "Can only cancel pending or in-progress tasks"
        );

        TaskStatus oldStatus = task.status;
        task.status = TaskStatus.Cancelled;
        task.completedAt = block.timestamp;

        _removeFromPendingQueue(taskId);

        emit TaskStatusUpdated(taskId, oldStatus, TaskStatus.Cancelled);
    }

    /**
     * @dev Get next pending task (for processors)
     */
    function getNextPendingTask() external view returns (uint256) {
        if (pendingTasks.length == 0) {
            return 0;
        }

        // Find highest priority task
        uint256 highestPriority = 0;
        uint256 selectedTask = 0;

        for (uint256 i = 0; i < pendingTasks.length; i++) {
            uint256 taskId = pendingTasks[i];
            if (tasks[taskId].priority > highestPriority) {
                highestPriority = tasks[taskId].priority;
                selectedTask = taskId;
            }
        }

        return selectedTask;
    }

    /**
     * @dev Get user's tasks
     * @param user Address of the user
     */
    function getUserTasks(address user) external view returns (uint256[] memory) {
        return userTasks[user];
    }

    /**
     * @dev Get task details
     * @param taskId ID of the task
     */
    function getTask(uint256 taskId) external view validTaskId(taskId) returns (AnalyticsTask memory) {
        return tasks[taskId];
    }

    /**
     * @dev Get pending tasks count
     */
    function getPendingTasksCount() external view returns (uint256) {
        return pendingTasks.length;
    }

    /**
     * @dev Get total tasks count
     */
    function getTotalTasksCount() external view returns (uint256) {
        return _taskIdCounter.current();
    }

    /**
     * @dev Authorize a processor
     * @param processor Address to authorize
     */
    function authorizeProcessor(address processor) external onlyOwner {
        require(processor != address(0), "Invalid processor address");
        authorizedProcessors[processor] = true;
        emit ProcessorAuthorized(processor);
    }

    /**
     * @dev Revoke processor authorization
     * @param processor Address to revoke
     */
    function revokeProcessor(address processor) external onlyOwner {
        authorizedProcessors[processor] = false;
        emit ProcessorRevoked(processor);
    }

    // Internal functions
    function _addToPendingQueue(uint256 taskId) private {
        pendingTasks.push(taskId);
        _pendingTaskIndex[taskId] = pendingTasks.length - 1;
    }

    function _removeFromPendingQueue(uint256 taskId) private {
        uint256 index = _pendingTaskIndex[taskId];
        uint256 lastIndex = pendingTasks.length - 1;

        if (index != lastIndex) {
            uint256 lastTaskId = pendingTasks[lastIndex];
            pendingTasks[index] = lastTaskId;
            _pendingTaskIndex[lastTaskId] = index;
        }

        pendingTasks.pop();
        delete _pendingTaskIndex[taskId];
    }
}