# Documentation

This directory contains comprehensive documentation for the Audit Log Service, including API integration guides, Postman collection, and deployment instructions.

## Files Overview

### API Documentation

- **[INTEGRATION.md](./INTEGRATION.md)** - Complete integration guide with examples
- **[POSTMAN_SETUP.md](./POSTMAN_SETUP.md)** - Postman collection setup and usage guide

### Postman Collection

- **[audit-log-service.postman_collection.json](./audit-log-service.postman_collection.json)** - Complete API collection for testing
- **[audit-log-service.postman_environment.json](./audit-log-service.postman_environment.json)** - Environment variables for local development

## Quick Links

### For Developers

- [Quick Start Guide](./INTEGRATION.md#quick-start) - Get up and running in minutes
- [Configuration Guide](./INTEGRATION.md#configuration) - Environment-specific setup
- [API Reference](./INTEGRATION.md#api-reference) - Complete endpoint documentation

### For Testers

- [Postman Setup](./POSTMAN_SETUP.md) - Import and configure the collection
- [Test Examples](./POSTMAN_SETUP.md#example-workflow) - Common testing workflows
- [Authentication Methods](./POSTMAN_SETUP.md#authentication-methods) - JWT and API key setup

### For DevOps

- [Docker Deployment](./INTEGRATION.md#docker-deployment) - Containerized deployment
- [Kubernetes Deployment](./INTEGRATION.md#kubernetes-deployment) - Production deployment
- [Monitoring Setup](./INTEGRATION.md#monitoring--health-checks) - Health checks and metrics

## Getting Started

1. **Import Postman Collection**

   ```bash
   # Import both files into Postman:
   # - audit-log-service.postman_collection.json
   # - audit-log-service.postman_environment.json
   ```

2. **Configure Service**

   ```bash
   cp ../configs/config.example.yaml ../configs/config.yaml
   # Edit config.yaml with your settings
   ```

3. **Start Service**

   ```bash
   cd ..
   make dev
   ```

4. **Test with Postman**
   - Select the "Audit Log Service - Local Development" environment
   - Run the "Health Check" request to verify connectivity
   - Test audit log creation and retrieval

## Authentication Quick Reference

### API Key (Recommended for Testing)

```bash
curl -H "X-API-Key: your-api-key-1" \
     -H "Content-Type: application/json" \
     -d '{"tenant_id":"test","user_id":"user1","resource":"users","event":"login","method":"POST","ip":"127.0.0.1","environment":"development"}' \
     http://localhost:9025/api/v1/audit
```

### JWT Token

```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"tenant_id":"test","user_id":"user1","resource":"users","event":"login","method":"POST","ip":"127.0.0.1","environment":"development"}' \
     http://localhost:9025/api/v1/audit
```

## Common Use Cases

### Basic Audit Logging

Log user actions in your application:

```json
{
  "tenant_id": "company-123",
  "user_id": "user-456",
  "resource": "documents",
  "event": "document_downloaded",
  "method": "GET",
  "ip": "192.168.1.100",
  "environment": "production",
  "data": {
    "document_id": "doc-789",
    "document_name": "financial_report.pdf"
  }
}
```

### Security Event Logging

Track security-related events:

```json
{
  "tenant_id": "company-123",
  "user_id": "user-456",
  "resource": "auth",
  "event": "failed_login_attempt",
  "method": "POST",
  "ip": "203.0.113.1",
  "environment": "production",
  "data": {
    "attempted_username": "admin",
    "failure_reason": "invalid_password",
    "attempts_count": 3
  },
  "meta": {
    "user_agent": "Mozilla/5.0...",
    "risk_score": "high"
  }
}
```

### System Integration Logging

Track API calls and integrations:

```json
{
  "tenant_id": "company-123",
  "user_id": "system",
  "resource": "integrations",
  "event": "api_call_completed",
  "method": "POST",
  "ip": "10.0.0.25",
  "environment": "production",
  "data": {
    "external_api": "payment_gateway",
    "endpoint": "/v1/payments",
    "response_time_ms": 245,
    "status_code": 200
  }
}
```

## Support

For questions or issues:

1. Check the [Troubleshooting Guide](./INTEGRATION.md#troubleshooting)
2. Review the [Best Practices](./INTEGRATION.md#best-practices)
3. Contact the development team
4. Create an issue in the project repository

## Contributing

When contributing to the documentation:

1. Update relevant sections in INTEGRATION.md
2. Add new Postman requests to the collection
3. Test all examples before submitting
4. Follow the existing documentation format
