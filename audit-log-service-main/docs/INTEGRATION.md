# Audit Log Service - Integration Guide

## Table of Contents

1. [Overview](#overview)
2. [Quick Start](#quick-start)
3. [Configuration](#configuration)
4. [Authentication](#authentication)
5. [API Reference](#api-reference)
6. [Integration Examples](#integration-examples)
7. [Deployment](#deployment)
8. [Monitoring & Health Checks](#monitoring--health-checks)
9. [Troubleshooting](#troubleshooting)
10. [Best Practices](#best-practices)

## Overview

The Audit Log Service is a comprehensive solution for tracking and managing audit logs across your applications. It provides secure, scalable, and flexible logging capabilities with support for multiple authentication methods, filtering, and real-time notifications.

### Key Features

- 🔐 **Multiple Authentication Methods**: JWT and API Key authentication
- 📊 **Advanced Filtering**: Filter by tenant, user, resource, event type, date range, and more
- 📈 **Statistics & Analytics**: Get insights into audit log patterns
- 🔔 **Real-time Notifications**: Email, Slack, and webhook notifications
- 🚀 **High Performance**: Optimized for high-throughput environments
- 🐳 **Container Ready**: Docker and Kubernetes deployment support
- 💾 **PostgreSQL Backend**: Reliable data persistence with migration support

## Quick Start

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- Docker (optional)

### 1. Clone and Setup

```bash
git clone <repository-url>
cd audit-log-service
cp configs/config.example.yaml configs/config.yaml
```

### 2. Configure Database

Edit `configs/config.yaml`:

```yaml
database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "your_password"
  name: "audit_logs"
  ssl_mode: "disable"
```

### 3. Install Dependencies and Tools

```bash
make install-tools
go mod tidy
```

### 4. Run Database Migrations

```bash
make migrate-up
```

### 5. Start the Service

```bash
# Development mode (with hot reload)
make dev

# Production mode
make build && ./bin/audit-log-service
```

The service will start on `http://localhost:9025` by default.

## Configuration

### Environment-Specific Configuration

Create separate configuration files for different environments:

#### Development (`configs/config.dev.yaml`)

```yaml
server:
  host: "0.0.0.0"
  port: 9025

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "dev_password"
  name: "audit_logs_dev"

audit:
  enabled: true
  default_status: "pending"
  status_values:
    - "pending"
    - "processing"
    - "completed"
    - "failed"
    - "archived"

logging:
  level: "debug"
  format: "text"

notification:
  email:
    enabled: false
  slack:
    enabled: false
  webhook:
    enabled: false
```

#### Production (`configs/config.prod.yaml`)

```yaml
server:
  host: "0.0.0.0"
  port: 9025
  read_timeout: 30
  write_timeout: 30

database:
  host: "${DB_HOST}"
  port: ${DB_PORT}
  user: "${DB_USER}"
  password: "${DB_PASSWORD}"
  name: "${DB_NAME}"
  ssl_mode: "require"
  max_open_conns: 50
  max_idle_conns: 25

auth:
  jwt_secret: "${JWT_SECRET}"
  api_keys:
    - "${API_KEY_1}"
    - "${API_KEY_2}"

notification:
  email:
    enabled: true
    smtp_host: "${SMTP_HOST}"
    smtp_port: ${SMTP_PORT}
    username: "${SMTP_USERNAME}"
    password: "${SMTP_PASSWORD}"

logging:
  level: "info"
  format: "json"
  output: "both"
```

### Configuration Parameters

| Parameter           | Description         | Default    | Required |
| ------------------- | ------------------- | ---------- | -------- |
| `server.host`       | Server bind address | "0.0.0.0"  | No       |
| `server.port`       | Server port         | 9025       | No       |
| `database.host`     | PostgreSQL host     | localhost  | Yes      |
| `database.port`     | PostgreSQL port     | 5432       | No       |
| `database.user`     | Database user       | postgres   | Yes      |
| `database.password` | Database password   | -          | Yes      |
| `database.name`     | Database name       | audit_logs | Yes      |
| `auth.jwt_secret`   | JWT signing secret  | -          | Yes      |
| `auth.api_keys`     | Valid API keys      | []         | Yes      |
| `audit.enabled`     | Enable audit status validation | true | No |
| `audit.default_status` | Default status for new audit logs | "pending" | No |
| `audit.status_values` | Valid status values | ["pending", "processing", "completed", "failed", "archived"] | No |

## Authentication

The service supports two authentication methods:

### 1. JWT Authentication

Use JWT tokens in the Authorization header:

```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     -X GET http://localhost:9025/api/v1/audit
```

**JWT Token Structure:**

```json
{
  "sub": "user-id",
  "tenant_id": "tenant-123",
  "role": "admin",
  "exp": 1640995200
}
```

### 2. API Key Authentication

Use API keys in the X-API-Key header:

```bash
curl -H "X-API-Key: your-api-key-1" \
     -X GET http://localhost:9025/api/v1/audit
```

## API Reference

### Base URL

- **Local Development**: `http://localhost:9025`
- **Production**: `https://your-domain.com`

### Health Check Endpoints

#### GET /health

Basic health check - returns service status.

**Response:**

```json
{
  "status": "healthy",
  "timestamp": "2025-06-14T10:30:00Z",
  "version": "1.0.0",
  "uptime": "2h30m15s"
}
```

#### GET /ready

Readiness check - verifies database connectivity.

#### GET /live

Liveness check - for Kubernetes deployments.

### Audit Log Endpoints

#### POST /api/v1/audit

Create a new audit log entry.

**Request Body:**

```json
{
  "tenant_id": "tenant-123",
  "user_id": "user-456",
  "resource": "users",
  "event": "user_created",
  "method": "POST",
  "ip": "192.168.1.100",
  "status": "pending",
  "environment": "production",
  "data": {
    "user_email": "john.doe@example.com",
    "user_role": "admin"
  },
  "meta": {
    "request_id": "req-12345",
    "session_id": "sess-67890"
  }
}
```

**Response (201 Created):**

```json
{
  "id": "01234567-89ab-cdef-0123-456789abcdef",
  "tenant_id": "tenant-123",
  "user_id": "user-456",
  "resource": "users",
  "event": "user_created",
  "method": "POST",
  "ip": "192.168.1.100",
  "status": "pending",
  "environment": "production",
  "data": {...},
  "meta": {...},
  "timestamp": "2025-06-14T10:30:00Z",
  "created_at": "2025-06-14T10:30:00Z",
  "updated_at": "2025-06-14T10:30:00Z"
}
```

#### GET /api/v1/audit

List audit logs with filtering and pagination.

**Query Parameters:**

- `tenant_id` (string): Filter by tenant ID
- `user_id` (string): Filter by user ID
- `resource` (string): Filter by resource type
- `event` (string): Filter by event type
- `method` (string): Filter by HTTP method
- `status` (string): Filter by status (pending, processing, completed, failed, archived)
- `environment` (string): Filter by environment
- `start_date` (RFC3339): Start date for filtering
- `end_date` (RFC3339): End date for filtering
- `limit` (int): Number of results (default: 50, max: 1000)
- `offset` (int): Number of results to skip (default: 0)

**Example:**

```bash
GET /api/v1/audit?tenant_id=tenant-123&resource=users&limit=10&offset=0
```

**Response:**

```json
{
  "data": [
    {
      "id": "01234567-89ab-cdef-0123-456789abcdef",
      "tenant_id": "tenant-123",
      "user_id": "user-456",
      "resource": "users",
      "event": "user_created",
      "method": "POST",
      "ip": "192.168.1.100",
      "environment": "production",
      "timestamp": "2025-06-14T10:30:00Z",
      "created_at": "2025-06-14T10:30:00Z",
      "updated_at": "2025-06-14T10:30:00Z"
    }
  ],
  "total": 1,
  "limit": 10,
  "offset": 0,
  "has_more": false
}
```

#### GET /api/v1/audit/{id}

Get a specific audit log by ID.

#### PUT /api/v1/audit/{id}/status

Update the status of a specific audit log entry.

**Request Body:**

```json
{
  "status": "completed"
}
```

**Response (200 OK):**

```json
{
  "message": "Audit log status updated successfully",
  "id": "01234567-89ab-cdef-0123-456789abcdef",
  "status": "completed"
}
```

**Available Status Values (configurable in config.yaml):**
- `pending` (default)
- `processing` 
- `completed`
- `failed`
- `archived`

#### GET /api/v1/audit/stats

Get audit log statistics.

**Query Parameters:**

- `tenant_id` (string, required): Tenant ID
- `start_date` (RFC3339, required): Start date
- `end_date` (RFC3339, required): End date

**Response:**

```json
{
  "total_logs": 1250,
  "date_range": {
    "start": "2025-06-01T00:00:00Z",
    "end": "2025-06-14T23:59:59Z"
  },
  "breakdown_by_resource": {
    "users": 450,
    "orders": 380,
    "products": 290
  },
  "breakdown_by_event": {
    "created": 520,
    "updated": 430,
    "deleted": 180
  }
}
```

## Integration Examples

### Go Client Example

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type AuditLogRequest struct {
    TenantID    string                 `json:"tenant_id"`
    UserID      string                 `json:"user_id"`
    Resource    string                 `json:"resource"`
    Event       string                 `json:"event"`
    Method      string                 `json:"method"`
    IP          string                 `json:"ip"`
    Environment string                 `json:"environment"`
    Data        map[string]interface{} `json:"data"`
    Meta        map[string]interface{} `json:"meta"`
}

func createAuditLog(apiKey string, auditLog AuditLogRequest) error {
    jsonData, err := json.Marshal(auditLog)
    if err != nil {
        return err
    }

    req, err := http.NewRequest("POST", "http://localhost:9025/api/v1/audit", bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-API-Key", apiKey)

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusCreated {
        return fmt.Errorf("failed to create audit log: %s", resp.Status)
    }

    return nil
}

func main() {
    auditLog := AuditLogRequest{
        TenantID:    "tenant-123",
        UserID:      "user-456",
        Resource:    "users",
        Event:       "user_login",
        Method:      "POST",
        IP:          "192.168.1.100",
        Environment: "production",
        Data: map[string]interface{}{
            "login_method": "password",
            "user_agent":   "Mozilla/5.0...",
        },
        Meta: map[string]interface{}{
            "request_id": "req-12345",
        },
    }

    err := createAuditLog("your-api-key-1", auditLog)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
    } else {
        fmt.Println("Audit log created successfully")
    }
}
```

### Node.js Client Example

```javascript
const axios = require("axios");

class AuditLogClient {
  constructor(baseURL, apiKey) {
    this.client = axios.create({
      baseURL,
      headers: {
        "X-API-Key": apiKey,
        "Content-Type": "application/json",
      },
    });
  }

  async createAuditLog(auditLog) {
    try {
      const response = await this.client.post("/api/v1/audit", auditLog);
      return response.data;
    } catch (error) {
      throw new Error(
        `Failed to create audit log: ${
          error.response?.data?.message || error.message
        }`
      );
    }
  }

  async getAuditLogs(filters = {}) {
    try {
      const response = await this.client.get("/api/v1/audit", {
        params: filters,
      });
      return response.data;
    } catch (error) {
      throw new Error(
        `Failed to get audit logs: ${
          error.response?.data?.message || error.message
        }`
      );
    }
  }
}

// Usage
const auditClient = new AuditLogClient(
  "http://localhost:9025",
  "your-api-key-1"
);

async function main() {
  // Create audit log
  const auditLog = {
    tenant_id: "tenant-123",
    user_id: "user-456",
    resource: "orders",
    event: "order_created",
    method: "POST",
    ip: "192.168.1.100",
    environment: "production",
    data: {
      order_id: "order-789",
      amount: 99.99,
      currency: "USD",
    },
    meta: {
      request_id: "req-12345",
    },
  };

  try {
    const result = await auditClient.createAuditLog(auditLog);
    console.log("Audit log created:", result.id);

    // Get audit logs
    const logs = await auditClient.getAuditLogs({
      tenant_id: "tenant-123",
      resource: "orders",
      limit: 10,
    });
    console.log(`Found ${logs.total} audit logs`);
  } catch (error) {
    console.error("Error:", error.message);
  }
}

main();
```

### Python Client Example

```python
import requests
import json
from datetime import datetime
from typing import Dict, Any, Optional

class AuditLogClient:
    def __init__(self, base_url: str, api_key: str):
        self.base_url = base_url
        self.headers = {
            'X-API-Key': api_key,
            'Content-Type': 'application/json'
        }

    def create_audit_log(self, audit_log: Dict[str, Any]) -> Dict[str, Any]:
        """Create a new audit log entry."""
        response = requests.post(
            f"{self.base_url}/api/v1/audit",
            headers=self.headers,
            json=audit_log
        )
        response.raise_for_status()
        return response.json()

    def get_audit_logs(self, filters: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """Get audit logs with optional filters."""
        response = requests.get(
            f"{self.base_url}/api/v1/audit",
            headers=self.headers,
            params=filters or {}
        )
        response.raise_for_status()
        return response.json()

    def get_audit_log_stats(self, tenant_id: str, start_date: datetime, end_date: datetime) -> Dict[str, Any]:
        """Get audit log statistics."""
        params = {
            'tenant_id': tenant_id,
            'start_date': start_date.isoformat(),
            'end_date': end_date.isoformat()
        }
        response = requests.get(
            f"{self.base_url}/api/v1/audit/stats",
            headers=self.headers,
            params=params
        )
        response.raise_for_status()
        return response.json()

# Usage example
if __name__ == "__main__":
    client = AuditLogClient('http://localhost:9025', 'your-api-key-1')

    # Create audit log
    audit_log = {
        'tenant_id': 'tenant-123',
        'user_id': 'user-456',
        'resource': 'products',
        'event': 'product_updated',
        'method': 'PUT',
        'ip': '192.168.1.100',
        'environment': 'production',
        'data': {
            'product_id': 'prod-789',
            'updated_fields': ['price', 'description'],
            'old_price': 19.99,
            'new_price': 24.99
        },
        'meta': {
            'request_id': 'req-12345',
            'api_version': 'v2'
        }
    }

    try:
        result = client.create_audit_log(audit_log)
        print(f"Audit log created: {result['id']}")

        # Get recent logs
        logs = client.get_audit_logs({
            'tenant_id': 'tenant-123',
            'limit': 5
        })
        print(f"Found {logs['total']} audit logs")

    except requests.exceptions.HTTPError as e:
        print(f"HTTP Error: {e}")
    except Exception as e:
        print(f"Error: {e}")
```

## Deployment

### Docker Deployment

#### 1. Build Docker Image

```bash
make build-docker
```

#### 2. Run with Docker Compose

```yaml
# docker-compose.yml
version: "3.8"

services:
  audit-log-service:
    build: .
    ports:
      - "9025:9025"
    environment:
      - CONFIG_PATH=/app/configs/config.yaml
    volumes:
      - ./configs:/app/configs
    depends_on:
      - postgres

  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: audit_logs
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: your_password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

```bash
docker-compose up -d
```

### Kubernetes Deployment

#### 1. ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: audit-log-config
data:
  config.yaml: |
    server:
      host: "0.0.0.0"
      port: 9025
    database:
      host: "postgres-service"
      port: 5432
      user: "postgres"
      password: "your_password"
      name: "audit_logs"
    # ... rest of configuration
```

#### 2. Deployment

```yaml
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
            - name: CONFIG_PATH
              value: "/app/config/config.yaml"
          volumeMounts:
            - name: config
              mountPath: /app/config
          livenessProbe:
            httpGet:
              path: /live
              port: 9025
            initialDelaySeconds: 10
            periodSeconds: 30
          readinessProbe:
            httpGet:
              path: /ready
              port: 9025
            initialDelaySeconds: 5
            periodSeconds: 10
      volumes:
        - name: config
          configMap:
            name: audit-log-config
```

#### 3. Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: audit-log-service
spec:
  selector:
    app: audit-log-service
  ports:
    - port: 80
      targetPort: 9025
  type: LoadBalancer
```

## Monitoring & Health Checks

### Health Check Endpoints

- **GET /health**: Basic health status
- **GET /ready**: Database connectivity check
- **GET /live**: Kubernetes liveness probe

### Metrics & Monitoring

The service provides structured logging and can be integrated with monitoring solutions:

#### Prometheus Metrics (if enabled)

- `audit_logs_created_total`: Total number of audit logs created
- `audit_logs_requests_duration_seconds`: Request duration histogram
- `audit_logs_database_connections`: Number of active database connections

#### Log Structure

```json
{
  "timestamp": "2025-06-14T10:30:00Z",
  "level": "info",
  "message": "Audit log created",
  "tenant_id": "tenant-123",
  "user_id": "user-456",
  "resource": "users",
  "event": "user_created",
  "request_id": "req-12345",
  "duration_ms": 45
}
```

## Troubleshooting

### Common Issues

#### 1. Database Connection Failed

**Error**: `failed to connect to database`
**Solution**:

- Check database configuration in `config.yaml`
- Ensure PostgreSQL is running and accessible
- Verify network connectivity and firewall settings

#### 2. Authentication Failed

**Error**: `Unauthorized` (401)
**Solution**:

- Verify JWT token is valid and not expired
- Check API key is included in the configuration
- Ensure correct header format (`Authorization: Bearer TOKEN` or `X-API-Key: KEY`)

#### 3. Validation Errors

**Error**: `Invalid request body` (400)
**Solution**:

- Check required fields are present
- Verify field formats (IP addresses, environment values, HTTP methods)
- Ensure JSON is properly formatted

#### 4. High Memory Usage

**Solution**:

- Adjust database connection pool settings
- Implement pagination for large result sets
- Monitor and tune garbage collection settings

### Debug Mode

Enable debug logging in `config.yaml`:

```yaml
logging:
  level: "debug"
  format: "text"
```

### Database Troubleshooting

#### Check Migration Status

```bash
make migrate-status
```

#### Reset Database (Development Only!)

```bash
make migrate-reset
```

#### Manual Database Connection Test

```bash
psql -h localhost -U postgres -d audit_logs -c "SELECT COUNT(*) FROM audit_logs;"
```

## Best Practices

### Security

1. **Use HTTPS in Production**: Always use TLS/SSL for production deployments
2. **Rotate API Keys**: Regularly rotate API keys and JWT secrets
3. **Implement Rate Limiting**: Use reverse proxy or load balancer for rate limiting
4. **Validate Input**: The service validates input, but additional validation at the client side is recommended
5. **Database Security**: Use strong passwords and restrict database access

### Performance

1. **Use Pagination**: Always use limit/offset for large datasets
2. **Index Strategy**: Ensure proper database indexes for your query patterns
3. **Connection Pooling**: Configure appropriate database connection pool sizes
4. **Batch Operations**: For high-volume scenarios, consider batching audit log creation

### Data Management

1. **Data Retention**: Implement data retention policies based on compliance requirements
2. **Archiving**: Archive old audit logs to reduce database size
3. **Backup Strategy**: Implement regular database backups
4. **Partitioning**: Consider table partitioning for very large datasets

### Integration

1. **Async Processing**: Use queues for non-critical audit log creation
2. **Circuit Breaker**: Implement circuit breaker pattern for resilience
3. **Monitoring**: Monitor service health and performance metrics
4. **Error Handling**: Implement proper error handling and retry logic

### Example Integration with Circuit Breaker (Go)

```go
package main

import (
    "context"
    "time"
    "github.com/sony/gobreaker"
)

type AuditLogService struct {
    client  *http.Client
    breaker *gobreaker.CircuitBreaker
}

func NewAuditLogService() *AuditLogService {
    settings := gobreaker.Settings{
        Name:        "audit-log-service",
        MaxRequests: 3,
        Interval:    10 * time.Second,
        Timeout:     30 * time.Second,
    }

    return &AuditLogService{
        client:  &http.Client{Timeout: 5 * time.Second},
        breaker: gobreaker.NewCircuitBreaker(settings),
    }
}

func (s *AuditLogService) CreateAuditLog(ctx context.Context, auditLog AuditLogRequest) error {
    _, err := s.breaker.Execute(func() (interface{}, error) {
        return nil, s.createAuditLogHTTP(ctx, auditLog)
    })
    return err
}
```

---

## Support

For additional support:

1. Check the [GitHub Issues](https://github.com/your-org/audit-log-service/issues)
2. Review the [API Documentation](./api-docs.md)
3. Contact the development team

## License

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details.
