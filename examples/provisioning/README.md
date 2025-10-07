# Tenant Provisioning Example

This example demonstrates a complete tenant management system with:

- Tenant provisioning API
- Tenant metadata in master database
- Active/inactive tenant status
- Automatic schema creation and initialization
- Tenant statistics and monitoring

## Features

- **Master Database**: Stores tenant metadata (name, email, plan, status)
- **Tenant Databases**: Each tenant gets isolated schema with auto-migration
- **Admin API**: Create, list, update, and deactivate tenants
- **Validation**: Ensures only active tenants can access their data
- **Default Data**: Creates admin user on tenant provisioning

## Running

```bash
go run main.go
```

## API Endpoints

### Tenant Management (Public)

#### Create Tenant

```bash
curl -X POST http://localhost:3000/api/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "schema": "acme_corp",
    "name": "Acme Corporation",
    "email": "admin@acme.com",
    "plan": "enterprise"
  }'
```

Response:
```json
{
  "message": "Tenant created successfully",
  "tenant": {
    "id": 1,
    "schema": "acme_corp",
    "name": "Acme Corporation",
    "email": "admin@acme.com",
    "plan": "enterprise",
    "active": true
  },
  "admin_user": {
    "id": 1,
    "name": "Acme Corporation Admin",
    "email": "admin@acme.com"
  },
  "access_url": "http://acme_corp.localhost:3000"
}
```

#### List All Tenants

```bash
# All tenants
curl http://localhost:3000/api/tenants

# Active tenants only
curl http://localhost:3000/api/tenants?active=true

# Inactive tenants only
curl http://localhost:3000/api/tenants?active=false
```

#### Get Tenant Details

```bash
curl http://localhost:3000/api/tenants/acme_corp
```

Response includes tenant stats:
```json
{
  "tenant": {
    "id": 1,
    "schema": "acme_corp",
    "name": "Acme Corporation",
    "active": true
  },
  "stats": {
    "users": 5,
    "orders": 23
  }
}
```

#### Update Tenant

```bash
curl -X PUT http://localhost:3000/api/tenants/acme_corp \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corp Inc",
    "plan": "pro"
  }'
```

#### Deactivate Tenant

```bash
curl -X DELETE http://localhost:3000/api/tenants/acme_corp
```

This sets `active = false` and closes database connections. Data is preserved.

### Tenant Operations (Requires Subdomain)

#### Get Tenant Info

```bash
curl http://acme_corp.localhost:3000/info
```

#### Create User

```bash
curl -X POST http://acme_corp.localhost:3000/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@acme.com"
  }'
```

#### List Users

```bash
curl http://acme_corp.localhost:3000/users
```

#### Create Order

```bash
curl -X POST http://acme_corp.localhost:3000/orders \
  -H "Content-Type: application/json" \
  -d '{
    "total": 150.50,
    "status": "pending"
  }'
```

#### List Orders

```bash
curl http://acme_corp.localhost:3000/orders
```

## Complete Workflow

### 1. Create Two Tenants

```bash
# Create Acme Corp
curl -X POST http://localhost:3000/api/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "schema": "acme_corp",
    "name": "Acme Corporation",
    "email": "admin@acme.com",
    "plan": "enterprise"
  }'

# Create Globex
curl -X POST http://localhost:3000/api/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "schema": "globex",
    "name": "Globex Industries",
    "email": "admin@globex.com",
    "plan": "pro"
  }'
```

### 2. Add Users to Each Tenant

```bash
# Acme user
curl -X POST http://acme_corp.localhost:3000/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice Smith", "email": "alice@acme.com"}'

# Globex user
curl -X POST http://globex.localhost:3000/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Bob Jones", "email": "bob@globex.com"}'
```

### 3. Verify Data Isolation

```bash
# Acme users (shows only Acme users)
curl http://acme_corp.localhost:3000/users

# Globex users (shows only Globex users)
curl http://globex.localhost:3000/users
```

### 4. Check Tenant Statistics

```bash
curl http://localhost:3000/api/tenants/acme_corp
curl http://localhost:3000/api/tenants/globex
```

### 5. Deactivate a Tenant

```bash
curl -X DELETE http://localhost:3000/api/tenants/globex
```

Now requests to `http://globex.localhost:3000/users` will fail with:
```json
{
  "error": "Tenant not found or inactive"
}
```

## Database Schema

### Master Database (public schema)

```sql
CREATE TABLE tenants (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    schema VARCHAR UNIQUE NOT NULL,
    name VARCHAR,
    email VARCHAR UNIQUE,
    active BOOLEAN DEFAULT true,
    plan VARCHAR
);
```

### Tenant Schemas (e.g., acme_corp schema)

```sql
CREATE SCHEMA acme_corp;

CREATE TABLE acme_corp.users (
    id SERIAL PRIMARY KEY,
    name VARCHAR,
    email VARCHAR UNIQUE
);

CREATE TABLE acme_corp.orders (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP,
    total DECIMAL,
    status VARCHAR
);
```

## Use Cases

This example is perfect for:

- **SaaS applications** with customer onboarding
- **White-label platforms** where customers get their own space
- **Multi-organization systems** with strict data isolation
- **Testing/staging environments** for different clients

## Security Features

1. **Tenant validation**: Only active tenants can access data
2. **Schema isolation**: PostgreSQL enforces data separation
3. **Connection pooling**: Efficient resource usage
4. **Audit trail**: Created/updated timestamps on all tenants

## Production Considerations

- Add authentication/authorization to tenant management endpoints
- Implement tenant quotas (storage, users, API calls)
- Add tenant-specific rate limiting
- Implement backup strategies per tenant
- Add tenant usage analytics
- Consider tenant-specific feature flags
