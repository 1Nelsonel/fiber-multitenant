package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/1Nelsonel/fiber-multitenant/middleware"
	"github.com/1Nelsonel/fiber-multitenant/tenantstore"
)

type Product struct {
	ID    uint    `gorm:"primaryKey" json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

func main() {
	// Configure tenant store
	dsn := "host=localhost user=postgres password=postgres dbname=multitenant_demo port=5432 sslmode=disable"
	config := tenantstore.DefaultConfig(dsn)
	config.AutoMigrate = true
	config.Models = []interface{}{&Product{}}

	store, err := tenantstore.New(config)
	if err != nil {
		log.Fatalf("Failed to create tenant store: %v", err)
	}
	defer store.Close()

	app := fiber.New()
	app.Use(logger.New())

	// Use chained resolvers - tries multiple strategies in order
	// 1. First check X-Tenant-ID header
	// 2. Then try subdomain
	// 3. Finally try query parameter
	app.Use(middleware.New(middleware.Config{
		Store: store,
		Resolver: middleware.ChainResolvers(
			middleware.HeaderResolver("X-Tenant-ID"),
			middleware.SubdomainResolver,
			middleware.QueryParamResolver("tenant"),
		),
		OnTenantResolved: func(c *fiber.Ctx, tenant string) error {
			log.Printf("✓ Tenant resolved: %s", tenant)
			return nil
		},
	}))

	// Routes
	app.Get("/products", func(c *fiber.Ctx) error {
		db := middleware.GetTenantDB(c)
		tenant := middleware.GetTenant(c)

		var products []Product
		db.Find(&products)

		return c.JSON(fiber.Map{
			"tenant":   tenant,
			"products": products,
		})
	})

	app.Post("/products", func(c *fiber.Ctx) error {
		db := middleware.GetTenantDB(c)

		product := new(Product)
		if err := c.BodyParser(product); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		db.Create(product)
		return c.Status(fiber.StatusCreated).JSON(product)
	})

	log.Println("Server starting on :3000")
	log.Println("\nTry these requests:")
	log.Println("  1. Header:    curl -H 'X-Tenant-ID: tenant1' http://localhost:3000/products")
	log.Println("  2. Subdomain: curl http://tenant2.localhost:3000/products")
	log.Println("  3. Query:     curl http://localhost:3000/products?tenant=tenant3")

	log.Fatal(app.Listen(":3000"))
}
