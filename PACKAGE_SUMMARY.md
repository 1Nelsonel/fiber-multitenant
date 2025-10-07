# Fiber Multitenant - Package Summary

## Overview

This package provides a production-ready multitenancy solution for Go Fiber applications using PostgreSQL schema-based isolation with GORM. It was extracted from a real production application and designed to be shared with the Go community.

## What Problem Does This Solve?

Building multitenant SaaS applications in Go Fiber requires:
1. **Tenant isolation** - Keeping each customer's data separate
2. **Connection management** - Efficiently handling multiple database connections
3. **Tenant resolution** - Identifying which tenant is making each request
4. **Schema management** - Creating and migrating database schemas per tenant

Most Go/Fiber developers implement this from scratch. This package provides a tested, reusable solution.

## Package Structure

```
fiber-multitenant/
├── tenantstore/           # Database connection and schema management
│   ├── store.go          # Main TenantStore implementation
│   └── store_test.go     # Comprehensive tests
├── middleware/            # Fiber middleware for tenant resolution
│   ├── middleware.go     # Main middleware
│   ├── resolver.go       # Various tenant resolvers
│   └── middleware_test.go # Tests
├── examples/              # Working examples
│   ├── basic/            # Simple subdomain-based setup
│   ├── chained/          # Multiple resolution strategies
│   └── provisioning/     # Complete tenant management API
├── README.md             # Comprehensive documentation
├── CONTRIBUTING.md       # Contribution guidelines
├── PUBLISHING.md         # How to publish to GitHub
├── LICENSE               # MIT License
└── go.mod                # Go module definition
```

## Key Features

### 1. Schema-Based Multitenancy

Each tenant gets an isolated PostgreSQL schema:

```
database
  ├── tenant1 (schema)
  │   ├── users
  │   └── orders
  ├── tenant2 (schema)
  │   ├── users
  │   └── orders
```

**Benefits:**
- Strong data isolation
- Efficient resource usage (single database)
- Easy backup/restore per tenant
- Standard SQL works unchanged

### 2. Connection Pooling

- Caches database connections per tenant
- Automatic health checks
- Connection reuse for performance
- Lazy initialization (connects on first access)

### 3. Flexible Tenant Resolution

Multiple strategies out of the box:
- **Subdomain**: `tenant1.example.com` → `tenant1`
- **Header**: `X-Tenant-ID: tenant1`
- **Path**: `/tenant1/api/users`
- **Query**: `/api/users?tenant=tenant1`
- **Chained**: Try multiple strategies in order
- **Custom**: Your own logic

### 4. Auto-Migration

Automatically creates schemas and migrates GORM models on first access.

### 5. Zero Configuration

Works with sensible defaults, fully customizable when needed.

## Core Components

### TenantStore

Manages database connections and schemas:

```go
type TenantStore struct {
    // Caches connections per tenant
    // Handles schema creation
    // Performs health checks
}
```

**Key Methods:**
- `GetTenantDB(ctx, schema)` - Get/create tenant connection
- `GetMasterDB()` - Get master database connection
- `RemoveTenantDB(schema)` - Close tenant connection
- `GetAllTenantSchemas()` - List active tenants
- `Close()` - Close all connections

### Middleware

Fiber middleware that:
1. Resolves tenant from request
2. Gets database connection
3. Stores both in Fiber context
4. Makes available to route handlers

```go
app.Use(middleware.New(middleware.Config{
    Store: store,
    Resolver: middleware.SubdomainResolver,
}))
```

### Resolvers

Functions that extract tenant identifier from HTTP request:

```go
type TenantResolver func(c *fiber.Ctx) (string, error)
```

Built-in resolvers:
- `SubdomainResolver`
- `HeaderResolver(headerName)`
- `PathPrefixResolver`
- `QueryParamResolver(paramName)`
- `ChainResolvers(resolvers...)`
- `CustomResolver(fn)`

## Usage Pattern

### 1. Setup (Once at Startup)

```go
// Configure
config := tenantstore.DefaultConfig(databaseDSN)
config.AutoMigrate = true
config.Models = []interface{}{&User{}, &Post{}}

// Create store
store, err := tenantstore.New(config)
defer store.Close()

// Add middleware
app.Use(middleware.New(middleware.Config{
    Store: store,
}))
```

### 2. Use in Handlers

```go
app.Get("/users", func(c *fiber.Ctx) error {
    // Get tenant info from context
    tenant := middleware.GetTenant(c)
    db := middleware.GetTenantDB(c)

    // Query tenant's data
    var users []User
    db.Find(&users)

    return c.JSON(users)
})
```

## Example Applications

### Basic Example

Simplest setup with subdomain resolution:
- Run: `cd examples/basic && go run main.go`
- Access: `http://tenant1.localhost:3000/api/users`

### Chained Resolvers

Multiple resolution strategies:
- Header first, then subdomain, then query param
- Perfect for APIs + web + testing

### Provisioning Example

Complete tenant management system:
- Create/list/update/deactivate tenants
- Tenant metadata in master DB
- Statistics and monitoring
- Production-ready architecture

