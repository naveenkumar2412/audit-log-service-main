# Audit Log Service

A comprehensive, production-ready audit logging service built with Go, PostgreSQL, and modern cloud-native technologies.

## 🚀 Features

- **RESTful API** - Complete CRUD operations for audit logs
- **Multi-tenant Architecture** - Support for tenant isolation
- **Real-time Notifications** - Email, Slack, and Webhook alerts
- **Authentication & Authorization** - JWT and API key support
- **High Performance** - Connection pooling and optimized queries
- **Observability** - Structured logging and health checks
- **Cloud Native** - Docker, Kubernetes ready
- **Configurable** - YAML-based configuration
- **Database Migrations** - Automated schema management

## 📋 Table of Contents

- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [API Documentation](#api-documentation)
- [Configuration](#configuration)
- [Authentication](#authentication)
- [Notifications](#notifications)
- [Development](#development)
- [Deployment](#deployment)
- [Monitoring](#monitoring)

## 🏗 Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client Apps   │    │   Load Balancer │    │  Audit Service  │
│                 │───▶│    (Nginx)      │───▶│                 │
│ Web/Mobile/API  │    │                 │    │  Gin Framework  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                       │
                       ┌─────────────────┐             │
                       │  Notifications  │◀────────────┘
                       │                 │             │
                       │ Email│Slack│Web │             │
                       └─────────────────┘             │
                                                       │
                       ┌─────────────────┐             │
                       │   PostgreSQL    │◀────────────┘
                       │                 │
                       │  Audit Logs DB  │
                       └─────────────────┘
```

## 🚀 Quick Start

### Using Docker Compose (Recommended)

```bash
# Clone the repository
git clone https://github.com/your-org/audit-log-service.git
cd audit-log-service

# Start all services
make docker-run

# Check service health
curl http://localhost:9025/health
```

### Manual Setup

1. **Prerequisites**

   ```bash
   # Install required tools
   make install-tools

   # Verify Go version (1.21+ required)
   go version
   ```

2. **Database Setup**

   ```bash
   # Start PostgreSQL (using Docker)
   docker run --name audit-postgres -e POSTGRES_PASSWORD=password -d postgres:15

   # Run migrations
   make migrate-up
   ```

3. **Configuration**

   ```bash
   # Copy example config
   make config-example

   # Edit configuration
   vim configs/config.yaml
   ```

4. **Run the Service**

   ```bash
   # Development mode
   make dev

   # Or build and run
   make run
   ```

## 📚 API Documentation

### Authentication

The service supports two authentication methods:

- **JWT Bearer Token**: `Authorization: Bearer <token>`
- **API Key**: `X-API-Key: <api-key>` or `?api_key=<api-key>`

### Endpoints

#### Create Audit Log

```http
POST /api/v1/audit
Content-Type: application/json
X-API-Key: your-api-key

{
  "tenant_id": "tenant-123",
  "user_id": "user-456",
  "resource": "users",
  "event": "USER_CREATED",
  "method": "POST",
  "ip": "192.168.1.100",
  "environment": "production",
  "data": {
    "user_email": "john@example.com",
    "user_role": "admin"
  },
  "meta": {
    "request_id": "req-789",
    "session_id": "sess-012"
  }
}
```

#### List Audit Logs

```http
GET /api/v1/audit?tenant_id=tenant-123&limit=50&offset=0
X-API-Key: your-api-key
```

#### Get Audit Log by ID

```http
GET /api/v1/audit/{id}
X-API-Key: your-api-key
```

#### Delete Audit Log

```http
DELETE /api/v1/audit/{id}
X-API-Key: your-api-key
```

#### Get Statistics

```http
GET /api/v1/audit/stats?tenant_id=tenant-123&start_date=2024-01-01T00:00:00Z&end_date=2024-01-31T23:59:59Z
X-API-Key: your-api-key
```

### Query Parameters for Filtering

- `tenant_id` - Filter by tenant
- `user_id` - Filter by user
- `resource` - Filter by resource type
- `event` - Filter by event type
- `method` - Filter by HTTP method
- `environment` - Filter by environment
- `start_date` - Filter by start date (RFC3339)
- `end_date` - Filter by end date (RFC3339)
- `limit` - Number of results (default: 50, max: 1000)
- `offset` - Pagination offset

## ⚙️ Configuration

### Environment Variables

| Variable      | Description         | Default      |
| ------------- | ------------------- | ------------ |
| `SERVER_HOST` | Server bind address | `0.0.0.0`    |
| `SERVER_PORT` | Server port         | `9025`       |
| `DB_HOST`     | Database host       | `localhost`  |
| `DB_PORT`     | Database port       | `5432`       |
| `DB_USER`     | Database user       | `postgres`   |
| `DB_PASSWORD` | Database password   | `password`   |
| `DB_NAME`     | Database name       | `audit_logs` |
| `JWT_SECRET`  | JWT signing secret  | -            |
| `LOG_LEVEL`   | Logging level       | `info`       |

### Configuration File

```yaml
# configs/config.yaml
server:
  host: "0.0.0.0"
  port: 9025

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "password"
  name: "audit_logs"

auth:
  jwt_secret: "your-secret-key"
  api_keys:
    - "api-key-1"
    - "api-key-2"

notification:
  email:
    enabled: true
    smtp_host: "smtp.gmail.com"
    smtp_port: 587
    username: "your-email@gmail.com"
    password: "your-app-password"

  slack:
    enabled: true
    webhook_url: "https://hooks.slack.com/services/..."

  webhook:
    enabled: true
    urls:
      - "https://your-webhook.com/audit"
```

## 🔐 Authentication

### JWT Authentication

The service supports JWT tokens with the following claims:

```json
{
  "sub": "user-id",
  "tenant_id": "tenant-123",
  "roles": ["admin", "auditor"],
  "exp": 1672531200
}
```

### API Key Authentication

API keys can be configured in the `auth.api_keys` section of the configuration file.

## 📢 Notifications

### Email Notifications

Configure SMTP settings to receive email alerts:

```yaml
notification:
  email:
    enabled: true
    smtp_host: "smtp.gmail.com"
    smtp_port: 587
    username: "alerts@company.com"
    password: "app-specific-password"
    from: "noreply@company.com"
    to:
      - "admin@company.com"
      - "security@company.com"
```

### Slack Notifications

Set up Slack webhook for real-time alerts:

```yaml
notification:
  slack:
    enabled: true
    webhook_url: "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"
    channel: "#security-alerts"
    username: "Audit Bot"
```

### Webhook Notifications

Send audit events to external systems:

```yaml
notification:
  webhook:
    enabled: true
    urls:
      - "https://siem.company.com/api/events"
      - "https://analytics.company.com/audit"
    timeout: 30
```

## 💻 Development

### Development Setup

```bash
# Install development tools
make install-tools

# Run tests
make test

# Run with coverage
make test-coverage

# Run linting
make lint

# Format code
make fmt

# Run security checks
make security
```

### Database Migrations

```bash
# Create new migration
make migrate-create NAME=add_index_to_audit_logs

# Run migrations up
make migrate-up

# Run migrations down
make migrate-down
```

### Testing

```bash
# Run all tests
make test

# Run with race detection
make test-race

# Run benchmarks
make benchmark

# Test API endpoints
make test-api
```

## 🚢 Deployment

### Docker Deployment

```bash
# Build Docker image
make docker-build

# Run with Docker Compose
make docker-run

# View logs
make docker-logs

# Stop services
make docker-stop
```

### Kubernetes Deployment

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: audit-log-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: audit-log-service
  template:
    metadata:
      labels:
        app: audit-log-service
    spec:
      containers:
        - name: audit-log-service
          image: audit-log-service:latest
          ports:
            - containerPort: 9025
          env:
            - name: DB_HOST
              value: "postgres-service"
            - name: JWT_SECRET
              valueFrom:
                secretKeyRef:
                  name: audit-secrets
                  key: jwt-secret
```

### Production Considerations

1. **Security**

   - Use strong JWT secrets
   - Enable HTTPS/TLS
   - Rotate API keys regularly
   - Use database connection encryption

2. **Performance**

   - Configure connection pooling
   - Set up database indices
   - Use Redis for caching
   - Enable compression

3. **Monitoring**
   - Set up health checks
   - Configure log aggregation
   - Monitor database performance
   - Set up alerting

## 📊 Monitoring

### Health Checks

- `/health` - Overall service health
- `/ready` - Readiness probe
- `/live` - Liveness probe

### Metrics

The service exposes metrics for:

- Request count and duration
- Database connection pool status
- Notification delivery status
- Error rates by endpoint

### Logging

Structured JSON logging with the following levels:

- `debug` - Detailed debugging information
- `info` - General information
- `warn` - Warning conditions
- `error` - Error conditions

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices
- Write tests for new features
- Update documentation
- Use conventional commits
- Ensure all tests pass

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- 📧 Email: support@company.com
- 💬 Slack: #audit-service
- 📖 Wiki: [Internal Documentation](https://wiki.company.com/audit-service)
- 🐛 Issues: [GitHub Issues](https://github.com/company/audit-log-service/issues)

---

**Built with ❤️ by the Platform Team**
