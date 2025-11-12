package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/1Nelsonel/fiber-multitenant/middleware"
	"github.com/1Nelsonel/fiber-multitenant/tenantstore"
)

// Tenant metadata stored in master database
type Tenant struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Schema    string    `gorm:"uniqueIndex;not null" json:"schema"`
	Name      string    `json:"name"`
	Email     string    `gorm:"uniqueIndex" json:"email"`
	Active    bool      `gorm:"default:true" json:"active"`
	Plan      string    `json:"plan"` // free, pro, enterprise
}

// Models that live in tenant schemas
type User struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	Name  string `json:"name"`
	Email string `gorm:"uniqueIndex" json:"email"`
}

type Order struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Total     float64   `json:"total"`
	Status    string    `json:"status"`
}

var store *tenantstore.TenantStore

func main() {
	// Configure tenant store
	dsn := "host=localhost user=postgres password=postgres dbname=multitenant_demo port=5432 sslmode=disable"
	config := tenantstore.DefaultConfig(dsn)
	config.AutoMigrate = true
	config.Models = []interface{}{&User{}, &Order{}}

	var err error
	store, err = tenantstore.New(config)
	if err != nil {
		log.Fatalf("Failed to create tenant store: %v", err)
	}
	defer store.Close()

	// Migrate tenant metadata table in master DB
	masterDB := store.GetMasterDB()
	if err := masterDB.AutoMigrate(&Tenant{}); err != nil {
		log.Fatalf("Failed to migrate master DB: %v", err)
	}

	app := fiber.New()
	app.Use(logger.New())

	// Public routes (no tenant required)
	setupPublicRoutes(app)

	// Tenant routes (require tenant resolution)
	tenantRoutes := app.Group("")
	tenantRoutes.Use(middleware.New(middleware.Config{
		Store: store,
		OnTenantResolved: func(c *fiber.Ctx, tenant string) error {
			// Validate tenant is active
			var t Tenant
			if err := masterDB.Where("schema = ? AND active = ?", tenant, true).First(&t).Error; err != nil {
				return fiber.NewError(fiber.StatusForbidden, "Tenant not found or inactive")
			}
			c.Locals("tenant_info", t)
			return nil
		},
	}))
	setupTenantRoutes(tenantRoutes)

	log.Println("Server starting on :3000")
	log.Println("\n=== Tenant Provisioning API ===")
	log.Println("1. Create tenant:      POST   /api/tenants")
	log.Println("2. List tenants:       GET    /api/tenants")
	log.Println("3. Get tenant info:    GET    /api/tenants/:schema")
	log.Println("4. Deactivate tenant:  DELETE /api/tenants/:schema")
	log.Println("\n=== Tenant Operations ===")
	log.Println("Access via subdomain: http://<tenant-schema>.localhost:3000/users")

	log.Fatal(app.Listen(":3000"))
}

func setupPublicRoutes(app *fiber.App) {
	api := app.Group("/api")

	// Tenant management endpoints
	api.Post("/tenants", createTenant)
	api.Get("/tenants", listTenants)
	api.Get("/tenants/:schema", getTenant)
	api.Put("/tenants/:schema", updateTenant)
	api.Delete("/tenants/:schema", deactivateTenant)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"time":   time.Now(),
		})
	})
}

func setupTenantRoutes(group fiber.Router) {
	// User endpoints
	group.Get("/users", getUsers)
	group.Post("/users", createUser)

	// Order endpoints
	group.Get("/orders", getOrders)
	group.Post("/orders", createOrder)

	// Tenant info
	group.Get("/info", func(c *fiber.Ctx) error {
		tenantInfo := c.Locals("tenant_info").(Tenant)
		return c.JSON(fiber.Map{
			"tenant": middleware.GetTenant(c),
			"info":   tenantInfo,
		})
	})
}

// Tenant Management Handlers

