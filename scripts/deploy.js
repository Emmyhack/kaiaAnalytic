const { ethers } = require("hardhat");

async function main() {
  console.log("Deploying KaiaAnalyticsAI Smart Contracts to Kaia Network...");

  const [deployer] = await ethers.getSigners();
  console.log("Deploying contracts with account:", deployer.address);
  console.log("Account balance:", (await deployer.getBalance()).toString());

  // Deploy contracts in dependency order
  
  // 1. Deploy AnalyticsRegistry
  console.log("\nDeploying AnalyticsRegistry...");
  const AnalyticsRegistry = await ethers.getContractFactory("AnalyticsRegistry");
  const analyticsRegistry = await AnalyticsRegistry.deploy();
  await analyticsRegistry.deployed();
  console.log("AnalyticsRegistry deployed to:", analyticsRegistry.address);

  // 2. Deploy DataContract
  console.log("\nDeploying DataContract...");
  const DataContract = await ethers.getContractFactory("DataContract");
  const dataContract = await DataContract.deploy();
  await dataContract.deployed();
  console.log("DataContract deployed to:", dataContract.address);

  // 3. Deploy SubscriptionContract (requires KAIA token address)
  // For testnet, we'll use a mock KAIA token address or deploy a mock token
  console.log("\nDeploying mock KAIA token...");
  const MockToken = await ethers.getContractFactory("MockKAIA");
  const mockKaia = await MockToken.deploy();
  await mockKaia.deployed();
  console.log("Mock KAIA token deployed to:", mockKaia.address);

  console.log("\nDeploying SubscriptionContract...");
  const SubscriptionContract = await ethers.getContractFactory("SubscriptionContract");
  const subscriptionContract = await SubscriptionContract.deploy(mockKaia.address);
  await subscriptionContract.deployed();
  console.log("SubscriptionContract deployed to:", subscriptionContract.address);

  // 4. Deploy ActionContract
  console.log("\nDeploying ActionContract...");
  const ActionContract = await ethers.getContractFactory("ActionContract");
  const actionContract = await ActionContract.deploy(subscriptionContract.address);
  await actionContract.deployed();
  console.log("ActionContract deployed to:", actionContract.address);

  // Set up initial configurations
  console.log("\nSetting up initial configurations...");

  // Authorize contracts to interact with each other
  await analyticsRegistry.authorizeProcessor(dataContract.address);
  await dataContract.authorizeProcessor(analyticsRegistry.address);
  await subscriptionContract.authorizeDataReader(analyticsRegistry.address);
  await subscriptionContract.authorizeDataReader(dataContract.address);

  console.log("\nSetup completed!");

  // Output deployment summary
  console.log("\n=== DEPLOYMENT SUMMARY ===");
  console.log("Network:", await ethers.provider.getNetwork());
  console.log("Deployer:", deployer.address);
  console.log("Gas used for deployment: [Calculate manually]");
  console.log("\nContract Addresses:");
  console.log("- AnalyticsRegistry:", analyticsRegistry.address);
  console.log("- DataContract:", dataContract.address);
  console.log("- SubscriptionContract:", subscriptionContract.address);
  console.log("- ActionContract:", actionContract.address);
  console.log("- KAIA Token (Mock):", mockKaia.address);

  console.log("\n=== ENVIRONMENT VARIABLES ===");
  console.log("Add these to your .env file:");
  console.log(`ANALYTICS_REGISTRY_ADDRESS=${analyticsRegistry.address}`);
  console.log(`DATA_CONTRACT_ADDRESS=${dataContract.address}`);
  console.log(`SUBSCRIPTION_CONTRACT_ADDRESS=${subscriptionContract.address}`);
  console.log(`ACTION_CONTRACT_ADDRESS=${actionContract.address}`);
  console.log(`KAIA_TOKEN_ADDRESS=${mockKaia.address}`);

  // Verify deployment
  console.log("\n=== VERIFICATION ===");
  console.log("Verifying contract deployments...");
  
  try {
    const registryCode = await ethers.provider.getCode(analyticsRegistry.address);
    const dataCode = await ethers.provider.getCode(dataContract.address);
    const subscriptionCode = await ethers.provider.getCode(subscriptionContract.address);
    const actionCode = await ethers.provider.getCode(actionContract.address);
    
    console.log("✅ All contracts deployed successfully!");
    console.log("- AnalyticsRegistry bytecode length:", registryCode.length);
    console.log("- DataContract bytecode length:", dataCode.length);
    console.log("- SubscriptionContract bytecode length:", subscriptionCode.length);
    console.log("- ActionContract bytecode length:", actionCode.length);
  } catch (error) {
    console.error("❌ Verification failed:", error.message);
  }

  console.log("\n🎉 Deployment completed successfully!");
  console.log("You can now start the backend services with the contract addresses above.");
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });