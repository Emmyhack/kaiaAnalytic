const { expect } = require("chai");
const { ethers } = require("hardhat");

describe("KaiaAnalyticsAI Smart Contracts", function () {
  let analyticsRegistry, dataContract, subscriptionContract, actionContract;
  let owner, user1, user2, executor;
  let mockKaiaToken;

  beforeEach(async function () {
    [owner, user1, user2, executor] = await ethers.getSigners();

    // Deploy mock KAIA token for testing
    const MockToken = await ethers.getContractFactory("MockERC20");
    mockKaiaToken = await MockToken.deploy("Mock KAIA", "KAIA");
    await mockKaiaToken.deployed();

    // Deploy contracts
    const AnalyticsRegistry = await ethers.getContractFactory("AnalyticsRegistry");
    analyticsRegistry = await AnalyticsRegistry.deploy();
    await analyticsRegistry.deployed();

    const DataContract = await ethers.getContractFactory("DataContract");
    dataContract = await DataContract.deploy();
    await dataContract.deployed();

    const SubscriptionContract = await ethers.getContractFactory("SubscriptionContract");
    subscriptionContract = await SubscriptionContract.deploy(mockKaiaToken.address);
    await subscriptionContract.deployed();

    const ActionContract = await ethers.getContractFactory("ActionContract");
    actionContract = await ActionContract.deploy();
    await actionContract.deployed();

    // Set up initial configuration
    await analyticsRegistry.addAuthorizedSubmitter(dataContract.address);
    await dataContract.addAuthorizedSubmitter(actionContract.address);
    await actionContract.authorizeExecutor(executor.address);
  });

  describe("AnalyticsRegistry", function () {
    it("Should register a new analytics task", async function () {
      const taskType = "yield_analysis";
      const description = "Analyze yield opportunities across protocols";
      const reward = ethers.utils.parseEther("1");

      await expect(analyticsRegistry.connect(user1).registerTask(taskType, description, reward))
        .to.emit(analyticsRegistry, "TaskRegistered")
        .withArgs(1, user1.address, taskType, description, reward);

      const task = await analyticsRegistry.getTask(1);
      expect(task.taskId).to.equal(1);
      expect(task.creator).to.equal(user1.address);
      expect(task.taskType).to.equal(taskType);
      expect(task.isActive).to.be.true;
    });

    it("Should complete a task", async function () {
      await analyticsRegistry.connect(user1).registerTask("test", "test", 100);
      await analyticsRegistry.connect(owner).completeTask(1, executor.address);

      const task = await analyticsRegistry.getTask(1);
      expect(task.isActive).to.be.false;
      expect(task.executor).to.equal(executor.address);
    });

    it("Should cancel a task by creator", async function () {
      await analyticsRegistry.connect(user1).registerTask("test", "test", 100);
      await analyticsRegistry.connect(user1).cancelTask(1);

      const task = await analyticsRegistry.getTask(1);
      expect(task.isActive).to.be.false;
    });

    it("Should not allow non-creator to cancel task", async function () {
      await analyticsRegistry.connect(user1).registerTask("test", "test", 100);
      await expect(analyticsRegistry.connect(user2).cancelTask(1))
        .to.be.revertedWith("Only task creator can perform this action");
    });
  });

  describe("DataContract", function () {
    it("Should store analytics result", async function () {
      const taskId = 1;
      const dataType = "yield_analysis";
      const dataHash = ethers.utils.keccak256(ethers.utils.toUtf8Bytes("test data"));

      await expect(dataContract.connect(owner).storeAnalyticsResult(taskId, dataType, dataHash))
        .to.emit(dataContract, "AnalyticsResultStored")
        .withArgs(1, taskId, dataType, dataHash);

      const result = await dataContract.getAnalyticsResult(1);
      expect(result.resultId).to.equal(1);
      expect(result.dataType).to.equal(dataType);
      expect(result.dataHash).to.equal(dataHash);
    });

    it("Should store trade history", async function () {
      const user = user1.address;
      const tokenPair = "KAIA/USDC";
      const amount = ethers.utils.parseEther("100");
      const price = ethers.utils.parseEther("1.5");
      const tradeType = "buy";

      await expect(dataContract.connect(owner).storeTradeHistory(user, tokenPair, amount, price, tradeType))
        .to.emit(dataContract, "TradeHistoryStored");

      const trade = await dataContract.getTradeHistory(1);
      expect(trade.tradeId).to.equal(1);
      expect(trade.tokenPair).to.equal(tokenPair);
      expect(trade.amount).to.equal(amount);
      expect(trade.isAnonymized).to.be.true;
    });

    it("Should validate analytics result", async function () {
      await dataContract.connect(owner).storeAnalyticsResult(1, "test", ethers.utils.keccak256("test"));
      await dataContract.connect(owner).validateAnalyticsResult(1, 85);

      const result = await dataContract.getAnalyticsResult(1);
      expect(result.isValidated).to.be.true;
      expect(result.validationScore).to.equal(85);
    });
  });

  describe("SubscriptionContract", function () {
    beforeEach(async function () {
      // Create subscription plans
      const basicFeatures = ["Basic analytics", "Transaction tracking"];
      const premiumFeatures = ["Advanced analytics", "Chat support", "Yield analysis"];

      await subscriptionContract.connect(owner).createPlan(
        "Basic Plan",
        ethers.utils.parseEther("10"),
        30 * 24 * 60 * 60, // 30 days
        basicFeatures
      );

      await subscriptionContract.connect(owner).createPlan(
        "Premium Plan",
        ethers.utils.parseEther("50"),
        30 * 24 * 60 * 60, // 30 days
        premiumFeatures
      );

      // Mint tokens to users
      await mockKaiaToken.mint(user1.address, ethers.utils.parseEther("100"));
      await mockKaiaToken.mint(user2.address, ethers.utils.parseEther("100"));
    });

    it("Should purchase a subscription", async function () {
      await mockKaiaToken.connect(user1).approve(subscriptionContract.address, ethers.utils.parseEther("10"));

      await expect(subscriptionContract.connect(user1).purchaseSubscription(1))
        .to.emit(subscriptionContract, "SubscriptionPurchased");

      const subscription = await subscriptionContract.getUserActiveSubscription(user1.address);
      expect(subscription.planId).to.equal(1);
      expect(subscription.isActive).to.be.true;
    });

    it("Should check active subscription", async function () {
      await mockKaiaToken.connect(user1).approve(subscriptionContract.address, ethers.utils.parseEther("10"));
      await subscriptionContract.connect(user1).purchaseSubscription(1);

      const hasActive = await subscriptionContract.hasActiveSubscription(user1.address);
      expect(hasActive).to.be.true;
    });

    it("Should renew subscription", async function () {
      await mockKaiaToken.connect(user1).approve(subscriptionContract.address, ethers.utils.parseEther("20"));
      await subscriptionContract.connect(user1).purchaseSubscription(1);
      
      const originalSubscription = await subscriptionContract.getUserActiveSubscription(user1.address);
      await subscriptionContract.connect(user1).renewSubscription(1);
      
      const renewedSubscription = await subscriptionContract.getUserActiveSubscription(user1.address);
      expect(renewedSubscription.endTime).to.be.gt(originalSubscription.endTime);
    });

    it("Should cancel subscription", async function () {
      await mockKaiaToken.connect(user1).approve(subscriptionContract.address, ethers.utils.parseEther("10"));
      await subscriptionContract.connect(user1).purchaseSubscription(1);
      await subscriptionContract.connect(user1).cancelSubscription(1);

      const subscription = await subscriptionContract.getUserActiveSubscription(user1.address);
      expect(subscription.isActive).to.be.false;
    });
  });

  describe("ActionContract", function () {
    it("Should create staking action", async function () {
      const token = mockKaiaToken.address;
      const amount = ethers.utils.parseEther("100");
      const lockPeriod = 30 * 24 * 60 * 60; // 30 days

      await expect(actionContract.connect(user1).createStakingAction(token, amount, lockPeriod))
        .to.emit(actionContract, "ActionCreated")
        .withArgs(1, user1.address, "stake");

      const action = await actionContract.getAction(1);
      expect(action.actionId).to.equal(1);
      expect(action.user).to.equal(user1.address);
      expect(action.actionType).to.equal("stake");
      expect(action.isExecuted).to.be.false;
    });

    it("Should create voting action", async function () {
      const proposalId = 1;
      const support = true;
      const weight = ethers.utils.parseEther("100");

      await expect(actionContract.connect(user1).createVotingAction(proposalId, support, weight))
        .to.emit(actionContract, "ActionCreated")
        .withArgs(1, user1.address, "vote");

      const action = await actionContract.getAction(1);
      expect(action.actionType).to.equal("vote");
    });

    it("Should create swap action", async function () {
      const tokenIn = mockKaiaToken.address;
      const tokenOut = "0x1234567890123456789012345678901234567890";
      const amountIn = ethers.utils.parseEther("100");
      const minAmountOut = ethers.utils.parseEther("95");

      await expect(actionContract.connect(user1).createSwapAction(tokenIn, tokenOut, amountIn, minAmountOut))
        .to.emit(actionContract, "ActionCreated")
        .withArgs(1, user1.address, "swap");

      const action = await actionContract.getAction(1);
      expect(action.actionType).to.equal("swap");
    });

    it("Should execute action", async function () {
      await actionContract.connect(user1).createStakingAction(mockKaiaToken.address, 100, 3600);
      
      // Mock token approval for execution
      await mockKaiaToken.mint(user1.address, 100);
      await mockKaiaToken.connect(user1).approve(actionContract.address, 100);

      await expect(actionContract.connect(executor).executeAction(1))
        .to.emit(actionContract, "ActionExecuted");

      const action = await actionContract.getAction(1);
      expect(action.isExecuted).to.be.true;
    });

    it("Should not allow unauthorized execution", async function () {
      await actionContract.connect(user1).createStakingAction(mockKaiaToken.address, 100, 3600);
      
      await expect(actionContract.connect(user2).executeAction(1))
        .to.be.revertedWith("Not authorized executor");
    });
  });

  describe("Integration Tests", function () {
    it("Should integrate all contracts correctly", async function () {
      // 1. Register analytics task
      await analyticsRegistry.connect(user1).registerTask("yield_analysis", "test", 100);
      
      // 2. Store analytics result
      const dataHash = ethers.utils.keccak256("test data");
      await dataContract.connect(owner).storeAnalyticsResult(1, "yield_analysis", dataHash);
      
      // 3. Purchase subscription
      await mockKaiaToken.mint(user1.address, ethers.utils.parseEther("100"));
      await mockKaiaToken.connect(user1).approve(subscriptionContract.address, ethers.utils.parseEther("10"));
      await subscriptionContract.connect(user1).purchaseSubscription(1);
      
      // 4. Create and execute action
      await actionContract.connect(user1).createStakingAction(mockKaiaToken.address, 100, 3600);
      await mockKaiaToken.mint(user1.address, 100);
      await mockKaiaToken.connect(user1).approve(actionContract.address, 100);
      await actionContract.connect(executor).executeAction(1);
      
      // Verify integration
      const hasActive = await subscriptionContract.hasActiveSubscription(user1.address);
      expect(hasActive).to.be.true;
      
      const action = await actionContract.getAction(1);
      expect(action.isExecuted).to.be.true;
    });
  });
});

// Mock ERC20 token for testing
contract("MockERC20", function () {
  it("Should deploy mock token", async function () {
    const MockToken = await ethers.getContractFactory("MockERC20");
    const mockToken = await MockToken.deploy("Mock KAIA", "KAIA");
    await mockToken.deployed();
    
    expect(await mockToken.name()).to.equal("Mock KAIA");
    expect(await mockToken.symbol()).to.equal("KAIA");
  });
});