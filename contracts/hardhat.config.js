require("@nomicfoundation/hardhat-toolbox");
require("dotenv").config();

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
  solidity: {
    version: "0.8.19",
    settings: {
      optimizer: {
        enabled: true,
        runs: 200,
      },
    },
  },
  networks: {
    hardhat: {
      chainId: 31337,
    },
    kaia: {
      url: process.env.KAIA_RPC_URL || "https://rpc.kaia.io",
      accounts: process.env.PRIVATE_KEY ? [process.env.PRIVATE_KEY] : [],
      chainId: 8217,
    },
    "kaia-testnet": {
      url: process.env.KAIA_TESTNET_RPC_URL || "https://rpc-testnet.kaia.io",
      accounts: process.env.PRIVATE_KEY ? [process.env.PRIVATE_KEY] : [],
      chainId: 1001,
    },
    "kaia-mainnet": {
      url: process.env.KAIA_MAINNET_RPC_URL || "https://rpc-mainnet.kaia.io",
      accounts: process.env.PRIVATE_KEY ? [process.env.PRIVATE_KEY] : [],
      chainId: 8217,
    },
  },
  gasReporter: {
    enabled: process.env.REPORT_GAS !== undefined,
    currency: "USD",
  },
  etherscan: {
    apiKey: {
      kaia: process.env.KAIASCAN_API_KEY || "",
      "kaia-testnet": process.env.KAIASCAN_API_KEY || "",
    },
    customChains: [
      {
        network: "kaia",
        chainId: 8217,
        urls: {
          apiURL: "https://api.kaiascan.io/api",
          browserURL: "https://kaiascan.io"
        }
      },
      {
        network: "kaia-testnet",
        chainId: 1001,
        urls: {
          apiURL: "https://api-testnet.kaiascan.io/api",
          browserURL: "https://testnet.kaiascan.io"
        }
      }
    ]
  },
  paths: {
    sources: "./contracts",
    tests: "./test",
    cache: "./cache",
    artifacts: "./artifacts"
  },
  mocha: {
    timeout: 40000
  }
};
