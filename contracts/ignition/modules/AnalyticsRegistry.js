const { buildModule } = require("@nomicfoundation/hardhat-ignition/modules");

module.exports = buildModule("AnalyticsRegistry", (m) => {
  const analyticsRegistry = m.contract("AnalyticsRegistry", [m.getParameter("registrationFee", "1000000000000000")]); // 0.001 ETH

  return { analyticsRegistry };
});