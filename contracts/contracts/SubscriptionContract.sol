// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

/**
 * @title SubscriptionContract
 * @dev Manages premium subscriptions with KAIA token payments
 * @notice This contract handles subscription tiers and payments for the KaiaAnalyticsAI platform
 * @author Kaia Analytics AI Team
 */
contract SubscriptionContract is Ownable, ReentrancyGuard {
    
    /// @notice Structure for subscription tier
    struct SubscriptionTier {
        uint256 tierId;
        string name;
        uint256 price; // Price in KAIA tokens
        uint256 duration; // Duration in seconds
        bool isActive;
        string[] features;
    }
    
    /// @notice Structure for user subscription
    struct UserSubscription {
        uint256 subscriptionId;
        address user;
        uint256 tierId;
        uint256 startTime;
        uint256 endTime;
        bool isActive;
        uint256 amountPaid;
    }
    
    /// @notice KAIA token contract address
    IERC20 public kaiaToken;
    
    /// @notice Mapping from tier ID to subscription tier
    mapping(uint256 => SubscriptionTier) public subscriptionTiers;
    
    /// @notice Mapping from user address to active subscription
    mapping(address => UserSubscription) public userSubscriptions;
    
    /// @notice Mapping from subscription ID to user subscription
    mapping(uint256 => UserSubscription) public subscriptions;
    
    /// @notice Total number of subscription tiers
    uint256 public totalTiers;
    
    /// @notice Total number of subscriptions
    uint256 public totalSubscriptions;
    
    /// @notice Total revenue collected in KAIA tokens
    uint256 public totalRevenue;
    
    /// @notice Emitted when a new subscription tier is created
    /// @param tierId The unique tier identifier
    /// @param name The tier name
    /// @param price The tier price in KAIA tokens
    /// @param duration The subscription duration
    event SubscriptionTierCreated(
        uint256 indexed tierId,
        string name,
        uint256 price,
        uint256 duration
    );
    
    /// @notice Emitted when a subscription is purchased
    /// @param subscriptionId The unique subscription identifier
    /// @param user The user address
    /// @param tierId The tier ID
    /// @param amountPaid The amount paid in KAIA tokens
    /// @param startTime The subscription start time
    /// @param endTime The subscription end time
    event SubscriptionPurchased(
        uint256 indexed subscriptionId,
        address indexed user,
        uint256 tierId,
        uint256 amountPaid,
        uint256 startTime,
        uint256 endTime
    );
    
    /// @notice Emitted when a subscription is renewed
    /// @param subscriptionId The subscription ID
    /// @param user The user address
    /// @param newEndTime The new end time
    /// @param amountPaid The amount paid for renewal
    event SubscriptionRenewed(
        uint256 indexed subscriptionId,
        address indexed user,
        uint256 newEndTime,
        uint256 amountPaid
    );
    
    /// @notice Emitted when a subscription is cancelled
    /// @param subscriptionId The subscription ID
    /// @param user The user address
    /// @param refundAmount The refund amount in KAIA tokens
    event SubscriptionCancelled(
        uint256 indexed subscriptionId,
        address indexed user,
        uint256 refundAmount
    );
    
    /// @dev Thrown when tier ID doesn't exist
    error TierNotFound();
    
    /// @dev Thrown when tier is not active
    error TierNotActive();
    
    /// @dev Thrown when user already has an active subscription
    error UserAlreadySubscribed();
    
    /// @dev Thrown when subscription doesn't exist
    error SubscriptionNotFound();
    
    /// @dev Thrown when subscription is not active
    error SubscriptionNotActive();
    
    /// @dev Thrown when insufficient KAIA tokens are provided
    error InsufficientTokens();
    
    /// @dev Thrown when trying to cancel a non-existent subscription
    error CannotCancelSubscription();
    
    /// @dev Thrown when caller is not authorized
    error NotAuthorized();

    /**
     * @notice Creates a new SubscriptionContract
     * @param _kaiaToken The address of the KAIA token contract
     */
    constructor(address _kaiaToken) {
        require(_kaiaToken != address(0), "Invalid token address");
        kaiaToken = IERC20(_kaiaToken);
    }
    
    /**
     * @notice Creates a new subscription tier
     * @param _name The tier name
     * @param _price The tier price in KAIA tokens
     * @param _duration The subscription duration in seconds
     * @param _features Array of features included in this tier
     * @return tierId The unique identifier for the created tier
     */
    function createSubscriptionTier(
        string memory _name,
        uint256 _price,
        uint256 _duration,
        string[] memory _features
    ) external onlyOwner returns (uint256 tierId) {
        tierId = totalTiers + 1;
        totalTiers = tierId;
        
        subscriptionTiers[tierId] = SubscriptionTier({
            tierId: tierId,
            name: _name,
            price: _price,
            duration: _duration,
            isActive: true,
            features: _features
        });
        
        emit SubscriptionTierCreated(tierId, _name, _price, _duration);
    }
    
    /**
     * @notice Purchases a subscription
     * @param _tierId The tier ID to subscribe to
     * @return subscriptionId The unique identifier for the subscription
     */
    function purchaseSubscription(uint256 _tierId) external nonReentrant returns (uint256 subscriptionId) {
        if (_tierId == 0 || _tierId > totalTiers) {
            revert TierNotFound();
        }
        
        SubscriptionTier storage tier = subscriptionTiers[_tierId];
        if (!tier.isActive) {
            revert TierNotActive();
        }
        
        if (userSubscriptions[msg.sender].isActive) {
            revert UserAlreadySubscribed();
        }
        
        if (kaiaToken.balanceOf(msg.sender) < tier.price) {
            revert InsufficientTokens();
        }
        
        // Transfer KAIA tokens from user to contract
        require(kaiaToken.transferFrom(msg.sender, address(this), tier.price), "Token transfer failed");
        
        subscriptionId = totalSubscriptions + 1;
        totalSubscriptions = subscriptionId;
        totalRevenue += tier.price;
        
        uint256 startTime = block.timestamp;
        uint256 endTime = startTime + tier.duration;
        
        UserSubscription memory subscription = UserSubscription({
            subscriptionId: subscriptionId,
            user: msg.sender,
            tierId: _tierId,
            startTime: startTime,
            endTime: endTime,
            isActive: true,
            amountPaid: tier.price
        });
        
        userSubscriptions[msg.sender] = subscription;
        subscriptions[subscriptionId] = subscription;
        
        emit SubscriptionPurchased(subscriptionId, msg.sender, _tierId, tier.price, startTime, endTime);
    }
    
    /**
     * @notice Renews an existing subscription
     * @param _subscriptionId The subscription ID to renew
     */
    function renewSubscription(uint256 _subscriptionId) external nonReentrant {
        if (_subscriptionId == 0 || _subscriptionId > totalSubscriptions) {
            revert SubscriptionNotFound();
        }
        
        UserSubscription storage subscription = subscriptions[_subscriptionId];
        if (subscription.user != msg.sender) {
            revert NotAuthorized();
        }
        
        if (!subscription.isActive) {
            revert SubscriptionNotActive();
        }
        
        SubscriptionTier storage tier = subscriptionTiers[subscription.tierId];
        if (!tier.isActive) {
            revert TierNotActive();
        }
        
        if (kaiaToken.balanceOf(msg.sender) < tier.price) {
            revert InsufficientTokens();
        }
        
        // Transfer KAIA tokens from user to contract
        require(kaiaToken.transferFrom(msg.sender, address(this), tier.price), "Token transfer failed");
        
        subscription.endTime += tier.duration;
        subscription.amountPaid += tier.price;
        totalRevenue += tier.price;
        
        emit SubscriptionRenewed(_subscriptionId, msg.sender, subscription.endTime, tier.price);
    }
    
    /**
     * @notice Cancels an active subscription with refund
     * @param _subscriptionId The subscription ID to cancel
     */
    function cancelSubscription(uint256 _subscriptionId) external nonReentrant {
        if (_subscriptionId == 0 || _subscriptionId > totalSubscriptions) {
            revert SubscriptionNotFound();
        }
        
        UserSubscription storage subscription = subscriptions[_subscriptionId];
        if (subscription.user != msg.sender) {
            revert NotAuthorized();
        }
        
        if (!subscription.isActive) {
            revert SubscriptionNotActive();
        }
        
        // Calculate refund based on remaining time
        uint256 remainingTime = subscription.endTime > block.timestamp ? 
            subscription.endTime - block.timestamp : 0;
        
        uint256 totalDuration = subscription.endTime - subscription.startTime;
        uint256 refundAmount = 0;
        
        if (remainingTime > 0 && totalDuration > 0) {
            refundAmount = (subscription.amountPaid * remainingTime) / totalDuration;
        }
        
        subscription.isActive = false;
        userSubscriptions[msg.sender].isActive = false;
        
        if (refundAmount > 0) {
            totalRevenue -= refundAmount;
            require(kaiaToken.transfer(msg.sender, refundAmount), "Refund transfer failed");
        }
        
        emit SubscriptionCancelled(_subscriptionId, msg.sender, refundAmount);
    }
    
    /**
     * @notice Checks if a user has an active subscription
     * @param _user The user address to check
     * @return hasActiveSubscription True if user has active subscription
     * @return tierId The tier ID of the active subscription
     * @return endTime The end time of the subscription
     */
    function getUserSubscriptionStatus(address _user) external view returns (
        bool hasActiveSubscription,
        uint256 tierId,
        uint256 endTime
    ) {
        UserSubscription storage subscription = userSubscriptions[_user];
        hasActiveSubscription = subscription.isActive && subscription.endTime > block.timestamp;
        tierId = subscription.tierId;
        endTime = subscription.endTime;
    }
    
    /**
     * @notice Gets subscription details by ID
     * @param _subscriptionId The subscription ID to query
     * @return subscription The complete subscription structure
     */
    function getSubscription(uint256 _subscriptionId) external view returns (UserSubscription memory subscription) {
        if (_subscriptionId == 0 || _subscriptionId > totalSubscriptions) {
            revert SubscriptionNotFound();
        }
        return subscriptions[_subscriptionId];
    }
    
    /**
     * @notice Gets tier details by ID
     * @param _tierId The tier ID to query
     * @return tier The complete tier structure
     */
    function getSubscriptionTier(uint256 _tierId) external view returns (SubscriptionTier memory tier) {
        if (_tierId == 0 || _tierId > totalTiers) {
            revert TierNotFound();
        }
        return subscriptionTiers[_tierId];
    }
    
    /**
     * @notice Updates a subscription tier
     * @param _tierId The tier ID to update
     * @param _name The new tier name
     * @param _price The new tier price
     * @param _duration The new subscription duration
     * @param _features The new features array
     * @dev Only the contract owner can update tiers
     */
    function updateSubscriptionTier(
        uint256 _tierId,
        string memory _name,
        uint256 _price,
        uint256 _duration,
        string[] memory _features
    ) external onlyOwner {
        if (_tierId == 0 || _tierId > totalTiers) {
            revert TierNotFound();
        }
        
        subscriptionTiers[_tierId].name = _name;
        subscriptionTiers[_tierId].price = _price;
        subscriptionTiers[_tierId].duration = _duration;
        subscriptionTiers[_tierId].features = _features;
    }
    
    /**
     * @notice Activates or deactivates a subscription tier
     * @param _tierId The tier ID to toggle
     * @param _isActive The new active status
     * @dev Only the contract owner can toggle tier status
     */
    function toggleTierStatus(uint256 _tierId, bool _isActive) external onlyOwner {
        if (_tierId == 0 || _tierId > totalTiers) {
            revert TierNotFound();
        }
        
        subscriptionTiers[_tierId].isActive = _isActive;
    }
    
    /**
     * @notice Withdraws accumulated KAIA tokens
     * @dev Only the contract owner can withdraw tokens
     */
    function withdrawTokens() external onlyOwner {
        uint256 balance = kaiaToken.balanceOf(address(this));
        require(balance > 0, "No tokens to withdraw");
        
        require(kaiaToken.transfer(owner(), balance), "Token withdrawal failed");
    }
    
    /**
     * @notice Gets subscription statistics
     * @return _totalTiers Total number of subscription tiers
     * @return _totalSubscriptions Total number of subscriptions
     * @return _totalRevenue Total revenue in KAIA tokens
     * @return _activeSubscriptions Number of active subscriptions
     */
    function getSubscriptionStatistics() external view returns (
        uint256 _totalTiers,
        uint256 _totalSubscriptions,
        uint256 _totalRevenue,
        uint256 _activeSubscriptions
    ) {
        _totalTiers = totalTiers;
        _totalSubscriptions = totalSubscriptions;
        _totalRevenue = totalRevenue;
        
        for (uint256 i = 1; i <= totalSubscriptions; i++) {
            if (subscriptions[i].isActive && subscriptions[i].endTime > block.timestamp) {
                _activeSubscriptions++;
            }
        }
    }
}