func createTenant(c *fiber.Ctx) error {
	masterDB := store.GetMasterDB()

	type CreateTenantRequest struct {
		Schema string `json:"schema"`
		Name   string `json:"name"`
		Email  string `json:"email"`
		Plan   string `json:"plan"`
	}

	req := new(CreateTenantRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate schema name (alphanumeric and underscore only)
	if req.Schema == "" || req.Name == "" || req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Schema, name, and email are required",
		})
	}

	// Create tenant record
	tenant := Tenant{
		Schema: req.Schema,
		Name:   req.Name,
		Email:  req.Email,
		Plan:   req.Plan,
		Active: true,
	}

	if err := masterDB.Create(&tenant).Error; err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Tenant already exists or invalid data",
		})
	}

	// Initialize tenant database (creates schema and migrates)
	tenantDB, err := store.GetTenantDB(c.Context(), req.Schema)
	if err != nil {
		// Rollback tenant creation
		masterDB.Delete(&tenant)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to initialize tenant database: %v", err),
		})
	}

	// Create default admin user
	adminUser := User{
		Name:  fmt.Sprintf("%s Admin", tenant.Name),
		Email: tenant.Email,
	}
	tenantDB.Create(&adminUser)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":    "Tenant created successfully",
		"tenant":     tenant,
		"admin_user": adminUser,
		"access_url": fmt.Sprintf("http://%s.localhost:3000", tenant.Schema),
	})
}

func listTenants(c *fiber.Ctx) error {
	masterDB := store.GetMasterDB()

	var tenants []Tenant
	query := masterDB.Order("created_at DESC")

	// Filter by active status if provided
	if activeStr := c.Query("active"); activeStr != "" {
		if activeStr == "true" {
			query = query.Where("active = ?", true)
		} else {
			query = query.Where("active = ?", false)
		}
	}

	query.Find(&tenants)

	return c.JSON(fiber.Map{
		"count":   len(tenants),
		"tenants": tenants,
	})
}

func getTenant(c *fiber.Ctx) error {
	masterDB := store.GetMasterDB()
	schema := c.Params("schema")

	var tenant Tenant
	if err := masterDB.Where("schema = ?", schema).First(&tenant).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Tenant not found",
		})
	}

	// Get tenant stats
	tenantDB, err := store.GetTenantDB(c.Context(), schema)
	if err == nil {
		var userCount, orderCount int64
		tenantDB.Model(&User{}).Count(&userCount)
		tenantDB.Model(&Order{}).Count(&orderCount)

		return c.JSON(fiber.Map{
			"tenant": tenant,
			"stats": fiber.Map{
				"users":  userCount,
				"orders": orderCount,
			},
		})
	}

	return c.JSON(fiber.Map{
		"tenant": tenant,
	})
}

func updateTenant(c *fiber.Ctx) error {
	masterDB := store.GetMasterDB()
	schema := c.Params("schema")

	var tenant Tenant
	if err := masterDB.Where("schema = ?", schema).First(&tenant).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Tenant not found",
		})
	}

	updates := make(map[string]interface{})
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Don't allow schema changes
	delete(updates, "schema")
	delete(updates, "id")

	masterDB.Model(&tenant).Updates(updates)

	return c.JSON(tenant)
}

func deactivateTenant(c *fiber.Ctx) error {
	masterDB := store.GetMasterDB()
	schema := c.Params("schema")

	var tenant Tenant
	if err := masterDB.Where("schema = ?", schema).First(&tenant).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Tenant not found",
		})
	}

	// Deactivate instead of delete
	masterDB.Model(&tenant).Update("active", false)

	// Optionally close tenant database connection
	if err := store.RemoveTenantDB(schema); err != nil {
		log.Printf("Warning: Failed to remove tenant DB connection: %v", err)
	}

	return c.JSON(fiber.Map{
		"message": "Tenant deactivated successfully",
		"tenant":  tenant,
	})
}

// Tenant Data Handlers

func getUsers(c *fiber.Ctx) error {
	db := middleware.GetTenantDB(c)

	var users []User
	db.Find(&users)

	return c.JSON(fiber.Map{
		"tenant": middleware.GetTenant(c),
		"users":  users,
	})
}

func createUser(c *fiber.Ctx) error {
	db := middleware.GetTenantDB(c)

	user := new(User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := db.Create(user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

func getOrders(c *fiber.Ctx) error {
	db := middleware.GetTenantDB(c)

	var orders []Order
	db.Order("created_at DESC").Find(&orders)

	return c.JSON(fiber.Map{
		"tenant": middleware.GetTenant(c),
		"orders": orders,
	})
}

func createOrder(c *fiber.Ctx) error {
	db := middleware.GetTenantDB(c)

	order := new(Order)
	if err := c.BodyParser(order); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	order.CreatedAt = time.Now()
	if order.Status == "" {
		order.Status = "pending"
	}

	if err := db.Create(order).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(order)
}
