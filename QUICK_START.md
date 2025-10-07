# Quick Start Guide

Get up and running with Fiber Multitenant in 5 minutes!

## 1. Install

```bash
go get github.com/1Nelsonel/fiber-multitenant
```

## 2. Copy-Paste Example

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/1Nelsonel/fiber-multitenant/middleware"
    "github.com/1Nelsonel/fiber-multitenant/tenantstore"
)

type User struct {
    ID    uint   `gorm:"primaryKey" json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    // 1. Configure store
    config := tenantstore.DefaultConfig(
        "host=localhost user=postgres password=postgres dbname=myapp port=5432 sslmode=disable",
    )
    config.AutoMigrate = true
    config.Models = []interface{}{&User{}}

    // 2. Create store
    store, _ := tenantstore.New(config)
    defer store.Close()

    // 3. Create app
    app := fiber.New()

    // 4. Add middleware
    app.Use(middleware.New(middleware.Config{
        Store: store,
    }))

    // 5. Add routes
    app.Get("/users", func(c *fiber.Ctx) error {
        db := middleware.GetTenantDB(c)

        var users []User
        db.Find(&users)

        return c.JSON(users)
    })

    app.Post("/users", func(c *fiber.Ctx) error {
        db := middleware.GetTenantDB(c)

        user := new(User)
        c.BodyParser(user)
        db.Create(user)

        return c.JSON(user)
    })

    // 6. Start server
    app.Listen(":3000")
}
```

## 3. Run

```bash
go run main.go
```

## 4. Test

```bash
# Create user for tenant1
curl -X POST http://tenant1.localhost:3000/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "email": "alice@example.com"}'

# Create user for tenant2
curl -X POST http://tenant2.localhost:3000/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Bob", "email": "bob@example.com"}'

# Get users for tenant1 (only Alice)
curl http://tenant1.localhost:3000/users

# Get users for tenant2 (only Bob)
curl http://tenant2.localhost:3000/users
```

## That's It! 🎉

You now have fully isolated multitenancy!

## Common Configurations

### Use Header Instead of Subdomain

```go
app.Use(middleware.New(middleware.Config{
    Store:    store,
    Resolver: middleware.HeaderResolver("X-Tenant-ID"),
}))
```

Test:
```bash
curl http://localhost:3000/users -H "X-Tenant-ID: tenant1"
```

### Use Multiple Resolvers

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

Test:
```bash
# Any of these work:
curl http://localhost:3000/users -H "X-Tenant-ID: tenant1"
curl http://tenant1.localhost:3000/users
curl http://localhost:3000/users?tenant=tenant1
```

### Skip Certain Paths

```go
app.Use(middleware.New(middleware.Config{
    Store: store,
    Skip: func(c *fiber.Ctx) bool {
        return c.Path() == "/health" || c.Path() == "/metrics"
    },
}))
```

### Add Validation

```go
app.Use(middleware.New(middleware.Config{
    Store: store,
    OnTenantResolved: func(c *fiber.Ctx, tenant string) error {
        // Check if tenant is active
        if !isTenantActive(tenant) {
            return fiber.NewError(fiber.StatusForbidden, "Tenant inactive")
        }
        return nil
    },
}))
```

## Next Steps

- 📖 Read the full [README.md](./README.md)
- 💡 Check out [examples/](./examples/)
- 🚀 See [PUBLISHING.md](./PUBLISHING.md) to publish your own fork
- 🐛 Report issues on GitHub

## Need Help?

- Full docs: [README.md](./README.md)
- Examples: [examples/](./examples/)
- Issues: GitHub Issues
