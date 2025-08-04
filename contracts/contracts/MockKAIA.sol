// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

/**
 * @title MockKAIA
 * @dev Mock KAIA token for testing purposes
 */
contract MockKAIA is ERC20, Ownable {
    uint8 private _decimals = 18;
    
    constructor() ERC20("Mock KAIA", "KAIA") {
        // Mint initial supply to deployer
        _mint(msg.sender, 5000000000 * 10**decimals()); // 5 billion tokens
    }
    
    /**
     * @dev Mint tokens to specified address (for testing)
     */
    function mint(address to, uint256 amount) external onlyOwner {
        _mint(to, amount);
    }
    
    /**
     * @dev Faucet function for testing - anyone can get 1000 KAIA
     */
    function faucet() external {
        require(balanceOf(msg.sender) < 10000 * 10**decimals(), "Already has enough tokens");
        _mint(msg.sender, 1000 * 10**decimals());
    }
    
    /**
     * @dev Returns the number of decimals
     */
    function decimals() public view virtual override returns (uint8) {
        return _decimals;
    }
}