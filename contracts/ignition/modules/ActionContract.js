const { buildModule } = require("@nomicfoundation/hardhat-ignition/modules");

module.exports = buildModule("ActionContract", (m) => {
  const actionContract = m.contract("ActionContract", []);

  return { actionContract };
});