# GAIOL - Go AI Orchestration Layer

GAIOL is a comprehensive AI service orchestration platform built in Go, designed to provide unified access to various AI models and services.

## Project Structure

```
gaiol/
├── api/                    # API layer implementations
│   ├── rest/              # REST API endpoints
│   ├── grpc/              # gRPC service definitions
│   └── graphql/           # GraphQL schema and resolvers
├── cmd/                   # Main applications
│   ├── uaip-service/      # Unified AI Platform service
│   ├── orchestrator/      # Service orchestrator
│   ├── gateway/           # API Gateway
│   └── migrate/           # Database migration tool
├── internal/              # Private application code
│   ├── uaip/             # UAIP core logic
│   ├── orchestrator/      # Orchestration logic
│   ├── models/           # Model interfaces and implementations
│   │   └── adapters/     # AI model adapters
│   ├── security/         # Security and auth
│   └── storage/          # Data storage layer
├── web/                  # Web dashboard
│   └── dashboard/        # React-based admin dashboard
├── tests/                # Test suites
├── docs/                 # Documentation
├── examples/             # Example implementations
├── tools/                # Development tools
├── migrations/           # Database migrations
├── scripts/             # Utility scripts
└── deployments/         # Deployment configurations
```

## Getting Started

1. Install dependencies:
```bash
go mod download
```

2. Set up environment variables:
```bash
cp .env.example .env
```

3. Run the development environment:
```bash
docker-compose -f docker-compose.dev.yml up -d
```

4. Start the service:
```bash
make run
```

## Development

- Use `make test` to run tests
- Use `make build` to build all services
- Use `make lint` to run linters

## License

MIT License
