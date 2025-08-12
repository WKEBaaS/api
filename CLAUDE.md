# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Running the Application
- Main entry point: `go run .` or `go run main.go`
- Note: The README shows `go run ./cmd` but this appears to be outdated - the main.go is in the root directory

### Building
- Build binary: `go build -o baas-api`
- Docker build: `docker build -t baas-api .`

### Testing
- Run all tests: `go test ./...`
- Run specific package tests: `go test ./repo`
- Test with verbose output: `go test -v ./...`

### Dependencies
- Install/update dependencies: `go mod download`
- Tidy dependencies: `go mod tidy`

## Architecture Overview

This is a Backend-as-a-Service (BaaS) API built with Go that follows a layered architecture pattern. The system manages projects with Kubernetes integration and database operations.

### Key Architectural Layers

1. **Entry Point (main.go)**: Application initialization, dependency injection
2. **Router Layer (router/)**: HTTP routing using Chi and Huma v2 framework
3. **Controllers Layer (controllers/)**: HTTP request/response handling
4. **Services Layer (services/)**: Business logic implementation
5. **Repository Layer (repo/)**: Data access layer with GORM ORM
6. **Models Layer (models/)**: Database entity definitions
7. **DTO Layer (dto/)**: Data transfer objects for API contracts
8. **Infrastructure (i3s/)**: Database migrations and infrastructure services

### Framework Stack

- **HTTP Framework**: Huma v2 with Chi router
- **ORM**: GORM with PostgreSQL driver
- **Configuration**: Viper with YAML config files
- **Caching**: go-cache for in-memory caching
- **Kubernetes Client**: Official k8s.io/client-go
- **CLI**: humacli for command-line interface

### Key Components

#### Configuration System
- Uses Viper for configuration management
- Supports both YAML files and environment variables
- Configuration files: `config.yaml` (root) and `config/config.yaml`
- Environment variables override config file values with dot-to-underscore conversion

#### Dependency Injection Pattern
The main.go follows a clear dependency injection pattern:
1. Load configuration
2. Initialize database connection
3. Run migrations via i3s
4. Create repositories with database and cache dependencies
5. Create services with repository dependencies
6. Create controllers with service dependencies
7. Register controllers with router

#### Database Layer
- Uses GORM with PostgreSQL
- Custom migration system in `i3s/` package
- Repositories follow interface-based design
- Includes both regular and Kubernetes-specific repositories in `repo/kube/`

#### API Design
- RESTful APIs with Huma v2 framework
- Automatic OpenAPI documentation generation
- Middleware support (auth, CORS, logging)
- Server-Sent Events (SSE) support for real-time updates
- Versioned APIs under `/v1` prefix

#### Kubernetes Integration
- Custom Kubernetes client in `repo/kube/`
- Manages project deployments, services, and ingress
- YAML templates in `kube-files/` directory
- CloudNative PostgreSQL (CNPG) integration for database management

### File Organization Patterns

- **Controllers**: Interface-based design with private implementations
- **Services**: Business logic with clear separation of concerns  
- **Repositories**: Data access abstraction with caching layer
- **Models**: GORM entities with database tags
- **DTOs**: Separate input/output structures for API endpoints

### Testing Strategy

- Test files use `_test.go` suffix
- Integration tests connect to actual database
- Test setup in `TestMain` function for shared resources
- Repository layer has comprehensive test coverage

### Configuration Structure

The config system supports:
- App settings (host, port, CORS origins)
- Database connection URL
- Authentication service URL
- Kubernetes configuration and namespaces

### Special Notes

- The application uses "i3s" as an infrastructure services package for migrations
- Project references are 20-character strings (likely nanoid-based)
- Architecture diagram available in `images/api-architecture.mmd` (Mermaid format)
- Docker configuration builds a binary named "auth" (may need updating)