// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";

/**
 * @title AnalyticsRegistry
 * @dev Contract for registering and managing analytics tasks on the Kaia blockchain
 */
contract AnalyticsRegistry is Ownable, ReentrancyGuard {
    struct AnalyticsTask {
        uint256 taskId;
        address requester;
        string taskType; // "yield_analysis", "governance_sentiment", "trade_optimization"
        string parameters;
        uint256 priority;
        uint256 createdAt;
        uint256 completedAt;
        TaskStatus status;
        string resultHash; // IPFS hash of the result
    }

    enum TaskStatus {
        Pending,
        InProgress,
        Completed,
        Failed,
        Cancelled
    }

    // State variables
    uint256 private _taskCounter;
    mapping(uint256 => AnalyticsTask) public tasks;
    mapping(address => uint256[]) public userTasks;
    mapping(string => uint256) public taskTypePricing; // Task type to KAIA cost
    
    // Service providers who can execute tasks
    mapping(address => bool) public authorizedProviders;
    mapping(address => uint256) public providerRatings;

    // Events
    event TaskRegistered(
        uint256 indexed taskId,
        address indexed requester,
        string taskType,
        uint256 priority
    );
    
    event TaskStatusUpdated(
        uint256 indexed taskId,
        TaskStatus indexed newStatus,
        address indexed provider
    );
    
    event TaskCompleted(
        uint256 indexed taskId,
        string resultHash,
        address indexed provider
    );
    
    event ProviderAuthorized(address indexed provider, bool authorized);
    
    event TaskTypePricingUpdated(string taskType, uint256 price);

    // Custom errors
    error TaskNotFound();
    error UnauthorizedProvider();
    error InvalidTaskStatus();
    error InsufficientPayment();
    error TaskAlreadyCompleted();

    constructor(address initialOwner) Ownable(initialOwner) {}

    /**
     * @dev Register a new analytics task
     * @param taskType Type of analytics task
     * @param parameters JSON string containing task parameters
     * @param priority Task priority (1-10, higher is more urgent)
     */
    function registerTask(
        string memory taskType,
        string memory parameters,
        uint256 priority
    ) external payable nonReentrant returns (uint256) {
        require(priority >= 1 && priority <= 10, "Priority must be between 1-10");
        
        uint256 requiredPayment = taskTypePricing[taskType];
        if (requiredPayment > 0 && msg.value < requiredPayment) {
            revert InsufficientPayment();
        }

        _taskCounter++;
        uint256 taskId = _taskCounter;

        tasks[taskId] = AnalyticsTask({
            taskId: taskId,
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

        emit TaskRegistered(taskId, msg.sender, taskType, priority);
        return taskId;
    }

    /**
     * @dev Update task status (only authorized providers)
     * @param taskId Task ID to update
     * @param newStatus New status
     */
    function updateTaskStatus(
        uint256 taskId,
        TaskStatus newStatus
    ) external {
        if (!authorizedProviders[msg.sender] && msg.sender != owner()) {
            revert UnauthorizedProvider();
        }
        
        if (tasks[taskId].taskId == 0) {
            revert TaskNotFound();
        }
        
        if (tasks[taskId].status == TaskStatus.Completed) {
            revert TaskAlreadyCompleted();
        }

        tasks[taskId].status = newStatus;
        
        if (newStatus == TaskStatus.InProgress) {
            // Task picked up by provider
        }

        emit TaskStatusUpdated(taskId, newStatus, msg.sender);
    }

    /**
     * @dev Complete a task with results (only authorized providers)
     * @param taskId Task ID to complete
     * @param resultHash IPFS hash of the analysis result
     */
    function completeTask(
        uint256 taskId,
        string memory resultHash
    ) external {
        if (!authorizedProviders[msg.sender] && msg.sender != owner()) {
            revert UnauthorizedProvider();
        }
        
        if (tasks[taskId].taskId == 0) {
            revert TaskNotFound();
        }
        
        if (tasks[taskId].status == TaskStatus.Completed) {
            revert TaskAlreadyCompleted();
        }

        tasks[taskId].status = TaskStatus.Completed;
        tasks[taskId].completedAt = block.timestamp;
        tasks[taskId].resultHash = resultHash;

        emit TaskCompleted(taskId, resultHash, msg.sender);
        emit TaskStatusUpdated(taskId, TaskStatus.Completed, msg.sender);
    }

    /**
     * @dev Cancel a task (only requester or owner)
     * @param taskId Task ID to cancel
     */
    function cancelTask(uint256 taskId) external {
        if (tasks[taskId].taskId == 0) {
            revert TaskNotFound();
        }
        
        require(
            tasks[taskId].requester == msg.sender || msg.sender == owner(),
            "Only requester or owner can cancel"
        );
        
        if (tasks[taskId].status == TaskStatus.Completed) {
            revert TaskAlreadyCompleted();
        }

        tasks[taskId].status = TaskStatus.Cancelled;
        emit TaskStatusUpdated(taskId, TaskStatus.Cancelled, msg.sender);
    }

    /**
     * @dev Authorize or deauthorize service providers
     * @param provider Provider address
     * @param authorized Authorization status
     */
    function setProviderAuthorization(
        address provider,
        bool authorized
    ) external onlyOwner {
        authorizedProviders[provider] = authorized;
        emit ProviderAuthorized(provider, authorized);
    }

    /**
     * @dev Set pricing for task types
     * @param taskType Task type
     * @param price Price in wei
     */
    function setTaskTypePricing(
        string memory taskType,
        uint256 price
    ) external onlyOwner {
        taskTypePricing[taskType] = price;
        emit TaskTypePricingUpdated(taskType, price);
    }

    /**
     * @dev Get task details
     * @param taskId Task ID
     */
    function getTask(uint256 taskId) external view returns (AnalyticsTask memory) {
        if (tasks[taskId].taskId == 0) {
            revert TaskNotFound();
        }
        return tasks[taskId];
    }

    /**
     * @dev Get tasks by user
     * @param user User address
     */
    function getUserTasks(address user) external view returns (uint256[] memory) {
        return userTasks[user];
    }

    /**
     * @dev Get pending tasks by type for providers
     * @param taskType Task type to filter
     */
    function getPendingTasksByType(string memory taskType) 
        external 
        view 
        returns (uint256[] memory) 
    {
        uint256[] memory pendingTasks = new uint256[](_taskCounter);
        uint256 count = 0;

        for (uint256 i = 1; i <= _taskCounter; i++) {
            if (
                tasks[i].status == TaskStatus.Pending &&
                keccak256(bytes(tasks[i].taskType)) == keccak256(bytes(taskType))
            ) {
                pendingTasks[count] = i;
                count++;
            }
        }

        // Resize array to actual count
        uint256[] memory result = new uint256[](count);
        for (uint256 i = 0; i < count; i++) {
            result[i] = pendingTasks[i];
        }

        return result;
    }

    /**
     * @dev Get current task counter
     */
    function getCurrentTaskId() external view returns (uint256) {
        return _taskCounter;
    }

    /**
     * @dev Withdraw contract balance (only owner)
     */
    function withdraw() external onlyOwner {
        uint256 balance = address(this).balance;
        require(balance > 0, "No funds to withdraw");
        
        (bool success, ) = payable(owner()).call{value: balance}("");
        require(success, "Withdrawal failed");
    }

    /**
     * @dev Emergency pause function for critical updates
     */
    function emergencyPause() external onlyOwner {
        // Implementation for emergency pause if needed
    }
}