// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

/**
 * @title SubscriptionContract
 * @dev Contract for managing premium subscriptions with KAIA token payments
 */
contract SubscriptionContract is Ownable, ReentrancyGuard {
    using SafeERC20 for IERC20;

    struct Subscription {
        uint256 subscriptionId;
        address subscriber;
        SubscriptionTier tier;
        uint256 startTime;
        uint256 endTime;
        uint256 paidAmount;
        bool isActive;
        bool autoRenew;
        uint256 renewalCount;
    }

    struct SubscriptionPlan {
        string name;
        uint256 duration; // Duration in seconds
        uint256 price; // Price in KAIA tokens (wei)
        uint256 maxRequests; // Max API requests per day
        bool allowsAdvancedAnalytics;
        bool allowsYieldOptimization;
        bool allowsChatActions;
        bool allowsGovernanceInsights;
        string[] features;
        bool isActive;
    }

    enum SubscriptionTier {
        Free,
        Basic,
        Pro,
        Enterprise
    }

    // State variables
    IERC20 public kaiaToken;
    uint256 private _subscriptionCounter;
    
    mapping(uint256 => Subscription) public subscriptions;
    mapping(address => uint256) public userActiveSubscription;
    mapping(SubscriptionTier => SubscriptionPlan) public subscriptionPlans;
    mapping(address => uint256) public dailyRequestCount;
    mapping(address => uint256) public lastRequestDay;
    
    // Revenue tracking
    uint256 public totalRevenue;
    uint256 public totalActiveSubscriptions;
    mapping(SubscriptionTier => uint256) public tierSubscriptionCount;
    
    // Discounts and promotions
    mapping(bytes32 => uint256) public discountCodes; // code hash => discount percentage
    mapping(address => bool) public hasUsedDiscount;

    // Events
    event SubscriptionPurchased(
        uint256 indexed subscriptionId,
        address indexed subscriber,
        SubscriptionTier indexed tier,
        uint256 duration,
        uint256 amount
    );
    
    event SubscriptionRenewed(
        uint256 indexed subscriptionId,
        address indexed subscriber,
        uint256 newEndTime,
        uint256 amount
    );
    
    event SubscriptionCancelled(
        uint256 indexed subscriptionId,
        address indexed subscriber
    );
    
    event SubscriptionExpired(
        uint256 indexed subscriptionId,
        address indexed subscriber
    );
    
    event PlanUpdated(
        SubscriptionTier indexed tier,
        uint256 price,
        uint256 duration
    );
    
    event DiscountCodeAdded(bytes32 indexed codeHash, uint256 discount);
    event DiscountCodeUsed(address indexed user, bytes32 indexed codeHash, uint256 discount);

    // Custom errors
    error InvalidSubscriptionTier();
    error InsufficientKaiaBalance();
    error SubscriptionNotFound();
    error SubscriptionNotActive();
    error PlanNotActive();
    error MaxRequestsExceeded();
    error InvalidDiscountCode();
    error DiscountAlreadyUsed();
    error CannotDowngrade();

    constructor(address _kaiaToken, address initialOwner) Ownable(initialOwner) {
        kaiaToken = IERC20(_kaiaToken);
        _initializeDefaultPlans();
    }

    /**
     * @dev Initialize default subscription plans
     */
    function _initializeDefaultPlans() private {
        // Free tier
        subscriptionPlans[SubscriptionTier.Free] = SubscriptionPlan({
            name: "Free",
            duration: 0, // Unlimited
            price: 0,
            maxRequests: 100, // 100 requests per day
            allowsAdvancedAnalytics: false,
            allowsYieldOptimization: false,
            allowsChatActions: false,
            allowsGovernanceInsights: false,
            features: new string[](0),
            isActive: true
        });

        // Basic tier - 30 days
        subscriptionPlans[SubscriptionTier.Basic] = SubscriptionPlan({
            name: "Basic",
            duration: 30 days,
            price: 100 * 10**18, // 100 KAIA
            maxRequests: 1000, // 1000 requests per day
            allowsAdvancedAnalytics: true,
            allowsYieldOptimization: false,
            allowsChatActions: false,
            allowsGovernanceInsights: true,
            features: new string[](0),
            isActive: true
        });

        // Pro tier - 30 days
        subscriptionPlans[SubscriptionTier.Pro] = SubscriptionPlan({
            name: "Pro",
            duration: 30 days,
            price: 500 * 10**18, // 500 KAIA
            maxRequests: 10000, // 10k requests per day
            allowsAdvancedAnalytics: true,
            allowsYieldOptimization: true,
            allowsChatActions: true,
            allowsGovernanceInsights: true,
            features: new string[](0),
            isActive: true
        });

        // Enterprise tier - 30 days
        subscriptionPlans[SubscriptionTier.Enterprise] = SubscriptionPlan({
            name: "Enterprise",
            duration: 30 days,
            price: 2000 * 10**18, // 2000 KAIA
            maxRequests: 100000, // 100k requests per day
            allowsAdvancedAnalytics: true,
            allowsYieldOptimization: true,
            allowsChatActions: true,
            allowsGovernanceInsights: true,
            features: new string[](0),
            isActive: true
        });
    }

    /**
     * @dev Purchase a subscription
     * @param tier Subscription tier to purchase
     * @param autoRenew Whether to automatically renew
     * @param discountCode Optional discount code
     */
    function purchaseSubscription(
        SubscriptionTier tier,
        bool autoRenew,
        string memory discountCode
    ) external nonReentrant returns (uint256) {
        if (tier == SubscriptionTier.Free) {
            revert InvalidSubscriptionTier();
        }
        
        SubscriptionPlan memory plan = subscriptionPlans[tier];
        if (!plan.isActive) {
            revert PlanNotActive();
        }

        uint256 finalPrice = plan.price;
        
        // Apply discount if provided
        if (bytes(discountCode).length > 0) {
            bytes32 codeHash = keccak256(abi.encodePacked(discountCode));
            uint256 discount = discountCodes[codeHash];
            
            if (discount == 0) {
                revert InvalidDiscountCode();
            }
            
            if (hasUsedDiscount[msg.sender]) {
                revert DiscountAlreadyUsed();
            }
            
            finalPrice = (finalPrice * (100 - discount)) / 100;
            hasUsedDiscount[msg.sender] = true;
            
            emit DiscountCodeUsed(msg.sender, codeHash, discount);
        }

        // Check KAIA token balance
        if (kaiaToken.balanceOf(msg.sender) < finalPrice) {
            revert InsufficientKaiaBalance();
        }

        // Cancel existing subscription if any
        uint256 existingSubscription = userActiveSubscription[msg.sender];
        if (existingSubscription != 0) {
            _cancelSubscription(existingSubscription);
        }

        // Transfer KAIA tokens
        kaiaToken.safeTransferFrom(msg.sender, address(this), finalPrice);

        // Create new subscription
        _subscriptionCounter++;
        uint256 subscriptionId = _subscriptionCounter;
        
        uint256 startTime = block.timestamp;
        uint256 endTime = startTime + plan.duration;

        subscriptions[subscriptionId] = Subscription({
            subscriptionId: subscriptionId,
            subscriber: msg.sender,
            tier: tier,
            startTime: startTime,
            endTime: endTime,
            paidAmount: finalPrice,
            isActive: true,
            autoRenew: autoRenew,
            renewalCount: 0
        });

        userActiveSubscription[msg.sender] = subscriptionId;
        totalRevenue += finalPrice;
        totalActiveSubscriptions++;
        tierSubscriptionCount[tier]++;

        emit SubscriptionPurchased(
            subscriptionId,
            msg.sender,
            tier,
            plan.duration,
            finalPrice
        );

        return subscriptionId;
    }

    /**
     * @dev Renew an existing subscription
     * @param subscriptionId Subscription ID to renew
     */
    function renewSubscription(uint256 subscriptionId) external nonReentrant {
        Subscription storage subscription = subscriptions[subscriptionId];
        
        if (subscription.subscriptionId == 0) {
            revert SubscriptionNotFound();
        }
        
        require(subscription.subscriber == msg.sender, "Not subscription owner");
        
        SubscriptionPlan memory plan = subscriptionPlans[subscription.tier];
        if (!plan.isActive) {
            revert PlanNotActive();
        }

        // Check KAIA token balance
        if (kaiaToken.balanceOf(msg.sender) < plan.price) {
            revert InsufficientKaiaBalance();
        }

        // Transfer KAIA tokens
        kaiaToken.safeTransferFrom(msg.sender, address(this), plan.price);

        // Update subscription
        subscription.endTime = block.timestamp + plan.duration;
        subscription.isActive = true;
        subscription.renewalCount++;
        subscription.paidAmount += plan.price;

        totalRevenue += plan.price;

        emit SubscriptionRenewed(
            subscriptionId,
            msg.sender,
            subscription.endTime,
            plan.price
        );
    }

    /**
     * @dev Cancel a subscription
     * @param subscriptionId Subscription ID to cancel
     */
    function cancelSubscription(uint256 subscriptionId) external {
        Subscription storage subscription = subscriptions[subscriptionId];
        
        if (subscription.subscriptionId == 0) {
            revert SubscriptionNotFound();
        }
        
        require(
            subscription.subscriber == msg.sender || msg.sender == owner(),
            "Not authorized to cancel"
        );

        _cancelSubscription(subscriptionId);
    }

    /**
     * @dev Internal function to cancel subscription
     * @param subscriptionId Subscription ID to cancel
     */
    function _cancelSubscription(uint256 subscriptionId) internal {
        Subscription storage subscription = subscriptions[subscriptionId];
        
        if (subscription.isActive) {
            subscription.isActive = false;
            subscription.autoRenew = false;
            
            if (userActiveSubscription[subscription.subscriber] == subscriptionId) {
                userActiveSubscription[subscription.subscriber] = 0;
            }
            
            totalActiveSubscriptions--;
            tierSubscriptionCount[subscription.tier]--;
            
            emit SubscriptionCancelled(subscriptionId, subscription.subscriber);
        }
    }

    /**
     * @dev Check if user has access to a feature
     * @param user User address
     * @param feature Feature to check ("advanced_analytics", "yield_optimization", etc.)
     */
    function hasFeatureAccess(address user, string memory feature) 
        external 
        view 
        returns (bool) 
    {
        SubscriptionTier tier = getUserSubscriptionTier(user);
        SubscriptionPlan memory plan = subscriptionPlans[tier];
        
        bytes32 featureHash = keccak256(abi.encodePacked(feature));
        
        if (featureHash == keccak256("advanced_analytics")) {
            return plan.allowsAdvancedAnalytics;
        } else if (featureHash == keccak256("yield_optimization")) {
            return plan.allowsYieldOptimization;
        } else if (featureHash == keccak256("chat_actions")) {
            return plan.allowsChatActions;
        } else if (featureHash == keccak256("governance_insights")) {
            return plan.allowsGovernanceInsights;
        }
        
        return false;
    }

    /**
     * @dev Check and update daily request count
     * @param user User address
     */
    function checkAndUpdateRequestCount(address user) external returns (bool) {
        uint256 currentDay = block.timestamp / 1 days;
        
        // Reset counter if it's a new day
        if (lastRequestDay[user] < currentDay) {
            dailyRequestCount[user] = 0;
            lastRequestDay[user] = currentDay;
        }
        
        SubscriptionTier tier = getUserSubscriptionTier(user);
        SubscriptionPlan memory plan = subscriptionPlans[tier];
        
        if (dailyRequestCount[user] >= plan.maxRequests) {
            return false; // Max requests exceeded
        }
        
        dailyRequestCount[user]++;
        return true;
    }

    /**
     * @dev Get user's current subscription tier
     * @param user User address
     */
    function getUserSubscriptionTier(address user) public view returns (SubscriptionTier) {
        uint256 subscriptionId = userActiveSubscription[user];
        
        if (subscriptionId == 0) {
            return SubscriptionTier.Free;
        }
        
        Subscription memory subscription = subscriptions[subscriptionId];
        
        // Check if subscription is still valid
        if (!subscription.isActive || block.timestamp > subscription.endTime) {
            return SubscriptionTier.Free;
        }
        
        return subscription.tier;
    }

    /**
     * @dev Get subscription details
     * @param subscriptionId Subscription ID
     */
    function getSubscription(uint256 subscriptionId) 
        external 
        view 
        returns (Subscription memory) 
    {
        return subscriptions[subscriptionId];
    }

    /**
     * @dev Update subscription plan
     * @param tier Subscription tier
     * @param price New price
     * @param duration New duration
     * @param maxRequests Max daily requests
     */
    function updateSubscriptionPlan(
        SubscriptionTier tier,
        uint256 price,
        uint256 duration,
        uint256 maxRequests,
        bool allowsAdvancedAnalytics,
        bool allowsYieldOptimization,
        bool allowsChatActions,
        bool allowsGovernanceInsights
    ) external onlyOwner {
        require(tier != SubscriptionTier.Free, "Cannot modify free tier");
        
        subscriptionPlans[tier].price = price;
        subscriptionPlans[tier].duration = duration;
        subscriptionPlans[tier].maxRequests = maxRequests;
        subscriptionPlans[tier].allowsAdvancedAnalytics = allowsAdvancedAnalytics;
        subscriptionPlans[tier].allowsYieldOptimization = allowsYieldOptimization;
        subscriptionPlans[tier].allowsChatActions = allowsChatActions;
        subscriptionPlans[tier].allowsGovernanceInsights = allowsGovernanceInsights;
        
        emit PlanUpdated(tier, price, duration);
    }

    /**
     * @dev Add discount code
     * @param code Discount code
     * @param discountPercentage Discount percentage (1-100)
     */
    function addDiscountCode(string memory code, uint256 discountPercentage) 
        external 
        onlyOwner 
    {
        require(discountPercentage > 0 && discountPercentage <= 100, "Invalid discount");
        
        bytes32 codeHash = keccak256(abi.encodePacked(code));
        discountCodes[codeHash] = discountPercentage;
        
        emit DiscountCodeAdded(codeHash, discountPercentage);
    }

    /**
     * @dev Process expired subscriptions (can be called by anyone)
     */
    function processExpiredSubscriptions(uint256[] memory subscriptionIds) 
        external 
    {
        for (uint256 i = 0; i < subscriptionIds.length; i++) {
            uint256 subscriptionId = subscriptionIds[i];
            Subscription storage subscription = subscriptions[subscriptionId];
            
            if (subscription.isActive && block.timestamp > subscription.endTime) {
                if (subscription.autoRenew) {
                    // Try to auto-renew
                    SubscriptionPlan memory plan = subscriptionPlans[subscription.tier];
                    if (kaiaToken.balanceOf(subscription.subscriber) >= plan.price) {
                        // Auto-renewal logic would go here
                        // For now, we'll just mark as expired
                    }
                }
                
                // Mark as expired
                subscription.isActive = false;
                userActiveSubscription[subscription.subscriber] = 0;
                totalActiveSubscriptions--;
                tierSubscriptionCount[subscription.tier]--;
                
                emit SubscriptionExpired(subscriptionId, subscription.subscriber);
            }
        }
    }

    /**
     * @dev Withdraw contract balance (only owner)
     */
    function withdraw(uint256 amount) external onlyOwner {
        require(amount <= kaiaToken.balanceOf(address(this)), "Insufficient balance");
        kaiaToken.safeTransfer(owner(), amount);
    }

    /**
     * @dev Get platform statistics
     */
    function getPlatformStats() external view returns (
        uint256 revenue,
        uint256 activeSubscriptions,
        uint256 basicCount,
        uint256 proCount,
        uint256 enterpriseCount
    ) {
        return (
            totalRevenue,
            totalActiveSubscriptions,
            tierSubscriptionCount[SubscriptionTier.Basic],
            tierSubscriptionCount[SubscriptionTier.Pro],
            tierSubscriptionCount[SubscriptionTier.Enterprise]
        );
    }
}