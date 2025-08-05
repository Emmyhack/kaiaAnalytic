const { buildModule } = require("@nomicfoundation/hardhat-ignition/modules");

module.exports = buildModule("SubscriptionContract", (m) => {
  const subscriptionContract = m.contract("SubscriptionContract", [
    m.getParameter("kaiaTokenAddress", "0x0000000000000000000000000000000000000000") // Placeholder for KAIA token
  ]);

  return { subscriptionContract };
});