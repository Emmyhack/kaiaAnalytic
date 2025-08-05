const { expect } = require("chai");
const { ethers } = require("hardhat");

describe("AnalyticsRegistry", function () {
  let analyticsRegistry;
  let owner;
  let user1;
  let user2;

  beforeEach(async function () {
    [owner, user1, user2] = await ethers.getSigners();
    
    const AnalyticsRegistry = await ethers.getContractFactory("AnalyticsRegistry");
    analyticsRegistry = await AnalyticsRegistry.deploy(ethers.parseEther("0.001"));
    await analyticsRegistry.waitForDeployment();
  });

  describe("Deployment", function () {
    it("Should set the correct registration fee", async function () {
      expect(await analyticsRegistry.registrationFee()).to.equal(ethers.parseEther("0.001"));
    });

    it("Should set the correct owner", async function () {
      expect(await analyticsRegistry.owner()).to.equal(owner.address);
    });
  });

  describe("Task Registration", function () {
    it("Should register a new task successfully", async function () {
      const taskType = "yield_analysis";
      const parameters = '{"protocol": "uniswap", "asset": "ETH"}';
      const fee = ethers.parseEther("0.001");

      await expect(analyticsRegistry.connect(user1).registerTask(taskType, parameters, { value: fee }))
        .to.emit(analyticsRegistry, "TaskRegistered")
        .withArgs(1, user1.address, taskType, parameters, await ethers.provider.getBlock("latest").then(b => b.timestamp));

      const task = await analyticsRegistry.getTask(1);
      expect(task.taskId).to.equal(1);
      expect(task.requester).to.equal(user1.address);
      expect(task.taskType).to.equal(taskType);
      expect(task.parameters).to.equal(parameters);
      expect(task.isActive).to.be.true;
    });

    it("Should revert when task type is empty", async function () {
      const parameters = '{"protocol": "uniswap"}';
      const fee = ethers.parseEther("0.001");

      await expect(analyticsRegistry.connect(user1).registerTask("", parameters, { value: fee }))
        .to.be.revertedWithCustomError(analyticsRegistry, "EmptyTaskType");
    });

    it("Should revert when parameters are empty", async function () {
      const taskType = "yield_analysis";
      const fee = ethers.parseEther("0.001");

      await expect(analyticsRegistry.connect(user1).registerTask(taskType, "", { value: fee }))
        .to.be.revertedWithCustomError(analyticsRegistry, "EmptyParameters");
    });

    it("Should revert when insufficient fee is provided", async function () {
      const taskType = "yield_analysis";
      const parameters = '{"protocol": "uniswap"}';
      const fee = ethers.parseEther("0.0005"); // Less than required

      await expect(analyticsRegistry.connect(user1).registerTask(taskType, parameters, { value: fee }))
        .to.be.revertedWithCustomError(analyticsRegistry, "InsufficientRegistrationFee");
    });
  });

  describe("Task Completion", function () {
    beforeEach(async function () {
      const taskType = "yield_analysis";
      const parameters = '{"protocol": "uniswap"}';
      const fee = ethers.parseEther("0.001");

      await analyticsRegistry.connect(user1).registerTask(taskType, parameters, { value: fee });
    });

    it("Should complete a task successfully", async function () {
      const resultHash = "0x1234567890abcdef";
      
      await expect(analyticsRegistry.connect(owner).completeTask(1, resultHash))
        .to.emit(analyticsRegistry, "TaskCompleted")
        .withArgs(1, resultHash, await ethers.provider.getBlock("latest").then(b => b.timestamp));

      const task = await analyticsRegistry.getTask(1);
      expect(task.isActive).to.be.false;
      expect(task.resultHash).to.equal(resultHash);
    });

    it("Should revert when task doesn't exist", async function () {
      const resultHash = "0x1234567890abcdef";
      
      await expect(analyticsRegistry.connect(owner).completeTask(999, resultHash))
        .to.be.revertedWithCustomError(analyticsRegistry, "TaskNotFound");
    });

    it("Should revert when task is already completed", async function () {
      const resultHash = "0x1234567890abcdef";
      
      await analyticsRegistry.connect(owner).completeTask(1, resultHash);
      
      await expect(analyticsRegistry.connect(owner).completeTask(1, resultHash))
        .to.be.revertedWithCustomError(analyticsRegistry, "TaskAlreadyCompleted");
    });

    it("Should revert when non-owner tries to complete task", async function () {
      const resultHash = "0x1234567890abcdef";
      
      await expect(analyticsRegistry.connect(user1).completeTask(1, resultHash))
        .to.be.revertedWith("Ownable: caller is not the owner");
    });
  });

  describe("Task Queries", function () {
    beforeEach(async function () {
      const taskType = "yield_analysis";
      const parameters = '{"protocol": "uniswap"}';
      const fee = ethers.parseEther("0.001");

      await analyticsRegistry.connect(user1).registerTask(taskType, parameters, { value: fee });
      await analyticsRegistry.connect(user2).registerTask(taskType, parameters, { value: fee });
    });

    it("Should get active tasks by type", async function () {
      const activeTasks = await analyticsRegistry.getActiveTasksByType("yield_analysis");
      expect(activeTasks.length).to.equal(2);
      expect(activeTasks[0]).to.equal(1);
      expect(activeTasks[1]).to.equal(2);
    });

    it("Should return empty array for non-existent task type", async function () {
      const activeTasks = await analyticsRegistry.getActiveTasksByType("non_existent");
      expect(activeTasks.length).to.equal(0);
    });
  });

  describe("Statistics", function () {
    beforeEach(async function () {
      const taskType = "yield_analysis";
      const parameters = '{"protocol": "uniswap"}';
      const fee = ethers.parseEther("0.001");

      await analyticsRegistry.connect(user1).registerTask(taskType, parameters, { value: fee });
      await analyticsRegistry.connect(user2).registerTask(taskType, parameters, { value: fee });
      
      // Complete one task
      await analyticsRegistry.connect(owner).completeTask(1, "0x1234567890abcdef");
    });

    it("Should return correct statistics", async function () {
      const stats = await analyticsRegistry.getTaskStatistics();
      expect(stats._totalTasks).to.equal(2);
      expect(stats._activeTasks).to.equal(1);
      expect(stats._completedTasks).to.equal(1);
    });
  });

  describe("Fee Management", function () {
    it("Should update registration fee", async function () {
      const newFee = ethers.parseEther("0.002");
      
      await expect(analyticsRegistry.connect(owner).updateRegistrationFee(newFee))
        .to.emit(analyticsRegistry, "RegistrationFeeUpdated")
        .withArgs(ethers.parseEther("0.001"), newFee);

      expect(await analyticsRegistry.registrationFee()).to.equal(newFee);
    });

    it("Should revert when non-owner tries to update fee", async function () {
      const newFee = ethers.parseEther("0.002");
      
      await expect(analyticsRegistry.connect(user1).updateRegistrationFee(newFee))
        .to.be.revertedWith("Ownable: caller is not the owner");
    });

    it("Should withdraw fees", async function () {
      const taskType = "yield_analysis";
      const parameters = '{"protocol": "uniswap"}';
      const fee = ethers.parseEther("0.001");

      await analyticsRegistry.connect(user1).registerTask(taskType, parameters, { value: fee });

      const initialBalance = await ethers.provider.getBalance(owner.address);
      await analyticsRegistry.connect(owner).withdrawFees();
      const finalBalance = await ethers.provider.getBalance(owner.address);

      expect(finalBalance).to.be.gt(initialBalance);
    });
  });
});