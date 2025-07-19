# M-Data-Storage

A high-performance financial market data collection and storage service built with Go, designed to collect real-time market data from multiple brokers and exchanges.

## Overview

M-Data-Storage is a robust service that collects, processes, and stores financial market data from various sources including cryptocurrency exchanges and stock brokers. The service follows Clean Architecture principles and provides a scalable solution for handling high-frequency market data.

### Key Features

- **Multi-Broker Support**: Connects to multiple cryptocurrency exchanges and stock brokers
- **Real-time Data Collection**: WebSocket-based real-time data streaming
- **Dual Storage Architecture**: SQLite for metadata + QuestDB for time series data
- **High Performance**: Batch processing with configurable worker pools
- **Clean Architecture**: Well-structured codebase following SOLID principles
- **Comprehensive Validation**: Built-in data validation and error handling
- **RESTful API**: HTTP API for data access and management
- **Configurable**: YAML-based configuration with environment variable support
- **Monitoring**: Health checks and structured logging

### Supported Data Types

- **Tickers**: Real-time price and volume data
- **Candles/OHLCV**: Historical price data with configurable timeframes
- **Order Books**: Market depth data with bid/ask levels

### Supported Markets

- **Cryptocurrency**: Spot and futures markets
- **Stock Markets**: Equities, ETFs, and bonds

## Architecture

The project follows Clean Architecture with clear separation of concerns:

```
backend/
├── cmd/server/          # Application entry point
├── internal/
│   ├── domain/          # Domain layer (entities, interfaces)
│   ├── application/     # Application layer (services, use cases)
│   ├── infrastructure/  # Infrastructure layer (databases, external APIs)
│   └── presentation/    # Presentation layer (HTTP handlers, middleware)
├── configs/             # Configuration files
└── docs/               # Documentation
```

## Quick Start

### Prerequisites

- Go 1.21 or higher
- QuestDB (for time series data storage)
- SQLite (automatically created)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd m-data-storage
```

2. Navigate to the backend directory:
```bash
cd backend
```

3. Install dependencies:
```bash
go mod download
```

4. Build the application:
```bash
go build -o bin/m-data-storage ./cmd/server
```

### Configuration

1. Copy the example configuration:
```bash
cp configs/config.yaml configs/config.local.yaml
```

2. Edit the configuration file to match your environment:
```bash
# Edit database connections, API settings, etc.
vim configs/config.local.yaml
```

3. Configure broker connections:
```bash
# Edit broker-specific configurations
vim configs/brokers/binance.yaml
```

### Running the Service

1. Start QuestDB (if using Docker):
```bash
docker run -p 9000:9000 -p 9009:9009 -p 8812:8812 questdb/questdb
```

2. Run the service:
```bash
# Using default configuration
./bin/m-data-storage

# Using custom configuration
./bin/m-data-storage -config=configs/config.local.yaml
```

3. The service will start and begin collecting data. Check the logs for status updates.

### API Access

Once running, the REST API will be available at:
```
http://localhost:8080
```

Example endpoints:
- `GET /health` - Health check
- `GET /api/v1/tickers` - Get ticker data
- `GET /api/v1/candles` - Get candle data
- `GET /api/v1/instruments` - Get instrument information

For complete API documentation, see [docs/API.md](docs/API.md).

## Development

For detailed development information, see [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md).

### Project Structure

- **Domain Layer**: Contains business entities and interfaces
- **Application Layer**: Contains business logic and use cases
- **Infrastructure Layer**: Contains external dependencies (databases, APIs)
- **Presentation Layer**: Contains HTTP handlers and middleware

### Building

```bash
# Build the main application
go build -o bin/m-data-storage ./cmd/server

# Build with race detection (for development)
go build -race -o bin/m-data-storage ./cmd/server

# Cross-compile for different platforms
GOOS=linux GOARCH=amd64 go build -o bin/m-data-storage-linux ./cmd/server
```

### Testing

The project includes comprehensive unit and integration tests:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run integration tests only
go test -v -run "Integration" ./...

# Run tests for specific package
go test ./internal/domain/entities

# Run benchmarks
go test -bench=. ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

**Test Coverage:**
- Unit tests for all core components
- Integration tests for API endpoints
- Integration tests for database operations
- Integration tests for broker connections

### Code Quality

```bash
# Format code
go fmt ./...

# Run linter (requires golangci-lint)
golangci-lint run

# Check for potential issues
go vet ./...

# Generate documentation
godoc -http=:6060
```

## Configuration

### Environment Variables

Key environment variables:

- `APP_ENVIRONMENT`: Application environment (development/production)
- `APP_DEBUG`: Enable debug mode (true/false)
- `API_PORT`: API server port (default: 8080)
- `DB_SQLITE_PATH`: SQLite database path
- `DB_QUESTDB_HOST`: QuestDB host
- `DB_QUESTDB_PORT`: QuestDB port
- `LOG_LEVEL`: Logging level (debug/info/warn/error)

### Configuration Files

- `configs/config.yaml`: Main application configuration
- `configs/brokers/*.yaml`: Broker-specific configurations

For detailed configuration options, see the example configuration files in the `configs/` directory.

## Monitoring

### Health Checks

The service provides health check endpoints:

- `GET /health`: Overall service health
- `GET /health/database`: Database connectivity
- `GET /health/brokers`: Broker connection status

### Logging

Structured JSON logging with configurable levels and outputs:

- Console output for development
- File output with rotation for production
- Contextual logging with request IDs and component information

### Metrics

The service collects various metrics:

- Data collection rates
- Processing latencies
- Error rates
- Connection status

## Contributing

1. Follow Go coding standards and conventions
2. Write tests for new functionality
3. Update documentation as needed
4. Use conventional commit messages
5. Ensure all tests pass before submitting

## License

[License information to be added]

## Deployment

For production deployment instructions, see [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md).

## Documentation

- [API Documentation](docs/API.md) - Complete REST API reference
- [Development Guide](docs/DEVELOPMENT.md) - Developer setup and guidelines
- [Deployment Guide](docs/DEPLOYMENT.md) - Production deployment instructions
- [Architecture Documentation](docs/) - Detailed architecture and design documents

## Support

For questions and support, please refer to the documentation in the `docs/` directory or create an issue in the repository.
