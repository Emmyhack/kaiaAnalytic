// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/utils/Counters.sol";

/**
 * @title SubscriptionContract
 * @dev Manages premium subscriptions with KAIA token payments
 */
contract SubscriptionContract is Ownable, ReentrancyGuard {
    using Counters for Counters.Counter;

    struct SubscriptionPlan {
        uint256 planId;
        string name;
        uint256 price; // Price in KAIA tokens (with 18 decimals)
        uint256 duration; // Duration in seconds
        bool isActive;
        string[] features;
    }

    struct UserSubscription {
        uint256 subscriptionId;
        address user;
        uint256 planId;
        uint256 startTime;
        uint256 endTime;
        bool isActive;
        uint256 lastPayment;
    }

    IERC20 public kaiaToken;
    
    Counters.Counter private _planIds;
    Counters.Counter private _subscriptionIds;
    
    mapping(uint256 => SubscriptionPlan) public subscriptionPlans;
    mapping(uint256 => UserSubscription) public userSubscriptions;
    mapping(address => uint256) public userActiveSubscription;
    mapping(address => uint256[]) public userSubscriptionHistory;
    
    uint256 public constant PRECISION = 1e18;
    uint256 public constant MIN_SUBSCRIPTION_DURATION = 1 days;
    uint256 public constant MAX_SUBSCRIPTION_DURATION = 365 days;
    
    event PlanCreated(uint256 indexed planId, string name, uint256 price, uint256 duration);
    event PlanUpdated(uint256 indexed planId, string name, uint256 price, uint256 duration);
    event PlanDeactivated(uint256 indexed planId);
    event SubscriptionPurchased(uint256 indexed subscriptionId, address indexed user, uint256 planId, uint256 startTime, uint256 endTime);
    event SubscriptionRenewed(uint256 indexed subscriptionId, address indexed user, uint256 newEndTime);
    event SubscriptionCancelled(uint256 indexed subscriptionId, address indexed user);
    event PaymentReceived(address indexed user, uint256 amount, uint256 planId);
    
    modifier planExists(uint256 planId) {
        require(subscriptionPlans[planId].planId != 0, "Plan does not exist");
        _;
    }
    
    modifier planActive(uint256 planId) {
        require(subscriptionPlans[planId].isActive, "Plan is not active");
        _;
    }
    
    modifier subscriptionExists(uint256 subscriptionId) {
        require(userSubscriptions[subscriptionId].subscriptionId != 0, "Subscription does not exist");
        _;
    }

    constructor(address _kaiaToken) Ownable(msg.sender) {
        require(_kaiaToken != address(0), "Invalid KAIA token address");
        kaiaToken = IERC20(_kaiaToken);
    }

    /**
     * @dev Create a new subscription plan
     * @param name Plan name
     * @param price Price in KAIA tokens
     * @param duration Duration in seconds
     * @param features Array of plan features
     */
    function createPlan(
        string memory name,
        uint256 price,
        uint256 duration,
        string[] memory features
    ) external onlyOwner {
        require(bytes(name).length > 0, "Plan name cannot be empty");
        require(price > 0, "Price must be greater than 0");
        require(duration >= MIN_SUBSCRIPTION_DURATION, "Duration too short");
        require(duration <= MAX_SUBSCRIPTION_DURATION, "Duration too long");

        _planIds.increment();
        uint256 planId = _planIds.current();

        SubscriptionPlan memory newPlan = SubscriptionPlan({
            planId: planId,
            name: name,
            price: price,
            duration: duration,
            isActive: true,
            features: features
        });

        subscriptionPlans[planId] = newPlan;

        emit PlanCreated(planId, name, price, duration);
    }

    /**
     * @dev Update an existing subscription plan
     * @param planId ID of the plan to update
     * @param name New plan name
     * @param price New price in KAIA tokens
     * @param duration New duration in seconds
     * @param features New array of plan features
     */
    function updatePlan(
        uint256 planId,
        string memory name,
        uint256 price,
        uint256 duration,
        string[] memory features
    ) external onlyOwner planExists(planId) {
        require(bytes(name).length > 0, "Plan name cannot be empty");
        require(price > 0, "Price must be greater than 0");
        require(duration >= MIN_SUBSCRIPTION_DURATION, "Duration too short");
        require(duration <= MAX_SUBSCRIPTION_DURATION, "Duration too long");

        SubscriptionPlan storage plan = subscriptionPlans[planId];
        plan.name = name;
        plan.price = price;
        plan.duration = duration;
        plan.features = features;

        emit PlanUpdated(planId, name, price, duration);
    }

    /**
     * @dev Deactivate a subscription plan
     * @param planId ID of the plan to deactivate
     */
    function deactivatePlan(uint256 planId) external onlyOwner planExists(planId) {
        subscriptionPlans[planId].isActive = false;
        emit PlanDeactivated(planId);
    }

    /**
     * @dev Purchase a subscription
     * @param planId ID of the plan to purchase
     */
    function purchaseSubscription(uint256 planId) 
        external 
        nonReentrant 
        planExists(planId) 
        planActive(planId) 
    {
        SubscriptionPlan memory plan = subscriptionPlans[planId];
        
        // Check if user has sufficient KAIA tokens
        require(kaiaToken.balanceOf(msg.sender) >= plan.price, "Insufficient KAIA balance");
        
        // Transfer KAIA tokens from user to contract
        require(kaiaToken.transferFrom(msg.sender, address(this), plan.price), "Token transfer failed");

        _subscriptionIds.increment();
        uint256 subscriptionId = _subscriptionIds.current();

        uint256 startTime = block.timestamp;
        uint256 endTime = startTime + plan.duration;

        UserSubscription memory newSubscription = UserSubscription({
            subscriptionId: subscriptionId,
            user: msg.sender,
            planId: planId,
            startTime: startTime,
            endTime: endTime,
            isActive: true,
            lastPayment: startTime
        });

        userSubscriptions[subscriptionId] = newSubscription;
        userActiveSubscription[msg.sender] = subscriptionId;
        userSubscriptionHistory[msg.sender].push(subscriptionId);

        emit SubscriptionPurchased(subscriptionId, msg.sender, planId, startTime, endTime);
        emit PaymentReceived(msg.sender, plan.price, planId);
    }

    /**
     * @dev Renew an existing subscription
     * @param subscriptionId ID of the subscription to renew
     */
    function renewSubscription(uint256 subscriptionId) 
        external 
        nonReentrant 
        subscriptionExists(subscriptionId) 
    {
        UserSubscription storage subscription = userSubscriptions[subscriptionId];
        require(subscription.user == msg.sender, "Not subscription owner");
        require(subscription.isActive, "Subscription is not active");
        
        SubscriptionPlan memory plan = subscriptionPlans[subscription.subscriptionId];
        require(plan.isActive, "Plan is not active");

        // Check if user has sufficient KAIA tokens
        require(kaiaToken.balanceOf(msg.sender) >= plan.price, "Insufficient KAIA balance");
        
        // Transfer KAIA tokens from user to contract
        require(kaiaToken.transferFrom(msg.sender, address(this), plan.price), "Token transfer failed");

        // Extend subscription
        subscription.endTime += plan.duration;
        subscription.lastPayment = block.timestamp;

        emit SubscriptionRenewed(subscriptionId, msg.sender, subscription.endTime);
        emit PaymentReceived(msg.sender, plan.price, plan.planId);
    }

    /**
     * @dev Cancel a subscription
     * @param subscriptionId ID of the subscription to cancel
     */
    function cancelSubscription(uint256 subscriptionId) 
        external 
        subscriptionExists(subscriptionId) 
    {
        UserSubscription storage subscription = userSubscriptions[subscriptionId];
        require(subscription.user == msg.sender, "Not subscription owner");
        require(subscription.isActive, "Subscription is not active");

        subscription.isActive = false;
        
        // Clear active subscription if this is the user's active subscription
        if (userActiveSubscription[msg.sender] == subscriptionId) {
            userActiveSubscription[msg.sender] = 0;
        }

        emit SubscriptionCancelled(subscriptionId, msg.sender);
    }

    /**
     * @dev Check if user has active subscription
     * @param user User address
     * @return bool True if user has active subscription
     */
    function hasActiveSubscription(address user) external view returns (bool) {
        uint256 subscriptionId = userActiveSubscription[user];
        if (subscriptionId == 0) return false;
        
        UserSubscription memory subscription = userSubscriptions[subscriptionId];
        return subscription.isActive && subscription.endTime > block.timestamp;
    }

    /**
     * @dev Get user's active subscription
     * @param user User address
     * @return subscription User's active subscription details
     */
    function getUserActiveSubscription(address user) external view returns (UserSubscription memory subscription) {
        uint256 subscriptionId = userActiveSubscription[user];
        if (subscriptionId == 0) {
            return UserSubscription({
                subscriptionId: 0,
                user: address(0),
                planId: 0,
                startTime: 0,
                endTime: 0,
                isActive: false,
                lastPayment: 0
            });
        }
        return userSubscriptions[subscriptionId];
    }

    /**
     * @dev Get subscription plan details
     * @param planId ID of the plan
     * @return plan Plan details
     */
    function getPlan(uint256 planId) external view returns (SubscriptionPlan memory plan) {
        require(subscriptionPlans[planId].planId != 0, "Plan does not exist");
        return subscriptionPlans[planId];
    }

    /**
     * @dev Get all active plans
     * @return planIds Array of active plan IDs
     */
    function getActivePlans() external view returns (uint256[] memory planIds) {
        uint256 totalPlans = _planIds.current();
        uint256 activeCount = 0;
        
        // Count active plans
        for (uint256 i = 1; i <= totalPlans; i++) {
            if (subscriptionPlans[i].isActive) {
                activeCount++;
            }
        }
        
        // Create array of active plan IDs
        planIds = new uint256[](activeCount);
        uint256 index = 0;
        for (uint256 i = 1; i <= totalPlans; i++) {
            if (subscriptionPlans[i].isActive) {
                planIds[index] = i;
                index++;
            }
        }
    }

    /**
     * @dev Withdraw accumulated KAIA tokens (only owner)
     * @param amount Amount to withdraw
     */
    function withdrawTokens(uint256 amount) external onlyOwner {
        require(amount <= kaiaToken.balanceOf(address(this)), "Insufficient contract balance");
        require(kaiaToken.transfer(msg.sender, amount), "Token transfer failed");
    }

    /**
     * @dev Get total number of subscription plans
     * @return Total plan count
     */
    function getTotalPlans() external view returns (uint256) {
        return _planIds.current();
    }

    /**
     * @dev Get total number of subscriptions
     * @return Total subscription count
     */
    function getTotalSubscriptions() external view returns (uint256) {
        return _subscriptionIds.current();
    }
}