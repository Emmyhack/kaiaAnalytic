const { buildModule } = require("@nomicfoundation/hardhat-ignition/modules");

module.exports = buildModule("DataContract", (m) => {
  const dataContract = m.contract("DataContract", [
    m.getParameter("analyticsStorageFee", "500000000000000"), // 0.0005 ETH
    m.getParameter("tradeStorageFee", "200000000000000")      // 0.0002 ETH
  ]);

  return { dataContract };
});