# WKE BaaS API

A Backend-as-a-Service (BaaS) API built with Go that provides project management capabilities with Kubernetes integration. This service allows users to create, manage, and deploy projects on Kubernetes clusters with automated database provisioning.

## Features

- **Project Management**: Create, update, and manage BaaS projects
- **Kubernetes Integration**: Automated deployment of projects to Kubernetes clusters
- **Database Management**: PostgreSQL database provisioning using CloudNative PostgreSQL (CNPG)
- **Authentication**: Integrated authentication system
- **Real-time Updates**: Server-Sent Events (SSE) for live project status updates
- **API Documentation**: Auto-generated OpenAPI documentation with Huma v2

## Architecture

The application follows a layered architecture pattern:

- **Entry Point** (`main.go`): Application initialization and dependency injection
- **Router Layer** (`router/`): HTTP routing using Chi and Huma v2 framework
- **Controllers Layer** (`controllers/`): HTTP request/response handling
- **Services Layer** (`services/`): Business logic implementation
- **Repository Layer** (`repo/`): Data access layer with GORM ORM
- **Models Layer** (`models/`): Database entity definitions
- **DTO Layer** (`dto/`): Data transfer objects for API contracts
- **Infrastructure** (`i3s/`): Database migrations and infrastructure services

## Technology Stack

- **Language**: Go 1.24
- **HTTP Framework**: [Huma v2](https://github.com/danielgtaylor/huma) with [Chi router](https://github.com/go-chi/chi)
- **ORM**: [GORM](https://gorm.io/) with PostgreSQL driver
- **Configuration**: [Viper](https://github.com/spf13/viper) with YAML support
- **Caching**: [go-cache](https://github.com/patrickmn/go-cache) for in-memory caching
- **Kubernetes**: Official [client-go](https://github.com/kubernetes/client-go)
- **Database Migrations**: [golang-migrate](https://github.com/golang-migrate/migrate)
- **ID Generation**: [nanoid](https://github.com/matoous/go-nanoid) for unique identifiers

## Quick Start

### Prerequisites

- Go 1.24 or later
- PostgreSQL database
- Kubernetes cluster access (optional, for deployment features)
- Docker (optional, for containerized deployment)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd baas-api
```

2. Install dependencies:
```bash
go mod download
```

3. Configure the application:
```bash
cp config.yaml.example config.yaml
# Edit config.yaml with your settings
```

4. Run database migrations:
```bash
go run . migrate
```

5. Start the application:
```bash
go run .
```

The API will be available at `http://localhost:8080` (or the port specified in your config).

## Configuration

The application uses a YAML-based configuration system with support for environment variable overrides. Configuration can be provided in two locations:

- `config.yaml` (root directory)
- `config/config.yaml`

### Configuration Structure

```yaml
# Application settings
app:
  port: "8080"
  host: "0.0.0.0"
  trustedOrigins:
    - "http://localhost:5173"
  externalDomain: "baas.wke.csie.ncnu.edu.tw"
  externalSecure: true

# Authentication service
auth:
  url: "http://localhost:3000/api/auth"

# Database connection
database:
  url: "postgres://user:password@localhost:5432/database"

# Kubernetes configuration
kube:
  configPath: "/path/to/kubeconfig"
  project:
    namespace: "baas-project"
    tlsSecretName: "baas-wildcard-tls"
```

### Environment Variables

Configuration values can be overridden using environment variables with dot-to-underscore conversion:

- `app.port` → `APP_PORT`
- `database.url` → `DATABASE_URL`
- `kube.configPath` → `KUBE_CONFIG_PATH`

## Development

### Running the Application

```bash
# Development mode
go run .

# With specific config file
go run . --config path/to/config.yaml
```

### Building

```bash
# Build binary
go build -o baas-api

# Build Docker image
docker build -t baas-api .
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./repo
go test ./services
```

### Database Migrations

```bash
# Run migrations
go run . migrate

# Create new migration
migrate create -ext sql -dir i3s/migrations -seq migration_name
```

## API Documentation

The API provides auto-generated OpenAPI documentation accessible at:

- **Swagger UI**: `http://localhost:8080/docs`
- **OpenAPI Spec**: `http://localhost:8080/openapi.json`

### Main Endpoints

- `GET /api/v1/projects` - List all projects
- `POST /api/v1/projects` - Create a new project
- `GET /api/v1/projects/{id}` - Get project details
- `PUT /api/v1/projects/{id}` - Update project
- `DELETE /api/v1/projects/{id}` - Delete project
- `GET /api/v1/projects/{id}/events` - SSE endpoint for project events

## Kubernetes Integration

The application provides seamless Kubernetes integration for project deployment:

### Features

- **Automatic Deployment**: Projects are automatically deployed to Kubernetes
- **Database Provisioning**: PostgreSQL clusters using CloudNative PostgreSQL (CNPG)
- **Ingress Management**: Automatic ingress configuration for project access
- **Resource Management**: CPU, memory, and storage allocation

### Kubernetes Templates

The `kube-files/` directory contains YAML templates for:

- `project-cnpg-cluster.yaml` - PostgreSQL cluster definition
- `project-cnpg-database.yaml` - Database creation
- `project-ingressroute.yaml` - HTTP ingress routing
- `project-ingressroutetcp.yaml` - TCP ingress routing

## Project Structure

```
├── main.go                 # Application entry point
├── config/                 # Configuration management
├── controllers/            # HTTP controllers
├── services/               # Business logic
├── repo/                   # Data access layer
│   └── kube/              # Kubernetes repositories
├── models/                 # Database models
├── dto/                   # Data transfer objects
├── router/                # HTTP routing
├── i3s/                   # Infrastructure services
│   └── migrations/        # Database migrations
├── kube-files/            # Kubernetes YAML templates
├── utils/                 # Utility functions
└── scripts/               # Database and utility scripts
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`go test ./...`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Development Guidelines

- Follow Go best practices and conventions
- Write tests for new functionality
- Update documentation for API changes
- Ensure all tests pass before submitting PR
- Use meaningful commit messages

## License

This project is licensed under the terms specified in the LICENSE file.

## Support

For questions, issues, or contributions, please:

1. Check the existing issues in the repository
2. Create a new issue for bugs or feature requests
3. Review the documentation and configuration examples
4. Check the Kubernetes logs for deployment issues

## Changelog

See git commit history for detailed changes and updates.