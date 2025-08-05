# KaiaAnalyticsAI - Decentralized Analytics Platform

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![Node Version](https://img.shields.io/badge/Node.js-18+-green.svg)](https://nodejs.org/)
[![React Version](https://img.shields.io/badge/React-19+-blue.svg)](https://reactjs.org/)
[![Solidity Version](https://img.shields.io/badge/Solidity-0.8.28+-orange.svg)](https://soliditylang.org/)

> A comprehensive decentralized analytics platform built on the Kaia blockchain, providing real-time analytics for traders, developers, and governance participants.

## 📋 Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [API Documentation](#api-documentation)
- [Smart Contracts](#smart-contracts)
- [Deployment](#deployment)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

## 🎯 Overview

KaiaAnalyticsAI is a production-ready decentralized analytics platform that leverages the Kaia blockchain's high transaction throughput, Service Chains for scalability, and the Kaia Agent Kit for on-chain interactions. The platform provides real-time analytics, AI-powered insights, and interactive features for the DeFi ecosystem.

### Key Capabilities

- **Real-time Analytics**: Live blockchain data and market insights
- **AI-Powered Insights**: Machine learning-based trading suggestions
- **Interactive Chat**: Natural language queries and on-chain actions
- **Premium Features**: KAIA token-based subscription system
- **Multi-chain Support**: Ready for cross-chain analytics

## ✨ Features

### 🔍 Analytics & Insights
- **Yield Opportunity Analysis**: Identify optimal yield farming opportunities
- **Trading Suggestions**: AI-powered recommendations based on user history
- **Portfolio Optimization**: Advanced portfolio analysis and rebalancing
- **Governance Sentiment**: Real-time analysis of governance proposals
- **Risk Assessment**: Comprehensive DeFi position risk analysis

### 💬 Interactive Features
- **Natural Language Chat**: Query data and execute actions via chat
- **Real-time Updates**: WebSocket-powered live data streams
- **On-chain Actions**: Execute staking, voting, and trading directly
- **Premium Subscriptions**: Tier-based access with KAIA tokens

### 🔗 Blockchain Integration
- **Smart Contract Suite**: Complete decentralized functionality
- **Data Storage**: On-chain analytics results and trade history
- **Subscription Management**: Premium tier access control
- **Event Monitoring**: Real-time blockchain event tracking

## 🏗️ Architecture

### Smart Contracts (Solidity)
```
contracts/
├── AnalyticsRegistry.sol    # Task registration and tracking
├── DataContract.sol         # Analytics results storage
├── SubscriptionContract.sol # Premium subscription management
└── ActionContract.sol       # On-chain action execution
```

### Backend Services (GoLang)
```
backend/
├── services/
│   ├── analytics_engine.go  # Analytics computation engine
│   ├── data_collector.go    # Multi-source data collection
│   └── chat_engine.go       # Natural language processing
└── main.go                  # API server and routing
```

### Frontend Dashboard (React + TypeScript)
```
frontend/
├── src/
│   ├── components/          # Reusable UI components
│   ├── services/           # API integration services
│   ├── hooks/              # Custom React hooks
│   └── utils/              # Utility functions
└── public/                 # Static assets
```

## 🚀 Quick Start

### Prerequisites

- **Node.js**: v18.0.0 or higher
- **Go**: v1.21.0 or higher
- **Git**: Latest version
- **Kaia Wallet**: For blockchain interactions

### 1. Clone Repository

```bash
git clone https://github.com/your-org/kaia-analytics-ai.git
cd kaia-analytics-ai
```

### 2. Install Dependencies

```bash
# Install smart contract dependencies
cd contracts && npm install

# Install frontend dependencies
cd ../frontend && npm install

# Install backend dependencies
cd ../backend && go mod download
```

### 3. Environment Setup

```bash
# Copy environment templates
cp backend/.env.example backend/.env
cp frontend/.env.example frontend/.env

# Configure your environment variables
# See Configuration section for details
```

### 4. Deploy Smart Contracts

```bash
cd contracts

# Start local blockchain
npx hardhat node

# Deploy contracts (in new terminal)
npx hardhat ignition deploy ./ignition/modules/AnalyticsRegistry.js --network localhost
npx hardhat ignition deploy ./ignition/modules/DataContract.js --network localhost
npx hardhat ignition deploy ./ignition/modules/SubscriptionContract.js --network localhost
npx hardhat ignition deploy ./ignition/modules/ActionContract.js --network localhost
```

### 5. Start Services

```bash
# Start backend (Terminal 1)
cd backend && go run main.go

# Start frontend (Terminal 2)
cd frontend && npm start
```

### 6. Access Platform

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **API Documentation**: http://localhost:8080/docs

## 📦 Installation

### Detailed Installation Steps

#### Smart Contracts Setup

```bash
cd contracts

# Install dependencies
npm install

# Compile contracts
npx hardhat compile

# Run tests
npx hardhat test

# Generate type definitions
npx hardhat typechain
```

#### Backend Setup

```bash
cd backend

# Install Go dependencies
go mod download

# Run tests
go test ./...

# Build binary
go build -o kaia-analytics-backend main.go

# Run with hot reload (development)
go run main.go
```

#### Frontend Setup

```bash
cd frontend

# Install dependencies
npm install

# Run tests
npm test

# Build for production
npm run build

# Start development server
npm start
```

## 📚 API Documentation

### Authentication

All API endpoints require authentication via API key or wallet signature:

```bash
# API Key Authentication
Authorization: Bearer YOUR_API_KEY

# Wallet Authentication
X-Wallet-Signature: SIGNATURE
X-Wallet-Address: ADDRESS
```

### Core Endpoints

#### Analytics API

```http
POST /api/v1/analytics/yield
Content-Type: application/json

{
  "protocols": ["uniswap", "aave"],
  "minAPY": 5.0,
  "maxRisk": 0.3
}
```

```http
POST /api/v1/analytics/trading-suggestions
Content-Type: application/json

{
  "userAddress": "0x...",
  "portfolio": [...],
  "riskTolerance": "moderate"
}
```

#### Data Collection API

```http
GET /api/v1/data/market?symbols=ETH,USDC,KAIA
GET /api/v1/data/protocols?category=defi
GET /api/v1/data/gas?network=kaia
```

#### Chat API

```http
POST /api/v1/chat/message
Content-Type: application/json

{
  "message": "What are the best yield opportunities?",
  "userId": "user123",
  "sessionId": "session456"
}
```

#### WebSocket Connection

```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/chat/ws');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Received:', data);
};
```

### Response Format

All API responses follow a standardized format:

```json
{
  "success": true,
  "data": {
    // Response data
  },
  "timestamp": 1640995200,
  "requestId": "req_123456"
}
```

## 🔐 Smart Contracts

### AnalyticsRegistry

Manages analytics task registration and completion tracking.

```solidity
// Register a new analytics task
function registerTask(
    string memory _taskType,
    string memory _parameters
) external payable returns (uint256 taskId);

// Complete a task with results
function completeTask(
    uint256 _taskId,
    string memory _resultHash
) external onlyOwner;
```

### DataContract

Stores analytics results and anonymized user data on-chain.

```solidity
// Store analytics result
function storeAnalyticsResult(
    uint256 _taskId,
    string memory _resultData,
    string memory _resultHash
) external payable;

// Store trade data
function storeTradeData(
    address _user,
    string memory _tradeData,
    string memory _hash
) external payable;
```

### SubscriptionContract

Manages premium subscriptions with KAIA token payments.

```solidity
// Purchase subscription
function purchaseSubscription(uint256 _tierId) 
    external returns (uint256 subscriptionId);

// Check subscription status
function getUserSubscriptionStatus(address _user) 
    external view returns (bool, uint256, uint256);
```

### ActionContract

Executes on-chain actions triggered by chat interface.

```solidity
// Request action execution
function requestAction(
    string memory _actionType,
    bytes memory _parameters
) external returns (uint256 requestId);

// Execute approved action
function executeAction(uint256 _requestId) external;
```

## 🚀 Deployment

### Vercel Deployment

The frontend is optimized for Vercel deployment:

1. **Connect Repository**: Link your GitHub repository to Vercel
2. **Configure Build**: Set build command to `npm run build`
3. **Set Environment Variables**: Configure all required environment variables
4. **Deploy**: Vercel will automatically deploy on push to main branch

#### Vercel Configuration

```json
{
  "buildCommand": "npm run build",
  "outputDirectory": "build",
  "installCommand": "npm install",
  "framework": "create-react-app"
}
```

### Backend Deployment

#### Docker Deployment

```dockerfile
# Dockerfile for backend
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
```

#### Cloud Deployment

```bash
# Deploy to Google Cloud Run
gcloud run deploy kaia-analytics-backend \
  --source . \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated

# Deploy to AWS ECS
aws ecs create-service \
  --cluster kaia-analytics \
  --service-name backend \
  --task-definition backend:1
```

### Smart Contract Deployment

#### Kaia Testnet

```bash
# Deploy to Kaia testnet
npx hardhat ignition deploy ./ignition/modules/AnalyticsRegistry.js --network kaia-testnet
npx hardhat ignition deploy ./ignition/modules/DataContract.js --network kaia-testnet
npx hardhat ignition deploy ./ignition/modules/SubscriptionContract.js --network kaia-testnet
npx hardhat ignition deploy ./ignition/modules/ActionContract.js --network kaia-testnet
```

#### Kaia Mainnet

```bash
# Deploy to Kaia mainnet (production)
npx hardhat ignition deploy ./ignition/modules/AnalyticsRegistry.js --network kaia-mainnet
npx hardhat ignition deploy ./ignition/modules/DataContract.js --network kaia-mainnet
npx hardhat ignition deploy ./ignition/modules/SubscriptionContract.js --network kaia-mainnet
npx hardhat ignition deploy ./ignition/modules/ActionContract.js --network kaia-mainnet
```

## 🧪 Testing

### Smart Contract Testing

```bash
cd contracts

# Run all tests
npx hardhat test

# Run specific test file
npx hardhat test test/AnalyticsRegistry.test.js

# Run with coverage
npx hardhat coverage

# Gas optimization
npx hardhat size-contracts
```

### Backend Testing

```bash
cd backend

# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test ./services -v

# Benchmark tests
go test -bench=. ./...
```

### Frontend Testing

```bash
cd frontend

# Run unit tests
npm test

# Run with coverage
npm test -- --coverage

# Run E2E tests
npm run test:e2e

# Run specific test
npm test -- --testNamePattern="Analytics"
```

### Integration Testing

```bash
# Run full integration test suite
npm run test:integration

# Test API endpoints
npm run test:api

# Test WebSocket connections
npm run test:websocket
```

## 🔧 Configuration

### Environment Variables

#### Backend (.env)

```bash
# Server Configuration
PORT=8080
ENVIRONMENT=production
LOG_LEVEL=info

# Blockchain Configuration
ETH_NODE_URL=https://kaia-mainnet.infura.io/v3/YOUR_PROJECT_ID
KAIA_NODE_URL=https://kaia-mainnet.kaia.io
NETWORK_ID=1

# Contract Addresses
ANALYTICS_REGISTRY_ADDRESS=0x...
DATA_CONTRACT_ADDRESS=0x...
SUBSCRIPTION_CONTRACT_ADDRESS=0x...
ACTION_CONTRACT_ADDRESS=0x...

# API Keys
COINGECKO_API_KEY=your-coingecko-api-key
KAIA_API_KEY=your-kaia-api-key

# Database Configuration
DATABASE_URL=postgresql://user:password@localhost:5432/kaia_analytics
REDIS_URL=redis://localhost:6379

# Security
JWT_SECRET=your-jwt-secret
CORS_ORIGIN=https://your-domain.vercel.app
```

#### Frontend (.env)

```bash
# API Configuration
REACT_APP_API_URL=https://your-backend.herokuapp.com
REACT_APP_WS_URL=wss://your-backend.herokuapp.com

# Blockchain Configuration
REACT_APP_NETWORK_ID=1
REACT_APP_RPC_URL=https://kaia-mainnet.infura.io/v3/YOUR_PROJECT_ID

# Contract Addresses
REACT_APP_ANALYTICS_REGISTRY_ADDRESS=0x...
REACT_APP_DATA_CONTRACT_ADDRESS=0x...
REACT_APP_SUBSCRIPTION_CONTRACT_ADDRESS=0x...
REACT_APP_ACTION_CONTRACT_ADDRESS=0x...

# Feature Flags
REACT_APP_ENABLE_CHAT=true
REACT_APP_ENABLE_PREMIUM=true
REACT_APP_ENABLE_ANALYTICS=true
```

### Network Configuration

```javascript
// hardhat.config.js
module.exports = {
  networks: {
    localhost: {
      url: "http://127.0.0.1:8545"
    },
    kaiaTestnet: {
      url: "https://kaia-testnet.kaia.io",
      chainId: 1337,
      accounts: [process.env.PRIVATE_KEY]
    },
    kaiaMainnet: {
      url: "https://kaia-mainnet.kaia.io",
      chainId: 1,
      accounts: [process.env.PRIVATE_KEY]
    }
  }
};
```

## 📊 Performance

### Benchmarks

- **API Response Time**: < 200ms average
- **WebSocket Latency**: < 50ms
- **Smart Contract Gas**: Optimized for efficiency
- **Concurrent Users**: 1000+ supported
- **Data Freshness**: Real-time blockchain data

### Optimization Features

- **Service Worker**: Offline functionality
- **Code Splitting**: Lazy loading of components
- **Image Optimization**: WebP format with fallbacks
- **Caching Strategy**: Redis-based caching
- **CDN Integration**: Global content delivery

## 🔒 Security

### Smart Contract Security

- ✅ **Reentrancy Protection**: All contracts use `ReentrancyGuard`
- ✅ **Access Control**: Proper `Ownable` implementation
- ✅ **Input Validation**: Comprehensive parameter validation
- ✅ **Gas Optimization**: Custom errors and efficient patterns
- ✅ **Audit Ready**: Follows industry security standards

### Backend Security

- ✅ **Input Sanitization**: All user inputs validated
- ✅ **Rate Limiting**: DDoS protection implemented
- ✅ **CORS Configuration**: Secure cross-origin handling
- ✅ **Environment Protection**: Sensitive data in environment variables
- ✅ **HTTPS Enforcement**: SSL/TLS required in production

### Frontend Security

- ✅ **XSS Protection**: Content Security Policy implemented
- ✅ **CSRF Protection**: Token-based request validation
- ✅ **Wallet Security**: Secure Web3Modal integration
- ✅ **Dependency Scanning**: Regular security audits

## 🤝 Contributing

We welcome contributions from the community! Please read our contributing guidelines before submitting pull requests.

### Development Setup

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Make your changes**: Follow the coding standards
4. **Add tests**: Ensure all new code is tested
5. **Run tests**: `npm test && go test ./...`
6. **Submit PR**: Create a pull request with detailed description

### Coding Standards

#### Solidity
- Follow [Solidity Style Guide](https://docs.soliditylang.org/en/latest/style-guide.html)
- Use NatSpec documentation
- Implement comprehensive error handling
- Write extensive test coverage

#### Go
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Implement proper error handling
- Write unit tests for all functions

#### React/TypeScript
- Follow [React Best Practices](https://reactjs.org/docs/hooks-rules.html)
- Use TypeScript strict mode
- Implement proper component testing
- Follow accessibility guidelines

### Pull Request Process

1. **Update documentation** for any new features
2. **Add tests** for new functionality
3. **Update CHANGELOG.md** with changes
4. **Ensure all tests pass** before submitting
5. **Request review** from maintainers

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

### Getting Help

- **Documentation**: [https://docs.kaiaanalytics.ai](https://docs.kaiaanalytics.ai)
- **GitHub Issues**: [https://github.com/your-org/kaia-analytics-ai/issues](https://github.com/your-org/kaia-analytics-ai/issues)
- **Discord Community**: [https://discord.gg/kaiaanalytics](https://discord.gg/kaiaanalytics)
- **Email Support**: support@kaiaanalytics.ai

### Community

- **Discord**: Join our community for discussions and support
- **Twitter**: Follow [@KaiaAnalytics](https://twitter.com/KaiaAnalytics) for updates
- **Blog**: Read our [technical blog](https://blog.kaiaanalytics.ai)
- **Newsletter**: Subscribe to our [monthly newsletter](https://kaiaanalytics.ai/newsletter)

## 🙏 Acknowledgments

- **Kaia Blockchain Team**: For the amazing blockchain infrastructure
- **OpenZeppelin**: For secure smart contract libraries
- **React Team**: For the excellent frontend framework
- **Go Team**: For the powerful backend language
- **Community Contributors**: For all the valuable contributions

---

**Built with ❤️ for the Kaia ecosystem**

[![Kaia Analytics AI](https://img.shields.io/badge/KaiaAnalyticsAI-Platform-blue.svg)](https://kaiaanalytics.ai)
[![Made with Love](https://img.shields.io/badge/Made%20with-Love-red.svg)](https://kaiaanalytics.ai)