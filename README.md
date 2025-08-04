# Kaia Analytics AI

A blockchain-based analytics platform built with Go backend and Ethereum smart contracts using Hardhat.

## Project Structure

```
├── backend/          # Go backend application
├── contracts/        # Smart contracts and deployment scripts
│   ├── contracts/    # Solidity smart contracts
│   ├── test/         # Contract tests
│   ├── ignition/     # Deployment modules
│   └── hardhat.config.js
├── frontend/         # React TypeScript frontend
│   ├── src/          # Source code
│   ├── public/       # Static assets
│   └── package.json
├── README.md
└── .gitignore
```

## Prerequisites

- Node.js (v18 or higher)
- npm or yarn
- Go (v1.21.0 or higher)
- Git

## Installation

### 1. Clone the repository

```bash
git clone <repository-url>
cd kaia-analytics-ai
```

### 2. Install contract dependencies

```bash
cd contracts
npm install
```

### 3. Install frontend dependencies

```bash
cd ../frontend
npm install
```

### 4. Install Go dependencies

```bash
cd ../backend
go mod download
```

## Smart Contracts

The project includes a Lock contract demonstrating time-locked withdrawals.

### Running Contract Tests

```bash
cd contracts
npx hardhat test
```

### Deploying Contracts

```bash
# Deploy to local network
npx hardhat node

# In another terminal
npx hardhat ignition deploy ./ignition/modules/Lock.js --network localhost
```

### Contract Features

- **Lock Contract**: Time-locked Ether storage with owner-only withdrawal
- Comprehensive test coverage
- Gas optimization reporting
- Deployment automation with Hardhat Ignition

## Frontend

The React TypeScript frontend provides a modern web interface for blockchain analytics.

### Running the Frontend

```bash
cd frontend
npm start
```

The frontend will be available at `http://localhost:3000` and includes:
- Interactive blockchain data search
- Real-time network statistics
- Responsive dashboard design
- Block, transaction, and address exploration

## Backend

The Go backend is designed to provide analytics services for blockchain data.

### Running the Backend

```bash
cd backend
go run main.go
```

## Development

### Running Tests

```bash
# Contract tests
cd contracts && npx hardhat test

# Go tests
cd backend && go test ./...

# Frontend tests
cd frontend && npm test
```

### Code Quality

- Solidity contracts follow best practices
- Go code follows standard formatting
- Comprehensive test coverage
- Security auditing with npm audit

## Security

- Regular dependency updates
- Security vulnerability monitoring
- Smart contract best practices implementation
- Access control patterns

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run all tests
6. Submit a pull request

## License

This project is licensed under the UNLICENSED license.

## Support

For support and questions, please open an issue in the repository.