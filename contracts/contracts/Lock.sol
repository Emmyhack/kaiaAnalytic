// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

/**
 * @title Lock
 * @dev A time-locked Ether storage contract with enhanced security features
 * @notice This contract allows the owner to lock Ether for a specified period
 * @author Kaia Analytics AI Team
 */
contract Lock {
    /// @notice The timestamp when funds can be withdrawn
    uint256 public unlockTime;
    
    /// @notice The owner of the locked funds
    address payable public owner;
    
    /// @notice Whether the contract has been withdrawn from
    bool public withdrawn;
    
    /// @notice Emitted when funds are withdrawn
    /// @param amount The amount withdrawn in Wei
    /// @param when The timestamp of withdrawal
    /// @param recipient The address that received the funds
    event Withdrawal(uint256 amount, uint256 when, address recipient);
    
    /// @notice Emitted when funds are deposited
    /// @param amount The amount deposited in Wei
    /// @param when The timestamp of deposit
    /// @param depositor The address that deposited the funds
    event Deposit(uint256 amount, uint256 when, address depositor);
    
    /// @notice Emitted when ownership is transferred
    /// @param previousOwner The previous owner address
    /// @param newOwner The new owner address
    event OwnershipTransferred(address indexed previousOwner, address indexed newOwner);

    /// @dev Thrown when unlock time is not in the future
    error UnlockTimeNotInFuture();
    
    /// @dev Thrown when trying to withdraw before unlock time
    error WithdrawalTooEarly();
    
    /// @dev Thrown when caller is not the owner
    error NotOwner();
    
    /// @dev Thrown when funds have already been withdrawn
    error AlreadyWithdrawn();
    
    /// @dev Thrown when no funds are available to withdraw
    error NoFundsToWithdraw();
    
    /// @dev Thrown when transfer to new owner is the same as current owner
    error SameOwner();
    
    /// @dev Thrown when trying to transfer ownership to zero address
    error InvalidNewOwner();

    /// @notice Modifier to check if caller is the owner
    modifier onlyOwner() {
        if (msg.sender != owner) revert NotOwner();
        _;
    }
    
    /// @notice Modifier to check if funds haven't been withdrawn yet
    modifier notWithdrawn() {
        if (withdrawn) revert AlreadyWithdrawn();
        _;
    }

    /**
     * @notice Creates a new Lock contract
     * @param _unlockTime The timestamp when funds can be withdrawn (must be in the future)
     * @dev The contract must be deployed with some Ether to be useful
     */
    constructor(uint256 _unlockTime) payable {
        if (block.timestamp >= _unlockTime) {
            revert UnlockTimeNotInFuture();
        }

        unlockTime = _unlockTime;
        owner = payable(msg.sender);
        withdrawn = false;
        
        if (msg.value > 0) {
            emit Deposit(msg.value, block.timestamp, msg.sender);
        }
    }

    /**
     * @notice Withdraws all locked funds to the owner
     * @dev Can only be called by the owner after the unlock time
     */
    function withdraw() external onlyOwner notWithdrawn {
        if (block.timestamp < unlockTime) {
            revert WithdrawalTooEarly();
        }
        
        uint256 amount = address(this).balance;
        if (amount == 0) {
            revert NoFundsToWithdraw();
        }
        
        withdrawn = true;
        
        emit Withdrawal(amount, block.timestamp, owner);
        
        // Using call instead of transfer for better gas handling
        (bool success, ) = owner.call{value: amount}("");
        require(success, "Transfer failed");
    }
    
    /**
     * @notice Allows additional deposits to the contract
     * @dev Anyone can deposit, but only the owner can withdraw
     */
    function deposit() external payable {
        require(msg.value > 0, "Must send some Ether");
        emit Deposit(msg.value, block.timestamp, msg.sender);
    }
    
    /**
     * @notice Transfers ownership to a new address
     * @param newOwner The address of the new owner
     * @dev Can only be called by the current owner
     */
    function transferOwnership(address payable newOwner) external onlyOwner {
        if (newOwner == address(0)) {
            revert InvalidNewOwner();
        }
        if (newOwner == owner) {
            revert SameOwner();
        }
        
        address previousOwner = owner;
        owner = newOwner;
        
        emit OwnershipTransferred(previousOwner, newOwner);
    }
    
    /**
     * @notice Gets the current balance of the contract
     * @return The balance in Wei
     */
    function getBalance() external view returns (uint256) {
        return address(this).balance;
    }
    
    /**
     * @notice Gets the time remaining until unlock
     * @return The number of seconds until unlock, or 0 if already unlocked
     */
    function getTimeUntilUnlock() external view returns (uint256) {
        if (block.timestamp >= unlockTime) {
            return 0;
        }
        return unlockTime - block.timestamp;
    }
    
    /**
     * @notice Checks if the contract is currently locked
     * @return True if funds are still locked, false otherwise
     */
    function isLocked() external view returns (bool) {
        return block.timestamp < unlockTime && !withdrawn;
    }
    
    /**
     * @notice Fallback function to receive Ether
     * @dev Emits a Deposit event when Ether is received
     */
    receive() external payable {
        if (msg.value > 0) {
            emit Deposit(msg.value, block.timestamp, msg.sender);
        }
    }
}
