# KaiaAnalyticsAI

A decentralized analytics platform for the Kaia blockchain that provides real-time analytics, AI-powered insights, and on-chain action execution.

## 🚀 Features

### Core Analytics
- **Real-time Dashboard**: Live transaction volumes, gas trends, and market data
- **Yield Analysis**: Identify the best yield farming opportunities across Kaia protocols
- **Trading Insights**: AI-powered trading suggestions based on market analysis
- **Governance Tracking**: Monitor governance proposals and sentiment analysis

### AI Chat Assistant
- **Natural Language Queries**: Ask questions in plain English
- **Real-time Responses**: WebSocket-based chat with instant replies
- **On-chain Actions**: Execute staking, voting, and trading actions through chat
- **Contextual Suggestions**: Smart recommendations based on user queries

### Subscription System
- **Free Tier**: Basic analytics and transaction tracking
- **Premium Tier**: Advanced features, personalized insights, and priority access
- **KAIA Token Payments**: Subscription payments in native KAIA tokens

### Blockchain Integration
- **Smart Contracts**: Four main contracts for analytics, data storage, subscriptions, and actions
- **High TPS Support**: Leverages Kaia's high transaction throughput
- **Service Chain Integration**: Scalable computation using Kaia Service Chains
- **Kaia Agent Kit**: Seamless on-chain action execution

## 🏗️ Architecture

### Smart Contracts (Solidity)
- **AnalyticsRegistry**: Manages analytics tasks and metadata
- **DataContract**: Stores analytics results and anonymized trade history
- **SubscriptionContract**: Handles premium subscriptions with KAIA payments
- **ActionContract**: Executes on-chain actions triggered by chat

### Backend Services (GoLang)
- **AnalyticsEngine**: Statistical computations and ML-powered insights
- **DataCollector**: Real-time data collection from multiple sources
- **ChatEngine**: Natural language processing and WebSocket communication
- **BlockchainClient**: Kaia blockchain integration and contract interactions

### Frontend Dashboard (React)
- **Modern UI**: Responsive design with Tailwind CSS
- **Real-time Charts**: Interactive visualizations using go-echarts
- **WebSocket Chat**: Live communication with AI assistant
- **Wallet Integration**: Seamless wallet connection and transaction signing

## 📁 Project Structure

```
kaia-analytics-ai/
├── contracts/                 # Smart contracts
│   ├── contracts/
│   │   ├── AnalyticsRegistry.sol
│   │   ├── DataContract.sol
│   │   ├── SubscriptionContract.sol
│   │   ├── ActionContract.sol
│   │   └── MockERC20.sol
│   ├── scripts/
│   │   └── deploy.js
│   ├── test/
│   │   └── KaiaAnalyticsAI.test.js
│   └── package.json
├── backend/                   # Go backend services
│   ├── internal/
│   │   ├── analytics/
│   │   ├── chat/
│   │   ├── collector/
│   │   ├── config/
│   │   ├── contracts/
│   │   └── middleware/
│   ├── main.go
│   ├── go.mod
│   ├── Dockerfile
│   └── .env.example
├── frontend/                  # React frontend
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   ├── services/
│   │   ├── hooks/
│   │   └── App.js
│   ├── package.json
│   └── tailwind.config.js
└── README.md
```

## 🛠️ Installation & Setup

### Prerequisites
- Node.js 18+ and npm
- Go 1.21+
- Docker (optional)
- Kaia testnet access

### 1. Smart Contracts Setup

```bash
cd contracts
npm install
npm run compile
npm test
```

### 2. Backend Setup

```bash
cd backend
go mod download
cp .env.example .env
# Edit .env with your configuration
go run main.go
```

### 3. Frontend Setup

```bash
cd frontend
npm install
npm start
```

### 4. Environment Configuration

Create `.env` files with the following variables:

**Backend (.env)**
```env
SERVER_ADDRESS=:8080
LOG_LEVEL=info
KAIA_RPC_URL=https://testnet-rpc.kaia.network
KAIA_CHAIN_ID=1337
CONTRACT_ANALYTICS_REGISTRY=0x...
CONTRACT_DATA=0x...
CONTRACT_SUBSCRIPTION=0x...
CONTRACT_ACTION=0x...
KAIA_API_KEY=your_api_key
```

**Frontend (.env)**
```env
REACT_APP_API_URL=http://localhost:8080
```

## 🚀 Deployment

### Smart Contracts
```bash
cd contracts
npm run deploy:testnet  # Deploy to testnet
npm run deploy:mainnet  # Deploy to mainnet
```

### Backend Services
```bash
cd backend
docker build -t kaia-analytics-backend .
docker run -p 8080:8080 kaia-analytics-backend
```

### Frontend
```bash
cd frontend
npm run build
# Deploy to your preferred hosting service
```

## 🔧 Development

### Running Tests
```bash
# Smart contracts
cd contracts && npm test

# Backend
cd backend && go test ./...

# Frontend
cd frontend && npm test
```

### Code Quality
- Solidity: Use Hardhat for compilation and testing
- Go: Follow Go best practices and use `gofmt`
- React: Use ESLint and Prettier for code formatting

## 📊 API Documentation

### Analytics Endpoints
- `GET /api/v1/analytics/yield` - Yield opportunities
- `GET /api/v1/analytics/governance` - Governance sentiment
- `GET /api/v1/analytics/trading` - Trading suggestions
- `GET /api/v1/analytics/volume` - Transaction volume
- `GET /api/v1/analytics/gas` - Gas price trends

### Chat Endpoints
- `POST /api/v1/chat/query` - Send chat query
- `GET /api/v1/chat/ws` - WebSocket connection

### Subscription Endpoints
- `GET /api/v1/subscription/plans` - Available plans
- `POST /api/v1/subscription/purchase` - Purchase subscription
- `GET /api/v1/subscription/status/:address` - User subscription status

## 🔐 Security

### Smart Contract Security
- OpenZeppelin contracts for security best practices
- Reentrancy guards on all external functions
- Access control with `Ownable` pattern
- Comprehensive testing with edge cases

### Backend Security
- Input validation and sanitization
- Rate limiting on API endpoints
- CORS configuration
- Secure WebSocket connections

### Frontend Security
- Wallet signature verification
- Secure API communication
- Input validation
- XSS protection

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Documentation**: [Wiki](https://github.com/your-repo/wiki)
- **Issues**: [GitHub Issues](https://github.com/your-repo/issues)
- **Discord**: [Community Server](https://discord.gg/kaia)
- **Email**: support@kaiaanalytics.ai

## 🙏 Acknowledgments

- Kaia blockchain team for the innovative platform
- OpenZeppelin for secure smart contract libraries
- The Go and React communities for excellent tooling
- All contributors and early adopters

---

**Built with ❤️ for the Kaia ecosystem**