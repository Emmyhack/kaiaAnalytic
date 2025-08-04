const {
  time,
  loadFixture,
} = require("@nomicfoundation/hardhat-toolbox/network-helpers");
const { anyValue } = require("@nomicfoundation/hardhat-chai-matchers/withArgs");
const { expect } = require("chai");

describe("Lock", function () {
  // We define a fixture to reuse the same setup in every test.
  // We use loadFixture to run this setup once, snapshot that state,
  // and reset Hardhat Network to that snapshot in every test.
  async function deployOneYearLockFixture() {
    const ONE_YEAR_IN_SECS = 365 * 24 * 60 * 60;
    const ONE_GWEI = 1_000_000_000;

    const lockedAmount = ONE_GWEI;
    const unlockTime = (await time.latest()) + ONE_YEAR_IN_SECS;

    // Contracts are deployed using the first signer/account by default
    const [owner, otherAccount, thirdAccount] = await ethers.getSigners();

    const Lock = await ethers.getContractFactory("Lock");
    const lock = await Lock.deploy(unlockTime, { value: lockedAmount });

    return { lock, unlockTime, lockedAmount, owner, otherAccount, thirdAccount };
  }

  async function deployEmptyLockFixture() {
    const ONE_YEAR_IN_SECS = 365 * 24 * 60 * 60;
    const unlockTime = (await time.latest()) + ONE_YEAR_IN_SECS;

    const [owner, otherAccount] = await ethers.getSigners();

    const Lock = await ethers.getContractFactory("Lock");
    const lock = await Lock.deploy(unlockTime, { value: 0 });

    return { lock, unlockTime, owner, otherAccount };
  }

  describe("Deployment", function () {
    it("Should set the right unlockTime", async function () {
      const { lock, unlockTime } = await loadFixture(deployOneYearLockFixture);

      expect(await lock.unlockTime()).to.equal(unlockTime);
    });

    it("Should set the right owner", async function () {
      const { lock, owner } = await loadFixture(deployOneYearLockFixture);

      expect(await lock.owner()).to.equal(owner.address);
    });

    it("Should receive and store the funds to lock", async function () {
      const { lock, lockedAmount } = await loadFixture(deployOneYearLockFixture);

      expect(await ethers.provider.getBalance(lock.target)).to.equal(lockedAmount);
    });

    it("Should set withdrawn to false initially", async function () {
      const { lock } = await loadFixture(deployOneYearLockFixture);

      expect(await lock.withdrawn()).to.equal(false);
    });

    it("Should fail if the unlockTime is not in the future", async function () {
      const latestTime = await time.latest();
      const Lock = await ethers.getContractFactory("Lock");
      
      await expect(Lock.deploy(latestTime, { value: 1 }))
        .to.be.revertedWithCustomError(Lock, "UnlockTimeNotInFuture");
    });

    it("Should emit Deposit event on deployment with funds", async function () {
      const ONE_YEAR_IN_SECS = 365 * 24 * 60 * 60;
      const unlockTime = (await time.latest()) + ONE_YEAR_IN_SECS;
      const lockedAmount = 1000;
      const [deployer] = await ethers.getSigners();

      const Lock = await ethers.getContractFactory("Lock");
      const lock = await Lock.deploy(unlockTime, { value: lockedAmount });
      
      // Check the deployment transaction receipt for the Deposit event
      const receipt = await lock.deploymentTransaction().wait();
      const depositEvent = receipt.logs.find(log => {
        try {
          const parsed = lock.interface.parseLog(log);
          return parsed.name === "Deposit";
        } catch {
          return false;
        }
      });
      
      expect(depositEvent).to.not.be.undefined;
      const parsedEvent = lock.interface.parseLog(depositEvent);
      expect(parsedEvent.args.amount).to.equal(lockedAmount);
      expect(parsedEvent.args.depositor).to.equal(deployer.address);
    });
  });

  describe("Withdrawals", function () {
    describe("Validations", function () {
      it("Should revert with the right error if called too soon", async function () {
        const { lock } = await loadFixture(deployOneYearLockFixture);

        await expect(lock.withdraw())
          .to.be.revertedWithCustomError(lock, "WithdrawalTooEarly");
      });

      it("Should revert with the right error if called from another account", async function () {
        const { lock, unlockTime, otherAccount } = await loadFixture(deployOneYearLockFixture);

        await time.increaseTo(unlockTime);

        await expect(lock.connect(otherAccount).withdraw())
          .to.be.revertedWithCustomError(lock, "NotOwner");
      });

      it("Should revert if already withdrawn", async function () {
        const { lock, unlockTime } = await loadFixture(deployOneYearLockFixture);

        await time.increaseTo(unlockTime);
        await lock.withdraw();

        await expect(lock.withdraw())
          .to.be.revertedWithCustomError(lock, "AlreadyWithdrawn");
      });

      it("Should revert if no funds to withdraw", async function () {
        const { lock, unlockTime } = await loadFixture(deployEmptyLockFixture);

        await time.increaseTo(unlockTime);

        await expect(lock.withdraw())
          .to.be.revertedWithCustomError(lock, "NoFundsToWithdraw");
      });

      it("Shouldn't fail if the unlockTime has arrived and the owner calls it", async function () {
        const { lock, unlockTime } = await loadFixture(deployOneYearLockFixture);

        await time.increaseTo(unlockTime);

        await expect(lock.withdraw()).not.to.be.reverted;
      });
    });

    describe("Events", function () {
      it("Should emit an event on withdrawals", async function () {
        const { lock, unlockTime, lockedAmount, owner } = await loadFixture(deployOneYearLockFixture);

        await time.increaseTo(unlockTime);

        await expect(lock.withdraw())
          .to.emit(lock, "Withdrawal")
          .withArgs(lockedAmount, anyValue, owner.address);
      });
    });

    describe("Transfers", function () {
      it("Should transfer the funds to the owner", async function () {
        const { lock, unlockTime, lockedAmount, owner } = await loadFixture(deployOneYearLockFixture);

        await time.increaseTo(unlockTime);

        await expect(lock.withdraw()).to.changeEtherBalances(
          [owner, lock],
          [lockedAmount, -lockedAmount]
        );
      });

      it("Should set withdrawn to true after withdrawal", async function () {
        const { lock, unlockTime } = await loadFixture(deployOneYearLockFixture);

        await time.increaseTo(unlockTime);
        await lock.withdraw();

        expect(await lock.withdrawn()).to.equal(true);
      });
    });
  });

  describe("Deposits", function () {
    it("Should allow deposits from anyone", async function () {
      const { lock, otherAccount } = await loadFixture(deployOneYearLockFixture);
      const depositAmount = 500;

      await expect(lock.connect(otherAccount).deposit({ value: depositAmount }))
        .to.emit(lock, "Deposit")
        .withArgs(depositAmount, anyValue, otherAccount.address);
    });

    it("Should increase contract balance on deposit", async function () {
      const { lock, lockedAmount } = await loadFixture(deployOneYearLockFixture);
      const depositAmount = 500;

      await lock.deposit({ value: depositAmount });

      expect(await ethers.provider.getBalance(lock.target))
        .to.equal(lockedAmount + depositAmount);
    });

    it("Should revert on zero value deposit", async function () {
      const { lock } = await loadFixture(deployOneYearLockFixture);

      await expect(lock.deposit({ value: 0 }))
        .to.be.revertedWith("Must send some Ether");
    });
  });

  describe("Ownership Transfer", function () {
    it("Should transfer ownership to new address", async function () {
      const { lock, otherAccount } = await loadFixture(deployOneYearLockFixture);

      await expect(lock.transferOwnership(otherAccount.address))
        .to.emit(lock, "OwnershipTransferred")
        .withArgs(anyValue, otherAccount.address);

      expect(await lock.owner()).to.equal(otherAccount.address);
    });

    it("Should revert when transferring to zero address", async function () {
      const { lock } = await loadFixture(deployOneYearLockFixture);

      await expect(lock.transferOwnership(ethers.ZeroAddress))
        .to.be.revertedWithCustomError(lock, "InvalidNewOwner");
    });

    it("Should revert when transferring to same owner", async function () {
      const { lock, owner } = await loadFixture(deployOneYearLockFixture);

      await expect(lock.transferOwnership(owner.address))
        .to.be.revertedWithCustomError(lock, "SameOwner");
    });

    it("Should revert when non-owner tries to transfer ownership", async function () {
      const { lock, otherAccount, thirdAccount } = await loadFixture(deployOneYearLockFixture);

      await expect(lock.connect(otherAccount).transferOwnership(thirdAccount.address))
        .to.be.revertedWithCustomError(lock, "NotOwner");
    });

    it("Should allow new owner to withdraw after ownership transfer", async function () {
      const { lock, unlockTime, lockedAmount, otherAccount } = await loadFixture(deployOneYearLockFixture);

      await lock.transferOwnership(otherAccount.address);
      await time.increaseTo(unlockTime);

      await expect(lock.connect(otherAccount).withdraw())
        .to.changeEtherBalances([otherAccount, lock], [lockedAmount, -lockedAmount]);
    });
  });

  describe("View Functions", function () {
    it("Should return correct balance", async function () {
      const { lock, lockedAmount } = await loadFixture(deployOneYearLockFixture);

      expect(await lock.getBalance()).to.equal(lockedAmount);
    });

    it("Should return correct time until unlock", async function () {
      const { lock, unlockTime } = await loadFixture(deployOneYearLockFixture);
      const currentTime = await time.latest();

      expect(await lock.getTimeUntilUnlock()).to.equal(unlockTime - currentTime);
    });

    it("Should return zero when unlock time has passed", async function () {
      const { lock, unlockTime } = await loadFixture(deployOneYearLockFixture);

      await time.increaseTo(unlockTime + 1);

      expect(await lock.getTimeUntilUnlock()).to.equal(0);
    });

    it("Should return correct lock status", async function () {
      const { lock, unlockTime } = await loadFixture(deployOneYearLockFixture);

      expect(await lock.isLocked()).to.equal(true);

      await time.increaseTo(unlockTime);

      expect(await lock.isLocked()).to.equal(false);
    });

    it("Should return false for isLocked after withdrawal", async function () {
      const { lock, unlockTime } = await loadFixture(deployOneYearLockFixture);

      await time.increaseTo(unlockTime);
      await lock.withdraw();

      expect(await lock.isLocked()).to.equal(false);
    });
  });

  describe("Receive Function", function () {
    it("Should accept Ether sent directly to contract", async function () {
      const { lock, lockedAmount, otherAccount } = await loadFixture(deployOneYearLockFixture);
      const sendAmount = 1000;

      await expect(
        otherAccount.sendTransaction({
          to: lock.target,
          value: sendAmount,
        })
      ).to.emit(lock, "Deposit")
        .withArgs(sendAmount, anyValue, otherAccount.address);

      expect(await ethers.provider.getBalance(lock.target))
        .to.equal(lockedAmount + sendAmount);
    });
  });

  describe("Gas Optimization", function () {
    it("Should use reasonable gas for deployment", async function () {
      const ONE_YEAR_IN_SECS = 365 * 24 * 60 * 60;
      const unlockTime = (await time.latest()) + ONE_YEAR_IN_SECS;

      const Lock = await ethers.getContractFactory("Lock");
      const deployTx = await Lock.getDeployTransaction(unlockTime, { value: 1000 });
      
      // Estimate gas should be reasonable (less than 500k gas)
      const gasEstimate = await ethers.provider.estimateGas(deployTx);
      expect(gasEstimate).to.be.lessThan(500000);
    });

    it("Should use reasonable gas for withdrawal", async function () {
      const { lock, unlockTime } = await loadFixture(deployOneYearLockFixture);

      await time.increaseTo(unlockTime);

      const gasEstimate = await lock.withdraw.estimateGas();
      expect(gasEstimate).to.be.lessThan(100000);
    });
  });
});
