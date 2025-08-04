const { ethers } = require("hardhat");

async function main() {
  console.log("Deploying KaiaAnalyticsAI smart contracts...");

  // Get the deployer account
  const [deployer] = await ethers.getSigners();
  console.log("Deploying contracts with account:", deployer.address);
  console.log("Account balance:", (await deployer.getBalance()).toString());

  // Deploy AnalyticsRegistry
  console.log("\nDeploying AnalyticsRegistry...");
  const AnalyticsRegistry = await ethers.getContractFactory("AnalyticsRegistry");
  const analyticsRegistry = await AnalyticsRegistry.deploy();
  await analyticsRegistry.deployed();
  console.log("AnalyticsRegistry deployed to:", analyticsRegistry.address);

  // Deploy DataContract
  console.log("\nDeploying DataContract...");
  const DataContract = await ethers.getContractFactory("DataContract");
  const dataContract = await DataContract.deploy();
  await dataContract.deployed();
  console.log("DataContract deployed to:", dataContract.address);

  // Deploy SubscriptionContract (requires KAIA token address)
  console.log("\nDeploying SubscriptionContract...");
  const SubscriptionContract = await ethers.getContractFactory("SubscriptionContract");
  
  // For testing, we'll use a mock KAIA token address
  // In production, replace with actual KAIA token address
  const MOCK_KAIA_TOKEN = "0x1234567890123456789012345678901234567890";
  const subscriptionContract = await SubscriptionContract.deploy(MOCK_KAIA_TOKEN);
  await subscriptionContract.deployed();
  console.log("SubscriptionContract deployed to:", subscriptionContract.address);

  // Deploy ActionContract
  console.log("\nDeploying ActionContract...");
  const ActionContract = await ethers.getContractFactory("ActionContract");
  const actionContract = await ActionContract.deploy();
  await actionContract.deployed();
  console.log("ActionContract deployed to:", actionContract.address);

  // Set up initial configuration
  console.log("\nSetting up initial configuration...");

  // Add DataContract as authorized submitter to AnalyticsRegistry
  await analyticsRegistry.addAuthorizedSubmitter(dataContract.address);
  console.log("DataContract authorized as submitter in AnalyticsRegistry");

  // Add ActionContract as authorized submitter to DataContract
  await dataContract.addAuthorizedSubmitter(actionContract.address);
  console.log("ActionContract authorized as submitter in DataContract");

  // Create some initial subscription plans
  console.log("\nCreating initial subscription plans...");
  
  const basicPlanFeatures = [
    "Basic analytics dashboard",
    "Transaction volume tracking",
    "Gas price monitoring"
  ];
  
  const premiumPlanFeatures = [
    "Advanced yield analysis",
    "Personalized trading suggestions",
    "Real-time chat support",
    "Governance sentiment tracking",
    "Priority data access"
  ];

  // Basic plan: 10 KAIA tokens for 30 days
  await subscriptionContract.createPlan(
    "Basic Plan",
    ethers.utils.parseEther("10"),
    30 * 24 * 60 * 60, // 30 days in seconds
    basicPlanFeatures
  );
  console.log("Basic subscription plan created");

  // Premium plan: 50 KAIA tokens for 30 days
  await subscriptionContract.createPlan(
    "Premium Plan",
    ethers.utils.parseEther("50"),
    30 * 24 * 60 * 60, // 30 days in seconds
    premiumPlanFeatures
  );
  console.log("Premium subscription plan created");

  // Authorize deployer as executor in ActionContract
  await actionContract.authorizeExecutor(deployer.address);
  console.log("Deployer authorized as executor in ActionContract");

  console.log("\n=== Deployment Summary ===");
  console.log("AnalyticsRegistry:", analyticsRegistry.address);
  console.log("DataContract:", dataContract.address);
  console.log("SubscriptionContract:", subscriptionContract.address);
  console.log("ActionContract:", actionContract.address);
  console.log("Deployer:", deployer.address);
  console.log("========================\n");

  // Save deployment addresses for backend integration
  const deploymentInfo = {
    network: network.name,
    deployer: deployer.address,
    contracts: {
      analyticsRegistry: analyticsRegistry.address,
      dataContract: dataContract.address,
      subscriptionContract: subscriptionContract.address,
      actionContract: actionContract.address
    },
    timestamp: new Date().toISOString()
  };

  console.log("Deployment info:", JSON.stringify(deploymentInfo, null, 2));
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });