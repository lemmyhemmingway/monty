# Monty - Advanced Health Monitoring Service

Monty is a comprehensive health monitoring service written in Go that tracks the health, SSL certificates, and domain expiration dates of web services and domains. Built with Fiber (web framework), GORM (ORM), and PostgreSQL.

## Features

- ✅ **HTTP Health Checks** - Monitor endpoint availability with configurable timeouts, expected status codes, and response time limits
- 🔄 **SSL Certificate Monitoring** - Track SSL certificate validity, expiration dates, issuers, and domain matching
- 📅 **Domain Expiration Control** - Monitor domain registration expiration dates
- 🔄 **Dynamic Discovery** - Automatically adjust monitoring as endpoints are added/removed/updated
- 📊 **Uptime Tracking** - Calculate uptime percentages based on custom success criteria
- 🐳 **Docker Ready** - Multi-stage Docker build for easy deployment
- 📈 **REST API** - Full REST API for managing endpoints and viewing statuses

## Quick Start

### Using Docker (Recommended)

1. Clone the repository:
```bash
git clone https://github.com/lemmyhemmingway/monty.git
cd monty
```

2. Start with Docker Compose:
```bash
docker-compose up -d
```

The service will be available at `http://localhost:3000`

### Manual Setup

1. Install dependencies:
```bash
go mod download
```

2. Set up PostgreSQL database and set environment variable:
```bash
export DATABASE_URL="postgres://user:password@localhost:5432/monty?sslmode=disable"
```

3. Run the application:
```bash
go run main.go
```

## API Documentation

### Health Check
- `GET /health` - Service health status

### Endpoints Management
- `GET /endpoints` - List all monitored endpoints with uptime percentages
- `POST /endpoints` - Create a new endpoint to monitor
- `GET /endpoint-urls` - Get list of all monitored URLs

### Status Monitoring
- `GET /statuses` - Get all status checks (ordered by most recent)
- `GET /endpoints/{id}/statuses` - Get status history for specific endpoint

### Create Endpoint

Create a new HTTP health check endpoint:

```bash
curl -X POST http://localhost:3000/endpoints \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://api.example.com/health",
    "interval": 60,
    "timeout": 30,
    "expected_status_codes": [200, 201],
    "max_response_time": 5000
  }'
```

**Parameters:**
- `url` (required): The URL to monitor
- `interval` (required): Check interval in seconds
- `timeout` (optional): Request timeout in seconds (default: 30)
- `expected_status_codes` (optional): Array of acceptable HTTP status codes (default: 2xx and 3xx)
- `max_response_time` (optional): Maximum response time in milliseconds (default: 5000)

### Response Examples

#### List Endpoints
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "url": "https://api.example.com/health",
    "interval": 60,
    "timeout": 30,
    "expected_status_codes": [200, 201],
    "max_response_time": 5000,
    "created_at": "2024-01-01T00:00:00Z",
    "uptime": 98.5
  }
]
```

#### Get Statuses
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "endpoint_id": "550e8400-e29b-41d4-a716-446655440000",
    "code": 200,
    "response_time": 145,
    "error_message": "",
    "checked_at": "2024-01-01T12:00:00Z"
  }
]
```

## Architecture

- **Handlers**: HTTP request handlers and API endpoints
- **Models**: Database schemas and GORM models
- **Worker**: Background monitoring goroutines with dynamic endpoint discovery
- **Database**: PostgreSQL with automatic migrations

## Development

### Build Commands
- Build: `go build -o monty`
- Run: `go run main.go` or `./monty`
- Test: `go test ./...`
- Lint: `golangci-lint run`

### Testing
Run tests with in-memory SQLite database:
```bash
go test ./...
```

### Project Structure
```
monty/
├── handlers/          # HTTP handlers
├── models/           # Database models
├── worker/           # Background monitoring
├── main.go           # Application entry point
├── docker-compose.yml
├── Dockerfile
└── README.md
```

## Environment Variables

- `DATABASE_URL`: PostgreSQL connection string (required)

## Future Features

- 🔄 SSL Certificate monitoring
- 📅 Domain expiration monitoring
- 📧 Email/Slack notifications for failures
- 📊 Dashboard UI
- 🔐 Authentication and user management

## License

MIT License
