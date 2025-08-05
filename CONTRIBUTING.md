# Contributing to KaiaAnalyticsAI

Thank you for your interest in contributing to KaiaAnalyticsAI! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Issue Reporting](#issue-reporting)
- [Community](#community)

## Code of Conduct

This project and its participants are governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## Getting Started

### Prerequisites

- **Node.js**: v18.0.0 or higher
- **Go**: v1.21.0 or higher
- **Git**: Latest version
- **Docker**: For containerized development (optional)

### Quick Start

1. **Fork the repository**
2. **Clone your fork**:
   ```bash
   git clone https://github.com/your-username/kaia-analytics-ai.git
   cd kaia-analytics-ai
   ```
3. **Set up development environment**:
   ```bash
   # Install dependencies
   cd contracts && npm install
   cd ../frontend && npm install
   cd ../backend && go mod download
   ```
4. **Configure environment**:
   ```bash
   cp backend/.env.example backend/.env
   cp frontend/.env.example frontend/.env
   # Edit the .env files with your configuration
   ```

## Development Setup

### Smart Contracts

```bash
cd contracts

# Install dependencies
npm install

# Compile contracts
npx hardhat compile

# Run tests
npx hardhat test

# Start local blockchain
npx hardhat node

# Deploy contracts
npx hardhat ignition deploy ./ignition/modules/AnalyticsRegistry.js --network localhost
```

### Backend Services

```bash
cd backend

# Install dependencies
go mod download

# Run tests
go test ./...

# Start development server
go run main.go

# Build for production
go build -o kaia-analytics-backend main.go
```

### Frontend

```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm start

# Run tests
npm test

# Build for production
npm run build
```

### Docker Development

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

## Coding Standards

### Solidity

- Follow the [Solidity Style Guide](https://docs.soliditylang.org/en/latest/style-guide.html)
- Use NatSpec documentation for all public functions
- Implement comprehensive error handling
- Write extensive test coverage (aim for >90%)
- Use custom errors instead of require strings
- Follow security best practices

```solidity
/**
 * @notice Registers a new analytics task
 * @param _taskType The type of analytics task
 * @param _parameters The task parameters
 * @return taskId The unique task identifier
 */
function registerTask(
    string memory _taskType,
    string memory _parameters
) external payable returns (uint256 taskId) {
    // Implementation
}
```

### Go

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Implement proper error handling
- Write unit tests for all functions
- Use meaningful variable names
- Add comments for complex logic

```go
// ProcessAnalyticsTask processes an analytics task and returns results
func (ae *AnalyticsEngine) ProcessAnalyticsTask(
    ctx context.Context, 
    taskType string, 
    parameters map[string]interface{},
) (*AnalyticsResult, error) {
    // Implementation
}
```

### React/TypeScript

- Follow [React Best Practices](https://reactjs.org/docs/hooks-rules.html)
- Use TypeScript strict mode
- Implement proper component testing
- Follow accessibility guidelines
- Use functional components with hooks
- Implement proper error boundaries

```typescript
interface AnalyticsProps {
  data: AnalyticsData;
  onUpdate: (data: AnalyticsData) => void;
}

const AnalyticsComponent: React.FC<AnalyticsProps> = ({ data, onUpdate }) => {
  // Implementation
};
```

## Testing

### Smart Contract Testing

```bash
cd contracts

# Run all tests
npx hardhat test

# Run with coverage
npx hardhat coverage

# Run specific test file
npx hardhat test test/AnalyticsRegistry.test.js

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

## Pull Request Process

### Before Submitting

1. **Update documentation** for any new features
2. **Add tests** for new functionality
3. **Update CHANGELOG.md** with changes
4. **Ensure all tests pass** before submitting
5. **Follow the coding standards** outlined above

### Pull Request Guidelines

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/amazing-feature
   ```

2. **Make your changes** following the coding standards

3. **Add tests** for new functionality:
   - Unit tests for all new functions
   - Integration tests for new features
   - Update existing tests if needed

4. **Update documentation**:
   - Update README.md if needed
   - Add API documentation
   - Update inline comments

5. **Commit your changes**:
   ```bash
   git add .
   git commit -m "feat: add amazing feature

   - Add new analytics functionality
   - Implement real-time data processing
   - Add comprehensive test coverage
   
   Closes #123"
   ```

6. **Push to your fork**:
   ```bash
   git push origin feature/amazing-feature
   ```

7. **Create a Pull Request**:
   - Use the provided PR template
   - Describe the changes clearly
   - Link any related issues
   - Request review from maintainers

### Commit Message Format

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Test changes
- `chore`: Build/tooling changes

### Pull Request Template

```markdown
## Description
Brief description of the changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests added/updated
- [ ] CHANGELOG.md updated

## Related Issues
Closes #123
```

## Issue Reporting

### Bug Reports

When reporting bugs, please include:

1. **Environment details**:
   - OS and version
   - Node.js version
   - Go version
   - Browser (if applicable)

2. **Steps to reproduce**:
   - Clear, step-by-step instructions
   - Sample data if applicable

3. **Expected vs actual behavior**:
   - What you expected to happen
   - What actually happened

4. **Additional context**:
   - Screenshots if applicable
   - Error messages
   - Console logs

### Feature Requests

When requesting features, please include:

1. **Problem description**:
   - What problem does this solve?
   - Why is this needed?

2. **Proposed solution**:
   - How should this work?
   - Any specific requirements?

3. **Alternatives considered**:
   - What other approaches were considered?

4. **Additional context**:
   - Any relevant examples
   - Related issues

## Community

### Getting Help

- **Documentation**: [https://docs.kaiaanalytics.ai](https://docs.kaiaanalytics.ai)
- **GitHub Issues**: [https://github.com/your-org/kaia-analytics-ai/issues](https://github.com/your-org/kaia-analytics-ai/issues)
- **Discord Community**: [https://discord.gg/kaiaanalytics](https://discord.gg/kaiaanalytics)
- **Email Support**: support@kaiaanalytics.ai

### Community Guidelines

- Be respectful and inclusive
- Help others learn and grow
- Share knowledge and best practices
- Report issues and bugs promptly
- Contribute to documentation

### Recognition

Contributors will be recognized in:
- Project README.md
- Release notes
- Community acknowledgments
- Contributor hall of fame

## License

By contributing to KaiaAnalyticsAI, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to KaiaAnalyticsAI! 🚀