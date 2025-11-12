# Fiber Multitenant

[![Go Version](https://img.shields.io/github/go-mod/go-version/1Nelsonel/fiber-multitenant)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/1Nelsonel/fiber-multitenant)](https://goreportcard.com/report/github.com/1Nelsonel/fiber-multitenant)
[![CI](https://github.com/1Nelsonel/fiber-multitenant/workflows/CI/badge.svg)](https://github.com/1Nelsonel/fiber-multitenant/actions)
[![GoDoc](https://pkg.go.dev/badge/github.com/1Nelsonel/fiber-multitenant)](https://pkg.go.dev/github.com/1Nelsonel/fiber-multitenant)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/1Nelsonel/fiber-multitenant)](https://github.com/1Nelsonel/fiber-multitenant/releases)

A production-ready multitenancy solution for [Go Fiber](https://gofiber.io/) with PostgreSQL schema-based isolation using [GORM](https://gorm.io/).

## Features

- 🏢 **Schema-based multitenancy** - Each tenant gets isolated PostgreSQL schema
- 🔄 **Connection pooling** - Cached tenant database connections with health checks
- 🌐 **Flexible tenant resolution** - Subdomain, header, path prefix, query param, or custom
- 🚀 **Auto-migration** - Automatic schema creation and model migration
- 🔒 **Secure isolation** - DSN-level `search_path` prevents cross-tenant data leaks
- ⚡ **Zero configuration** - Sensible defaults with full customization
- 🧪 **Chained resolvers** - Try multiple resolution strategies
- 🎯 **Context helpers** - Easy tenant and DB access in handlers

## Installation

```bash
go get github.com/1Nelsonel/fiber-multitenant
```

## Quick Start

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/1Nelsonel/fiber-multitenant/middleware"
    "github.com/1Nelsonel/fiber-multitenant/tenantstore"
)

type User struct {
    ID   uint   `gorm:"primaryKey"`
    Name string
}

func main() {
    // Configure tenant store
    config := tenantstore.DefaultConfig(
        "host=localhost user=postgres password=postgres dbname=myapp port=5432 sslmode=disable",
    )
    config.Models = []interface{}{&User{}}

    // Create tenant store
    store, err := tenantstore.New(config)
    if err != nil {
        panic(err)
    }
    defer store.Close()

    // Create Fiber app
    app := fiber.New()

    // Add multitenancy middleware (subdomain-based by default)
    app.Use(middleware.New(middleware.Config{
        Store: store,
    }))

    // Define routes
    app.Get("/users", func(c *fiber.Ctx) error {
        // Get tenant DB from context
        db := middleware.GetTenantDB(c)

        var users []User
        db.Find(&users)

        return c.JSON(users)
    })

    app.Listen(":3000")
}
```

Now requests to `tenant1.localhost:3000/users` will query the `tenant1` schema, while `tenant2.localhost:3000/users` queries the `tenant2` schema!

## Tenant Resolution Strategies

### Subdomain (Default)

Extracts tenant from subdomain:

```go
// tenant1.example.com → "tenant1"
app.Use(middleware.New(middleware.Config{
    Store:    store,
    Resolver: middleware.SubdomainResolver,
}))
```

### Header

Extracts tenant from custom header:

```go
// Header: X-Tenant-ID: tenant1
app.Use(middleware.New(middleware.Config{
    Store:    store,
    Resolver: middleware.HeaderResolver("X-Tenant-ID"),
}))
```

### Path Prefix

Extracts tenant from URL path:

```go
// /tenant1/users → "tenant1"
app.Use(middleware.New(middleware.Config{
    Store:    store,
    Resolver: middleware.PathPrefixResolver,
}))
```

### Query Parameter

Extracts tenant from query parameter:

```go
// /users?tenant=tenant1 → "tenant1"
app.Use(middleware.New(middleware.Config{
    Store:    store,
    Resolver: middleware.QueryParamResolver("tenant"),
}))
```

### Chained Resolvers

Try multiple strategies in order:

```go
app.Use(middleware.New(middleware.Config{
    Store: store,
    Resolver: middleware.ChainResolvers(
        middleware.HeaderResolver("X-Tenant-ID"),
        middleware.SubdomainResolver,
        middleware.QueryParamResolver("tenant"),
    ),
}))
```

### Custom Resolver

Implement your own logic:

```go
app.Use(middleware.New(middleware.Config{
    Store: store,
    Resolver: func(c *fiber.Ctx) (string, error) {
        // Your custom logic
        return "my-tenant", nil
    },
}))
```

## Advanced Configuration

### Auto-Migration

Automatically create schemas and migrate models:

```go
config := tenantstore.DefaultConfig(dsn)
config.AutoMigrate = true
config.Models = []interface{}{&User{}, &Post{}, &Comment{}}
```

### Custom DSN Builder

Control how tenant DSN is generated:

```go
config := tenantstore.DefaultConfig(dsn)
config.GetTenantDSN = func(tenantSchema string) string {
    return fmt.Sprintf(
        "host=localhost user=%s password=secret dbname=myapp search_path=%s,public",
        tenantSchema, // Use schema name as username
        tenantSchema,
    )
}
```

### Skip Middleware for Certain Paths

```go
app.Use(middleware.New(middleware.Config{
    Store: store,
    Skip: func(c *fiber.Ctx) bool {
        return c.Path() == "/health" || c.Path() == "/metrics"
    },
}))
```

### Custom Error Handler

```go
app.Use(middleware.New(middleware.Config{
    Store: store,
    ErrorHandler: func(c *fiber.Ctx, err error) error {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "Invalid tenant",
        })
    },
}))
```

### Post-Resolution Callback

Execute logic after tenant resolution:

```go
app.Use(middleware.New(middleware.Config{
    Store: store,
    OnTenantResolved: func(c *fiber.Ctx, tenant string) error {
        // Log tenant access
        log.Printf("Tenant %s accessed", tenant)

        // Validate tenant status
        if !isTenantActive(tenant) {
            return fiber.NewError(fiber.StatusForbidden, "Tenant inactive")
        }

        return nil
    },
}))
```

## Accessing Tenant Context

### In Handlers

```go
app.Get("/users", func(c *fiber.Ctx) error {
    // Get tenant name
    tenant := middleware.GetTenant(c)

    // Get tenant database
    db := middleware.GetTenantDB(c)

    var users []User
    db.Find(&users)

    return c.JSON(fiber.Map{
        "tenant": tenant,
        "users":  users,
    })
})
```

### Must Helpers (Panic if Not Found)

```go
app.Get("/users", func(c *fiber.Ctx) error {
    // Panics if tenant not in context
    tenant := middleware.MustGetTenant(c)
    db := middleware.MustGetTenantDB(c)

    // ...
})
```

## Master Database Access

For operations that need master database access (e.g., tenant provisioning):

```go
app.Post("/tenants", func(c *fiber.Ctx) error {
    // Get master DB (not tenant-specific)
    masterDB := store.GetMasterDB()

    // Create new tenant record
    tenant := Tenant{Name: "New Tenant"}
    masterDB.Create(&tenant)

    // The schema will be created automatically on first access
    return c.JSON(tenant)
})
```

## Database Operations

### Query Tenant Data

```go
app.Get("/posts", func(c *fiber.Ctx) error {
    db := middleware.GetTenantDB(c)

    var posts []Post
    db.Where("published = ?", true).Find(&posts)

    return c.JSON(posts)
})
```

### Create Records

```go
app.Post("/posts", func(c *fiber.Ctx) error {
    db := middleware.GetTenantDB(c)

    post := new(Post)
    if err := c.BodyParser(post); err != nil {
        return err
    }

    db.Create(post)

    return c.JSON(post)
})
```

### Transactions

```go
app.Post("/transfer", func(c *fiber.Ctx) error {
    db := middleware.GetTenantDB(c)

    return db.Transaction(func(tx *gorm.DB) error {
        // Your transaction logic
        return nil
    })
})
```

## Production Considerations

### Connection Pooling

Configure PostgreSQL connection pool:

```go
config := tenantstore.DefaultConfig(dsn)
store, _ := tenantstore.New(config)

sqlDB, _ := store.GetMasterDB().DB()
sqlDB.SetMaxIdleConns(10)
sqlDB.SetMaxOpenConns(100)
sqlDB.SetConnMaxLifetime(time.Hour)
```

### Health Checks

The store automatically performs periodic health checks on tenant connections:

```go
config := tenantstore.DefaultConfig(dsn)
config.HealthCheckInterval = 5 * time.Minute // Default is 5 minutes
```

### Logging

Enable GORM logging for debugging:

```go
import "gorm.io/gorm/logger"

config := tenantstore.DefaultConfig(dsn)
config.Logger = logger.Default.LogMode(logger.Info)
```

### Removing Inactive Tenants

Close connections for tenants that are no longer active:

```go
if err := store.RemoveTenantDB("old-tenant"); err != nil {
    log.Printf("Failed to remove tenant: %v", err)
}
```

### List All Active Tenants

```go
schemas := store.GetAllTenantSchemas()
log.Printf("Active tenants: %v", schemas)
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Fiber App                            │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Tenant Middleware                          │
│  1. Resolve tenant (subdomain/header/path/query/custom)     │
│  2. Get/create tenant DB connection                          │
│  3. Store tenant & DB in context                             │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Tenant Store                            │
│  - Manages connection pool per tenant                        │
│  - Creates schemas on-demand                                 │
│  - Performs health checks                                    │
│  - Sets search_path for isolation                            │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    PostgreSQL Database                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ tenant1      │  │ tenant2      │  │ tenant3      │     │
│  │ (schema)     │  │ (schema)     │  │ (schema)     │     │
│  │ - users      │  │ - users      │  │ - users      │     │
│  │ - posts      │  │ - posts      │  │ - posts      │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└─────────────────────────────────────────────────────────────┘
```

## Schema Isolation

Each tenant gets its own PostgreSQL schema with isolated tables:

```sql
-- Tenant 1 schema
CREATE SCHEMA tenant1;
CREATE TABLE tenant1.users (...);
CREATE TABLE tenant1.posts (...);

-- Tenant 2 schema
CREATE SCHEMA tenant2;
CREATE TABLE tenant2.users (...);
CREATE TABLE tenant2.posts (...);
```

The `search_path` in the connection DSN ensures queries only access the tenant's schema:

```go
// Connection DSN for tenant1
"host=localhost dbname=myapp search_path=tenant1,public"
```

This provides:
- **Strong isolation**: Queries can't accidentally access other tenants' data
- **Efficient**: All tenants share same database, minimizing overhead
- **Simple migrations**: Use standard GORM migrations per tenant
- **Backup flexibility**: Can backup individual schemas or entire database

## Testing

```go
func TestTenantResolution(t *testing.T) {
    app := fiber.New()

    // Setup test store
    config := tenantstore.DefaultConfig("postgresql://...")
    store, _ := tenantstore.New(config)
    defer store.Close()

    // Add middleware
    app.Use(middleware.New(middleware.Config{
        Store: store,
    }))

    app.Get("/tenant", func(c *fiber.Ctx) error {
        return c.SendString(middleware.GetTenant(c))
    })

    // Test subdomain resolution
    req := httptest.NewRequest("GET", "http://tenant1.example.com/tenant", nil)
    resp, _ := app.Test(req)

    body, _ := io.ReadAll(resp.Body)
    assert.Equal(t, "tenant1", string(body))
}
```

## Examples

See the [examples](./examples) directory for complete working examples:

- [Basic Setup](./examples/basic) - Simple subdomain-based multitenancy
- [Header-based](./examples/header) - Using custom headers for tenant resolution
- [Chained Resolvers](./examples/chained) - Multiple resolution strategies
- [Tenant Provisioning](./examples/provisioning) - API for creating/managing tenants

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Credits

Built with:
- [Fiber](https://gofiber.io/) - Express-inspired web framework
- [GORM](https://gorm.io/) - The fantastic ORM library
- Inspired by multitenancy implementations in the Go community

## Related Projects

- [fiber](https://github.com/gofiber/fiber) - Web framework
- [gorm](https://github.com/go-gorm/gorm) - ORM library
- [pgx](https://github.com/jackc/pgx) - PostgreSQL driver

---

**Need help?** Open an issue or discussion on GitHub!
