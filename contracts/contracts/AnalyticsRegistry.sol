// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/utils/Counters.sol";

/**
 * @title AnalyticsRegistry
 * @dev Manages registration of analytics tasks and their metadata
 */
contract AnalyticsRegistry is Ownable, ReentrancyGuard {
    using Counters for Counters.Counter;

    struct AnalyticsTask {
        uint256 taskId;
        address creator;
        string taskType; // "yield_analysis", "governance_sentiment", "trading_suggestions"
        string description;
        uint256 reward;
        bool isActive;
        uint256 createdAt;
        uint256 completedAt;
        address executor;
    }

    Counters.Counter private _taskIds;
    
    mapping(uint256 => AnalyticsTask) public tasks;
    mapping(address => uint256[]) public userTasks;
    mapping(string => uint256[]) public tasksByType;
    
    event TaskRegistered(uint256 indexed taskId, address indexed creator, string taskType, string description, uint256 reward);
    event TaskCompleted(uint256 indexed taskId, address indexed executor, uint256 reward);
    event TaskCancelled(uint256 indexed taskId, address indexed creator);
    
    modifier taskExists(uint256 taskId) {
        require(tasks[taskId].taskId != 0, "Task does not exist");
        _;
    }
    
    modifier taskActive(uint256 taskId) {
        require(tasks[taskId].isActive, "Task is not active");
        _;
    }
    
    modifier onlyTaskCreator(uint256 taskId) {
        require(tasks[taskId].creator == msg.sender, "Only task creator can perform this action");
        _;
    }

    constructor() Ownable(msg.sender) {}

    /**
     * @dev Register a new analytics task
     * @param taskType Type of analytics task
     * @param description Task description
     * @param reward Reward amount in KAIA tokens
     */
    function registerTask(
        string memory taskType,
        string memory description,
        uint256 reward
    ) external nonReentrant {
        require(bytes(taskType).length > 0, "Task type cannot be empty");
        require(bytes(description).length > 0, "Description cannot be empty");
        require(reward > 0, "Reward must be greater than 0");

        _taskIds.increment();
        uint256 taskId = _taskIds.current();

        AnalyticsTask memory newTask = AnalyticsTask({
            taskId: taskId,
            creator: msg.sender,
            taskType: taskType,
            description: description,
            reward: reward,
            isActive: true,
            createdAt: block.timestamp,
            completedAt: 0,
            executor: address(0)
        });

        tasks[taskId] = newTask;
        userTasks[msg.sender].push(taskId);
        tasksByType[taskType].push(taskId);

        emit TaskRegistered(taskId, msg.sender, taskType, description, reward);
    }

    /**
     * @dev Complete a task and assign reward
     * @param taskId ID of the task to complete
     * @param executor Address of the task executor
     */
    function completeTask(uint256 taskId, address executor) 
        external 
        onlyOwner 
        taskExists(taskId) 
        taskActive(taskId) 
    {
        AnalyticsTask storage task = tasks[taskId];
        task.isActive = false;
        task.completedAt = block.timestamp;
        task.executor = executor;

        emit TaskCompleted(taskId, executor, task.reward);
    }

    /**
     * @dev Cancel a task (only by creator)
     * @param taskId ID of the task to cancel
     */
    function cancelTask(uint256 taskId) 
        external 
        taskExists(taskId) 
        taskActive(taskId) 
        onlyTaskCreator(taskId) 
    {
        tasks[taskId].isActive = false;
        emit TaskCancelled(taskId, msg.sender);
    }

    /**
     * @dev Get task details
     * @param taskId ID of the task
     * @return task Task details
     */
    function getTask(uint256 taskId) external view returns (AnalyticsTask memory task) {
        require(tasks[taskId].taskId != 0, "Task does not exist");
        return tasks[taskId];
    }

    /**
     * @dev Get all tasks by type
     * @param taskType Type of tasks to retrieve
     * @return taskIds Array of task IDs
     */
    function getTasksByType(string memory taskType) external view returns (uint256[] memory taskIds) {
        return tasksByType[taskType];
    }

    /**
     * @dev Get user's tasks
     * @param user Address of the user
     * @return taskIds Array of task IDs
     */
    function getUserTasks(address user) external view returns (uint256[] memory taskIds) {
        return userTasks[user];
    }

    /**
     * @dev Get total number of tasks
     * @return Total task count
     */
    function getTotalTasks() external view returns (uint256) {
        return _taskIds.current();
    }

    /**
     * @dev Get active tasks count
     * @return Active task count
     */
    function getActiveTasksCount() external view returns (uint256) {
        uint256 activeCount = 0;
        for (uint256 i = 1; i <= _taskIds.current(); i++) {
            if (tasks[i].isActive) {
                activeCount++;
            }
        }
        return activeCount;
    }
}