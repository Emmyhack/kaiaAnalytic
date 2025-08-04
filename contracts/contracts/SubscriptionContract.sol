// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/utils/Counters.sol";

/**
 * @title SubscriptionContract
 * @dev Contract for managing premium subscriptions with KAIA token payments
 */
contract SubscriptionContract is Ownable, ReentrancyGuard {
    using SafeERC20 for IERC20;
    using Counters for Counters.Counter;

    struct Subscription {
        uint256 id;
        address subscriber;
        SubscriptionTier tier;
        uint256 startTime;
        uint256 endTime;
        uint256 paidAmount;
        bool isActive;
        bool autoRenew;
        uint256 renewalCount;
    }

    struct SubscriptionTier {
        uint256 id;
        string name;
        uint256 price; // Price in KAIA tokens (18 decimals)
        uint256 duration; // Duration in seconds
        string[] features;
        uint256 maxQueries; // Max queries per month
        uint256 maxActions; // Max on-chain actions per month
        bool isActive;
    }

    struct UserUsage {
        uint256 queriesUsed;
        uint256 actionsUsed;
        uint256 lastResetTime;
    }

    // State variables
    IERC20 public kaiaToken;
    Counters.Counter private _subscriptionIdCounter;
    Counters.Counter private _tierIdCounter;

    mapping(uint256 => Subscription) public subscriptions;
    mapping(address => uint256[]) public userSubscriptions;
    mapping(address => uint256) public activeSubscriptions; // user => subscription ID
    mapping(uint256 => SubscriptionTier) public subscriptionTiers;
    mapping(address => UserUsage) public userUsage;

    // Pricing and discounts
    mapping(uint256 => uint256) public tierDiscounts; // renewal count => discount percentage
    uint256 public referralDiscount = 10; // 10% discount for referrals
    mapping(address => address) public referrals; // subscriber => referrer

    // Revenue tracking
    uint256 public totalRevenue;
    uint256 public totalSubscribers;
    mapping(address => uint256) public referralEarnings;

    // Events
    event SubscriptionPurchased(
        uint256 indexed subscriptionId,
        address indexed subscriber,
        uint256 indexed tierId,
        uint256 amount,
        uint256 duration
    );

    event SubscriptionRenewed(
        uint256 indexed subscriptionId,
        address indexed subscriber,
        uint256 amount,
        uint256 newEndTime
    );

    event SubscriptionCancelled(
        uint256 indexed subscriptionId,
        address indexed subscriber
    );

    event TierCreated(
        uint256 indexed tierId,
        string name,
        uint256 price,
        uint256 duration
    );

    event UsageUpdated(
        address indexed user,
        uint256 queriesUsed,
        uint256 actionsUsed
    );

    event ReferralReward(
        address indexed referrer,
        address indexed subscriber,
        uint256 reward
    );

    // Modifiers
    modifier validTier(uint256 tierId) {
        require(tierId > 0 && tierId <= _tierIdCounter.current(), "Invalid tier ID");
        require(subscriptionTiers[tierId].isActive, "Tier is not active");
        _;
    }

    modifier hasActiveSubscription(address user) {
        uint256 subId = activeSubscriptions[user];
        require(subId > 0, "No active subscription");
        require(subscriptions[subId].isActive, "Subscription is not active");
        require(block.timestamp <= subscriptions[subId].endTime, "Subscription expired");
        _;
    }

    constructor(address _kaiaToken) {
        require(_kaiaToken != address(0), "Invalid KAIA token address");
        kaiaToken = IERC20(_kaiaToken);
        
        // Create default tiers
        _createDefaultTiers();
    }

    /**
     * @dev Purchase a subscription
     * @param tierId ID of the subscription tier
     * @param referrer Address of the referrer (optional)
     */
    function purchaseSubscription(
        uint256 tierId,
        address referrer
    ) external validTier(tierId) nonReentrant returns (uint256) {
        require(activeSubscriptions[msg.sender] == 0, "Already has active subscription");
        
        SubscriptionTier memory tier = subscriptionTiers[tierId];
        uint256 price = tier.price;
        
        // Apply referral discount
        if (referrer != address(0) && referrer != msg.sender) {
            price = price * (100 - referralDiscount) / 100;
            referrals[msg.sender] = referrer;
        }

        // Transfer KAIA tokens
        kaiaToken.safeTransferFrom(msg.sender, address(this), price);

        // Create subscription
        _subscriptionIdCounter.increment();
        uint256 subscriptionId = _subscriptionIdCounter.current();

        uint256 startTime = block.timestamp;
        uint256 endTime = startTime + tier.duration;

        subscriptions[subscriptionId] = Subscription({
            id: subscriptionId,
            subscriber: msg.sender,
            tier: tier,
            startTime: startTime,
            endTime: endTime,
            paidAmount: price,
            isActive: true,
            autoRenew: false,
            renewalCount: 0
        });

        userSubscriptions[msg.sender].push(subscriptionId);
        activeSubscriptions[msg.sender] = subscriptionId;

        // Initialize user usage
        userUsage[msg.sender] = UserUsage({
            queriesUsed: 0,
            actionsUsed: 0,
            lastResetTime: startTime
        });

        // Update metrics
        totalRevenue += price;
        if (userSubscriptions[msg.sender].length == 1) {
            totalSubscribers++;
        }

        // Handle referral reward
        if (referrer != address(0) && referrer != msg.sender) {
            uint256 reward = tier.price * referralDiscount / 100;
            referralEarnings[referrer] += reward;
            emit ReferralReward(referrer, msg.sender, reward);
        }

        emit SubscriptionPurchased(subscriptionId, msg.sender, tierId, price, tier.duration);
        
        return subscriptionId;
    }

    /**
     * @dev Renew subscription
     * @param subscriptionId ID of the subscription to renew
     */
    function renewSubscription(uint256 subscriptionId) external nonReentrant {
        require(subscriptionId > 0 && subscriptionId <= _subscriptionIdCounter.current(), "Invalid subscription ID");
        
        Subscription storage subscription = subscriptions[subscriptionId];
        require(subscription.subscriber == msg.sender, "Not subscription owner");
        require(subscription.isActive, "Subscription is not active");

        SubscriptionTier memory tier = subscription.tier;
        require(tier.isActive, "Tier is no longer active");

        uint256 price = tier.price;
        
        // Apply loyalty discount based on renewal count
        uint256 discount = tierDiscounts[subscription.renewalCount];
        if (discount > 0) {
            price = price * (100 - discount) / 100;
        }

        // Transfer KAIA tokens
        kaiaToken.safeTransferFrom(msg.sender, address(this), price);

        // Update subscription
        subscription.endTime += tier.duration;
        subscription.paidAmount += price;
        subscription.renewalCount++;

        // Reset usage if needed
        UserUsage storage usage = userUsage[msg.sender];
        if (block.timestamp >= usage.lastResetTime + 30 days) {
            usage.queriesUsed = 0;
            usage.actionsUsed = 0;
            usage.lastResetTime = block.timestamp;
        }

        totalRevenue += price;

        emit SubscriptionRenewed(subscriptionId, msg.sender, price, subscription.endTime);
    }

    /**
     * @dev Cancel subscription
     * @param subscriptionId ID of the subscription to cancel
     */
    function cancelSubscription(uint256 subscriptionId) external {
        require(subscriptionId > 0 && subscriptionId <= _subscriptionIdCounter.current(), "Invalid subscription ID");
        
        Subscription storage subscription = subscriptions[subscriptionId];
        require(subscription.subscriber == msg.sender, "Not subscription owner");
        require(subscription.isActive, "Subscription already cancelled");

        subscription.isActive = false;
        subscription.autoRenew = false;
        
        if (activeSubscriptions[msg.sender] == subscriptionId) {
            activeSubscriptions[msg.sender] = 0;
        }

        emit SubscriptionCancelled(subscriptionId, msg.sender);
    }

    /**
     * @dev Set auto-renewal for subscription
     * @param subscriptionId ID of the subscription
     * @param autoRenew Whether to auto-renew
     */
    function setAutoRenew(uint256 subscriptionId, bool autoRenew) external {
        require(subscriptionId > 0 && subscriptionId <= _subscriptionIdCounter.current(), "Invalid subscription ID");
        
        Subscription storage subscription = subscriptions[subscriptionId];
        require(subscription.subscriber == msg.sender, "Not subscription owner");
        require(subscription.isActive, "Subscription is not active");

        subscription.autoRenew = autoRenew;
    }

    /**
     * @dev Update user usage (only by authorized contracts)
     * @param user Address of the user
     * @param queryIncrement Number of queries to add
     * @param actionIncrement Number of actions to add
     */
    function updateUsage(
        address user,
        uint256 queryIncrement,
        uint256 actionIncrement
    ) external onlyOwner {
        require(hasValidSubscription(user), "User has no valid subscription");

        UserUsage storage usage = userUsage[user];
        
        // Reset usage if a month has passed
        if (block.timestamp >= usage.lastResetTime + 30 days) {
            usage.queriesUsed = 0;
            usage.actionsUsed = 0;
            usage.lastResetTime = block.timestamp;
        }

        usage.queriesUsed += queryIncrement;
        usage.actionsUsed += actionIncrement;

        emit UsageUpdated(user, usage.queriesUsed, usage.actionsUsed);
    }

    /**
     * @dev Check if user can perform query
     * @param user Address of the user
     */
    function canPerformQuery(address user) external view returns (bool) {
        if (!hasValidSubscription(user)) {
            return false;
        }

        uint256 subId = activeSubscriptions[user];
        Subscription memory subscription = subscriptions[subId];
        UserUsage memory usage = userUsage[user];

        return usage.queriesUsed < subscription.tier.maxQueries;
    }

    /**
     * @dev Check if user can perform action
     * @param user Address of the user
     */
    function canPerformAction(address user) external view returns (bool) {
        if (!hasValidSubscription(user)) {
            return false;
        }

        uint256 subId = activeSubscriptions[user];
        Subscription memory subscription = subscriptions[subId];
        UserUsage memory usage = userUsage[user];

        return usage.actionsUsed < subscription.tier.maxActions;
    }

    /**
     * @dev Check if user has valid subscription
     * @param user Address of the user
     */
    function hasValidSubscription(address user) public view returns (bool) {
        uint256 subId = activeSubscriptions[user];
        if (subId == 0) return false;
        
        Subscription memory subscription = subscriptions[subId];
        return subscription.isActive && block.timestamp <= subscription.endTime;
    }

    /**
     * @dev Get user's active subscription
     * @param user Address of the user
     */
    function getUserSubscription(address user) external view returns (Subscription memory) {
        uint256 subId = activeSubscriptions[user];
        require(subId > 0, "No active subscription");
        return subscriptions[subId];
    }

    /**
     * @dev Get user's usage statistics
     * @param user Address of the user
     */
    function getUserUsage(address user) external view returns (UserUsage memory) {
        return userUsage[user];
    }

    /**
     * @dev Create new subscription tier (only owner)
     * @param name Name of the tier
     * @param price Price in KAIA tokens
     * @param duration Duration in seconds
     * @param features Array of feature names
     * @param maxQueries Maximum queries per month
     * @param maxActions Maximum actions per month
     */
    function createTier(
        string memory name,
        uint256 price,
        uint256 duration,
        string[] memory features,
        uint256 maxQueries,
        uint256 maxActions
    ) external onlyOwner returns (uint256) {
        require(bytes(name).length > 0, "Name cannot be empty");
        require(price > 0, "Price must be greater than 0");
        require(duration > 0, "Duration must be greater than 0");

        _tierIdCounter.increment();
        uint256 tierId = _tierIdCounter.current();

        subscriptionTiers[tierId] = SubscriptionTier({
            id: tierId,
            name: name,
            price: price,
            duration: duration,
            features: features,
            maxQueries: maxQueries,
            maxActions: maxActions,
            isActive: true
        });

        emit TierCreated(tierId, name, price, duration);
        
        return tierId;
    }

    /**
     * @dev Update tier status
     * @param tierId ID of the tier
     * @param isActive Whether the tier is active
     */
    function updateTierStatus(uint256 tierId, bool isActive) external onlyOwner {
        require(tierId > 0 && tierId <= _tierIdCounter.current(), "Invalid tier ID");
        subscriptionTiers[tierId].isActive = isActive;
    }

    /**
     * @dev Set tier discount for renewals
     * @param renewalCount Number of renewals
     * @param discountPercentage Discount percentage (0-100)
     */
    function setTierDiscount(uint256 renewalCount, uint256 discountPercentage) external onlyOwner {
        require(discountPercentage <= 100, "Discount cannot exceed 100%");
        tierDiscounts[renewalCount] = discountPercentage;
    }

    /**
     * @dev Withdraw KAIA tokens (only owner)
     * @param amount Amount to withdraw
     */
    function withdrawTokens(uint256 amount) external onlyOwner {
        require(amount <= kaiaToken.balanceOf(address(this)), "Insufficient balance");
        kaiaToken.safeTransfer(owner(), amount);
    }

    /**
     * @dev Claim referral earnings
     */
    function claimReferralEarnings() external nonReentrant {
        uint256 earnings = referralEarnings[msg.sender];
        require(earnings > 0, "No earnings to claim");
        require(earnings <= kaiaToken.balanceOf(address(this)), "Insufficient contract balance");

        referralEarnings[msg.sender] = 0;
        kaiaToken.safeTransfer(msg.sender, earnings);
    }

    /**
     * @dev Get all subscription tiers
     */
    function getAllTiers() external view returns (SubscriptionTier[] memory) {
        uint256 tierCount = _tierIdCounter.current();
        SubscriptionTier[] memory tiers = new SubscriptionTier[](tierCount);
        
        for (uint256 i = 1; i <= tierCount; i++) {
            tiers[i-1] = subscriptionTiers[i];
        }
        
        return tiers;
    }

    /**
     * @dev Get contract statistics
     */
    function getStats() external view returns (uint256, uint256, uint256, uint256) {
        return (
            totalRevenue,
            totalSubscribers,
            _subscriptionIdCounter.current(),
            _tierIdCounter.current()
        );
    }

    // Internal functions
    function _createDefaultTiers() private {
        // Basic Tier - 30 days
        string[] memory basicFeatures = new string[](3);
        basicFeatures[0] = "Basic Analytics";
        basicFeatures[1] = "Yield Opportunities";
        basicFeatures[2] = "Limited Chat Queries";
        
        _tierIdCounter.increment();
        subscriptionTiers[1] = SubscriptionTier({
            id: 1,
            name: "Basic",
            price: 100 * 10**18, // 100 KAIA
            duration: 30 days,
            features: basicFeatures,
            maxQueries: 100,
            maxActions: 10,
            isActive: true
        });

        // Premium Tier - 30 days
        string[] memory premiumFeatures = new string[](5);
        premiumFeatures[0] = "Advanced Analytics";
        premiumFeatures[1] = "Personalized Trading Suggestions";
        premiumFeatures[2] = "Unlimited Chat Queries";
        premiumFeatures[3] = "On-chain Actions";
        premiumFeatures[4] = "Priority Support";
        
        _tierIdCounter.increment();
        subscriptionTiers[2] = SubscriptionTier({
            id: 2,
            name: "Premium",
            price: 500 * 10**18, // 500 KAIA
            duration: 30 days,
            features: premiumFeatures,
            maxQueries: 1000,
            maxActions: 100,
            isActive: true
        });

        // Enterprise Tier - 30 days
        string[] memory enterpriseFeatures = new string[](6);
        enterpriseFeatures[0] = "All Premium Features";
        enterpriseFeatures[1] = "Custom Analytics";
        enterpriseFeatures[2] = "API Access";
        enterpriseFeatures[3] = "Bulk Actions";
        enterpriseFeatures[4] = "Dedicated Support";
        enterpriseFeatures[5] = "Custom Integrations";
        
        _tierIdCounter.increment();
        subscriptionTiers[3] = SubscriptionTier({
            id: 3,
            name: "Enterprise",
            price: 2000 * 10**18, // 2000 KAIA
            duration: 30 days,
            features: enterpriseFeatures,
            maxQueries: 10000,
            maxActions: 1000,
            isActive: true
        });
    }
}