# Basic Multitenant Example

This example demonstrates the simplest setup for a multitenant Fiber application using subdomain-based tenant resolution.

## Features

- Subdomain-based tenant resolution
- Auto-migration of models
- Complete CRUD operations for Users and Posts
- Isolated data per tenant schema

## Running the Example

1. Install dependencies:

```bash
go mod download
```

2. Start PostgreSQL:

```bash
docker run --name postgres-multitenant \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=multitenant_demo \
  -p 5432:5432 \
  -d postgres:15
```

3. Run the application:

```bash
go run main.go
```

4. Test with different subdomains:

```bash
# Create user for tenant1
curl -X POST http://tenant1.localhost:3000/api/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "email": "alice@tenant1.com"}'

# Create user for tenant2
curl -X POST http://tenant2.localhost:3000/api/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Bob", "email": "bob@tenant2.com"}'

# Get users for tenant1 (only shows Alice)
curl http://tenant1.localhost:3000/api/users

# Get users for tenant2 (only shows Bob)
curl http://tenant2.localhost:3000/api/users
```

## What's Happening?

1. When you access `tenant1.localhost:3000`, the middleware extracts `tenant1` from the subdomain
2. It creates/retrieves a database connection with `search_path=tenant1`
3. On first access, the `tenant1` schema is created automatically
4. Models (User, Post) are migrated to the `tenant1` schema
5. All queries execute within the `tenant1` schema
6. Data is completely isolated from other tenants

## Database Structure

After running the example, your PostgreSQL will have:

```
multitenant_demo (database)
  ├── tenant1 (schema)
  │   ├── users (table)
  │   └── posts (table)
  ├── tenant2 (schema)
  │   ├── users (table)
  │   └── posts (table)
  └── public (default schema)
```

## Testing with Hosts File

If you want to test with real domains instead of localhost, add to `/etc/hosts`:

```
127.0.0.1 tenant1.myapp.local
127.0.0.1 tenant2.myapp.local
```

Then access:
- http://tenant1.myapp.local:3000
- http://tenant2.myapp.local:3000
