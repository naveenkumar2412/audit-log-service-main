# Postman Collection Setup Guide

## Overview

This guide will help you set up and use the Audit Log Service Postman collection for testing and integration purposes.

## Files Included

- `audit-log-service.postman_collection.json` - Main API collection
- `audit-log-service.postman_environment.json` - Environment variables
- `INTEGRATION.md` - Comprehensive integration documentation

## Quick Setup

### 1. Import Collection and Environment

1. Open Postman
2. Click "Import" button
3. Import both files:
   - `audit-log-service.postman_collection.json`
   - `audit-log-service.postman_environment.json`

### 2. Configure Environment Variables

1. Select the "Audit Log Service - Local Development" environment
2. Update the following variables if needed:
   - `base_url`: Update if service runs on different host/port
   - `api_key`: Update with actual API key from your config.yaml
   - `tenant_id`: Update with your tenant ID
   - `user_id`: Update with your user ID

### 3. Start the Service

Before testing, ensure the Audit Log Service is running:

```bash
# Start in development mode
make dev

# Or start in production mode
make build && ./bin/audit-log-service
```

### 4. Test the Collection

#### Health Checks (No Authentication Required)

1. **Health Check** - Basic service status
2. **Readiness Check** - Database connectivity
3. **Liveness Check** - Kubernetes probe

#### Audit Log Operations (Authentication Required)

1. **Create Audit Log (JWT Auth)** - Create with JWT token
2. **Create Audit Log (API Key Auth)** - Create with API key
3. **List Audit Logs - Basic** - Simple listing
4. **List Audit Logs - Filtered** - Advanced filtering
5. **Get Audit Log by ID** - Retrieve specific log
6. **Get Audit Log Statistics** - Analytics data

#### Error Examples

1. **Unauthorized Request** - Missing authentication
2. **Invalid Request Body** - Validation errors

## Authentication Methods

### API Key Authentication

The easiest method for testing:

1. Use the `X-API-Key` header
2. Default key: `your-api-key-1` (update in environment variables)
3. Configure keys in `configs/config.yaml`:
   ```yaml
   auth:
     api_keys:
       - "your-api-key-1"
       - "your-api-key-2"
   ```

### JWT Authentication

For more advanced scenarios:

1. Use the `Authorization: Bearer <token>` header
2. Generate JWT tokens with appropriate claims
3. Configure JWT secret in `configs/config.yaml`:
   ```yaml
   auth:
     jwt_secret: "your-super-secret-jwt-key"
   ```

## Example Workflow

### 1. Test Service Health

Start with the health check endpoints to ensure the service is running:

- Run "Health Check" request
- Run "Readiness Check" to verify database connectivity

### 2. Create Audit Logs

Test audit log creation with different authentication methods:

- Use "Create Audit Log (API Key Auth)" for simplicity
- Modify the request body with your own data

### 3. Query Audit Logs

Test different query patterns:

- Use "List Audit Logs - Basic" for simple pagination
- Use "List Audit Logs - Filtered" to test filtering capabilities
- Adjust query parameters as needed

### 4. Retrieve Specific Data

- Use "Get Audit Log by ID" with an ID from previous responses
- Use "Get Audit Log Statistics" for analytics

## Request Body Examples

### Basic Audit Log Creation

```json
{
  "tenant_id": "tenant-123",
  "user_id": "user-456",
  "resource": "users",
  "event": "user_created",
  "method": "POST",
  "ip": "192.168.1.100",
  "environment": "production",
  "data": {
    "user_email": "john.doe@example.com",
    "user_role": "admin"
  },
  "meta": {
    "request_id": "req-12345"
  }
}
```

### Advanced Audit Log with Rich Data

```json
{
  "tenant_id": "tenant-456",
  "user_id": "user-789",
  "resource": "orders",
  "event": "order_payment_processed",
  "method": "POST",
  "ip": "10.0.0.25",
  "environment": "production",
  "data": {
    "order_id": "order-12345",
    "payment_method": "credit_card",
    "amount": 99.99,
    "currency": "USD",
    "payment_gateway": "stripe",
    "transaction_id": "txn_abc123"
  },
  "meta": {
    "request_id": "req-67890",
    "session_id": "sess-11111",
    "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
    "api_version": "v2",
    "client_version": "1.2.3"
  }
}
```

## Common Query Parameters

### Filtering

- `tenant_id` - Filter by tenant
- `user_id` - Filter by user
- `resource` - Filter by resource type (users, orders, products, etc.)
- `event` - Filter by event type (created, updated, deleted, etc.)
- `method` - Filter by HTTP method (GET, POST, PUT, DELETE, PATCH)
- `environment` - Filter by environment (development, staging, production)

### Date Filtering

- `start_date` - Start date in RFC3339 format (e.g., `2025-06-01T00:00:00Z`)
- `end_date` - End date in RFC3339 format (e.g., `2025-06-14T23:59:59Z`)

### Pagination

- `limit` - Number of results (default: 50, max: 1000)
- `offset` - Number of results to skip (default: 0)

## Environment-Specific Configurations

### Development Environment

```json
{
  "base_url": "http://localhost:9025",
  "api_key": "dev-api-key-123"
}
```

### Staging Environment

```json
{
  "base_url": "https://audit-staging.yourcompany.com",
  "api_key": "staging-api-key-456"
}
```

### Production Environment

```json
{
  "base_url": "https://audit.yourcompany.com",
  "api_key": "prod-api-key-789"
}
```

## Automated Testing Scripts

The collection includes pre-request and test scripts:

### Pre-request Scripts

- Generate dynamic request IDs
- Set timestamps
- Configure environment-specific variables

### Test Scripts

- Validate response times
- Check response formats
- Store response data for subsequent requests
- Validate status codes

## Troubleshooting

### Common Issues

#### 1. Connection Refused

- **Error**: `connect ECONNREFUSED 127.0.0.1:9025`
- **Solution**: Ensure the service is running on the correct port

#### 2. Authentication Failed

- **Error**: `401 Unauthorized`
- **Solution**: Check API key in environment variables matches config.yaml

#### 3. Database Connection Error

- **Error**: Health check fails with database error
- **Solution**: Ensure PostgreSQL is running and accessible

#### 4. Validation Errors

- **Error**: `400 Bad Request` with validation messages
- **Solution**: Check request body format and required fields

### Debug Tips

1. Check the Postman Console for detailed request/response logs
2. Use the service's debug logging mode
3. Verify environment variable values
4. Test with curl commands for comparison

## Advanced Usage

### Bulk Testing

Use Postman's Collection Runner to:

1. Run all requests in sequence
2. Test with different data sets
3. Performance testing with iterations

### Integration with Newman

Run the collection from command line:

```bash
newman run audit-log-service.postman_collection.json \
  -e audit-log-service.postman_environment.json \
  --reporters cli,html \
  --reporter-html-export audit-test-report.html
```

### CI/CD Integration

Add to your pipeline:

```yaml
# .github/workflows/api-tests.yml
- name: Run API Tests
  run: |
    npm install -g newman
    newman run docs/audit-log-service.postman_collection.json \
      -e docs/audit-log-service.postman_environment.json \
      --env-var base_url=${{ secrets.API_BASE_URL }} \
      --env-var api_key=${{ secrets.API_KEY }}
```

---

## Support

For issues with the Postman collection or API integration, please refer to the main [INTEGRATION.md](./INTEGRATION.md) documentation or contact the development team.