## Security & Isolation

### How Data Isolation Works

1. **DSN-level search_path**: Each connection has `search_path=tenantX,public`
2. **PostgreSQL enforcement**: Database enforces schema boundaries
3. **No application logic needed**: Can't accidentally query wrong tenant

```sql
-- Tenant1 connection
SET search_path = tenant1, public;
SELECT * FROM users;  -- Only sees tenant1.users

-- Tenant2 connection
SET search_path = tenant2, public;
SELECT * FROM users;  -- Only sees tenant2.users
```

### Security Features

- **Schema-level isolation** via PostgreSQL
- **Connection-level separation** (separate connections per tenant)
- **Validation hooks** (OnTenantResolved callback)
- **Error handling** for invalid tenants
- **Optional tenant status checks**

## Performance Considerations

### Connection Caching

First request creates connection, subsequent requests reuse:

```
Request 1: tenant1 → Create connection (slow)
Request 2: tenant1 → Reuse connection (fast)
Request 3: tenant1 → Reuse connection (fast)
```

### Health Checks

Periodic pings ensure connections are alive:
- Default: Every 5 minutes
- Configurable via `HealthCheckInterval`
- Non-blocking (doesn't slow down requests)

### Resource Usage

**Per Tenant:**
- 1 cached database connection
- 1 PostgreSQL schema (minimal overhead)

**Scalability:**
- Handles hundreds of tenants efficiently
- Can remove inactive tenant connections
- PostgreSQL handles schema isolation efficiently

## Production Deployment

### Recommended Setup

1. **Connection pooling**:
```go
sqlDB, _ := store.GetMasterDB().DB()
sqlDB.SetMaxIdleConns(10)
sqlDB.SetMaxOpenConns(100)
sqlDB.SetConnMaxLifetime(time.Hour)
```

2. **Logging**:
```go
import "gorm.io/gorm/logger"
config.Logger = logger.Default.LogMode(logger.Info)
```

3. **Monitoring**:
```go
config.OnTenantResolved = func(c *fiber.Ctx, tenant string) error {
    metrics.RecordTenantAccess(tenant)
    return nil
}
```

### Migration Strategy

**New Tenants:**
- Schema created automatically on first access
- Models migrated via `AutoMigrate`

**Existing Tenants:**
- Run migrations manually or via admin endpoint
- Use GORM's `AutoMigrate` or custom SQL

### Backup Strategy

**Per-Tenant Backup:**
```bash
pg_dump -n tenant1 dbname > tenant1_backup.sql
```

**Full Backup:**
```bash
pg_dump dbname > full_backup.sql
```

## Testing

Comprehensive test coverage included:

**TenantStore Tests:**
- Connection creation and caching
- Schema isolation
- Auto-migration
- Health checks
- Concurrent access

**Middleware Tests:**
- All resolver types
- Chained resolvers
- Context storage
- Error handling
- Skip functionality

Run tests:
```bash
go test ./...
```

## Comparison with Alternatives

### vs. Database-per-Tenant

**Advantages:**
- Lower overhead (single DB instance)
- Easier management
- Faster tenant provisioning
- Simpler backups

**Trade-offs:**
- All tenants in one database
- Shared connection pool

### vs. Row-level Multitenancy

**Advantages:**
- Stronger isolation
- Easier to understand
- Simpler queries (no tenant_id everywhere)
- Lower bug risk

**Trade-offs:**
- Slightly more complex migrations
- Can't easily share data between tenants

## Future Enhancements

Possible additions for future versions:

1. **Metrics**: Built-in Prometheus metrics
2. **Tenant Limits**: Quota enforcement
3. **Rate Limiting**: Per-tenant rate limits
4. **Caching**: Per-tenant cache management
5. **Read Replicas**: Tenant-aware read/write splitting
6. **Sharding**: Distribute tenants across databases

## Why Open Source This?

1. **Common need**: Many Go developers need this
2. **Proven solution**: Extracted from production app
3. **Save time**: Hours of development for others
4. **Community**: Get feedback, improvements, and tests
5. **Portfolio**: Demonstrate Go expertise

## How to Use This Package

### For Learning

Study the code to understand:
- Schema-based multitenancy patterns
- Fiber middleware development
- GORM connection management
- PostgreSQL schema isolation

### For Projects

Drop into your project:
```bash
go get github.com/1Nelsonel/fiber-multitenant
```

Minimal setup:
```go
store, _ := tenantstore.New(tenantstore.DefaultConfig(dsn))
app.Use(middleware.New(middleware.Config{Store: store}))
```

### For Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md):
- Report bugs
- Add features
- Improve docs
- Share examples

## License

MIT License - Use freely in commercial and personal projects.

## Credits

Extracted from production SaaS application built with:
- Go Fiber (web framework)
- GORM (ORM)
- PostgreSQL (database)

## Support

- **Issues**: GitHub Issues
- **Discussions**: GitHub Discussions
- **Examples**: See `examples/` directory
- **Docs**: See README.md

---

**Ready to publish?** See [PUBLISHING.md](./PUBLISHING.md) for step-by-step guide!
