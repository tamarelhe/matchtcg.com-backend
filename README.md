# MatchTCG Backend

A scalable Go-based API service for Trading Card Game event management, enabling players to create, discover, and manage tournaments and casual matches.

## Overview

MatchTCG backend serves as the single source of truth for all data and business logic, exposing RESTful APIs consumed by separate mobile and web applications. The system is designed for cost-efficiency while maintaining the ability to scale to millions of users.

### Key Features

- **User Management**: Registration, authentication, profile management with GDPR compliance
- **Event Management**: Create, discover, and manage TCG events with geospatial search
- **Group Management**: Private groups with role-based access control
- **RSVP System**: Capacity management with automatic waitlist handling
- **Calendar Integration**: ICS generation and Google Calendar deep links
- **Notification System**: Email notifications for events and updates
- **Geospatial Search**: PostGIS-powered location-based event discovery
- **Internationalization**: Portuguese and English language support

## Architecture

The backend follows clean architecture principles with clear separation of concerns:

```
├── cmd/                    # Application entry points
├── internal/
│   ├── domain/            # Business entities and domain logic
│   ├── usecase/           # Application use cases
│   ├── repository/        # Data access interfaces and implementations
│   ├── handler/           # HTTP handlers and middleware
│   ├── service/           # External service integrations
│   └── config/            # Configuration management
├── migrations/            # Database migrations
├── docker/               # Docker configurations
└── scripts/              # Development and deployment scripts
```

## Technology Stack

- **Language**: Go 1.21+
- **Database**: PostgreSQL 15+ with PostGIS extension
- **Authentication**: JWT with OAuth2 (Google, Apple)
- **Mapping**: OpenStreetMap with Nominatim geocoding
- **Email**: SMTP-based transactional emails
- **Containerization**: Docker and Docker Compose
- **CI/CD**: GitHub Actions

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- Make (optional, for convenience commands)

### Local Development Setup

1. Clone the repository:
```bash
git clone <repository-url>
cd matchtcg-backend
```

2. Start the development environment:
```bash
make dev-up
```

3. Run database migrations:
```bash
make migrate-up
```

4. Start the application:
```bash
make run
```

The API will be available at `http://localhost:8080`

### Environment Configuration

Copy the example environment file and configure your settings:
```bash
cp .env.example .env
```

Key configuration variables:
- `DATABASE_URL`: PostgreSQL connection string
- `JWT_SECRET`: Secret key for JWT token signing
- `SMTP_*`: Email service configuration
- `OAUTH_*`: OAuth provider credentials

## API Documentation

Once running, API documentation is available at:
- Swagger UI: `http://localhost:8080/docs`
- OpenAPI Spec: `http://localhost:8080/api/v1/openapi.json`

## Development

### Available Make Commands

```bash
make help          # Show available commands
make dev-up        # Start development environment
make dev-down      # Stop development environment
make run           # Run the application
make test          # Run all tests
make test-unit     # Run unit tests only
make test-integration # Run integration tests only
make lint          # Run linter
make migrate-up    # Apply database migrations
make migrate-down  # Rollback database migrations
make build         # Build the application
make clean         # Clean build artifacts
```

### Testing

The project includes comprehensive test coverage:

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific test package
go test ./internal/domain/...
```

### Database Migrations

Database schema changes are managed through migrations:

```bash
# Create a new migration
make migrate-create name=add_users_table

# Apply migrations
make migrate-up

# Rollback migrations
make migrate-down
```

## Deployment

### Production Build

```bash
# Build production binary
make build

# Build Docker image
docker build -t matchtcg-backend .
```

### Environment Variables

Required environment variables for production:

- `DATABASE_URL`: PostgreSQL connection string with PostGIS
- `JWT_SECRET`: Strong secret for JWT signing
- `JWT_REFRESH_SECRET`: Secret for refresh tokens
- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`: Email configuration
- `OAUTH_GOOGLE_CLIENT_ID`, `OAUTH_GOOGLE_CLIENT_SECRET`: Google OAuth
- `OAUTH_APPLE_CLIENT_ID`, `OAUTH_APPLE_PRIVATE_KEY`: Apple OAuth
- `NOMINATIM_BASE_URL`: Geocoding service URL
- `CORS_ALLOWED_ORIGINS`: Comma-separated list of allowed origins

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Standards

- Follow Go conventions and best practices
- Write comprehensive tests for new features
- Update documentation for API changes
- Ensure all CI checks pass

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For questions and support, please open an issue in the GitHub repository.