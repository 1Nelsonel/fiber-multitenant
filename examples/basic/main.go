package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"gorm.io/gorm"

	// Import the fiber-multitenant packages
	// Replace with actual import path when published
	"github.com/1Nelsonel/fiber-multitenant/middleware"
	"github.com/1Nelsonel/fiber-multitenant/tenantstore"
)

// User model - will be created in each tenant's schema
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Name      string         `json:"name"`
	Email     string         `gorm:"uniqueIndex" json:"email"`
}

// Post model - will be created in each tenant's schema
type Post struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Title     string         `json:"title"`
	Content   string         `json:"content"`
	UserID    uint           `json:"user_id"`
	User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func main() {
	// Configure tenant store
	// Update this DSN with your PostgreSQL credentials
	dsn := "host=localhost user=postgres password=postgres dbname=multitenant_demo port=5432 sslmode=disable"

	config := tenantstore.DefaultConfig(dsn)

	// Enable auto-migration for models
	config.AutoMigrate = true
	config.Models = []interface{}{&User{}, &Post{}}

	// Create tenant store
	store, err := tenantstore.New(config)
	if err != nil {
		log.Fatalf("Failed to create tenant store: %v", err)
	}
	defer store.Close()

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Multitenant Demo v1.0",
	})

	// Add request logger
	app.Use(logger.New())

	// Add multitenancy middleware (subdomain-based)
	app.Use(middleware.New(middleware.Config{
		Store:    store,
		Resolver: middleware.SubdomainResolver,
		OnTenantResolved: func(c *fiber.Ctx, tenant string) error {
			log.Printf("Resolved tenant: %s", tenant)
			return nil
		},
	}))

	// Routes
	setupRoutes(app)

	// Start server
	log.Println("Starting server on :3000")
	log.Println("Try accessing:")
	log.Println("  - http://tenant1.localhost:3000/users")
	log.Println("  - http://tenant2.localhost:3000/users")
	log.Fatal(app.Listen(":3000"))
}

func setupRoutes(app *fiber.App) {
	api := app.Group("/api")

	// User routes
	api.Get("/users", getUsers)
	api.Get("/users/:id", getUser)
	api.Post("/users", createUser)
	api.Put("/users/:id", updateUser)
	api.Delete("/users/:id", deleteUser)

	// Post routes
	api.Get("/posts", getPosts)
	api.Get("/posts/:id", getPost)
	api.Post("/posts", createPost)
	api.Put("/posts/:id", updatePost)
	api.Delete("/posts/:id", deletePost)

	// Tenant info
	api.Get("/tenant/info", func(c *fiber.Ctx) error {
		tenant := middleware.GetTenant(c)
		return c.JSON(fiber.Map{
			"tenant":    tenant,
			"timestamp": time.Now(),
		})
	})
}

// User Handlers
func getUsers(c *fiber.Ctx) error {
	db := middleware.GetTenantDB(c)

	var users []User
	result := db.Find(&users)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": result.Error.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"tenant": middleware.GetTenant(c),
		"count":  len(users),
		"users":  users,
	})
}

func getUser(c *fiber.Ctx) error {
	db := middleware.GetTenantDB(c)
	id := c.Params("id")

	var user User
	result := db.First(&user, id)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.JSON(user)
}

func createUser(c *fiber.Ctx) error {
	db := middleware.GetTenantDB(c)

	user := new(User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	result := db.Create(user)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": result.Error.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

func updateUser(c *fiber.Ctx) error {
	db := middleware.GetTenantDB(c)
	id := c.Params("id")

	var user User
	result := db.First(&user, id)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	updates := new(User)
	if err := c.BodyParser(updates); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	db.Model(&user).Updates(updates)

	return c.JSON(user)
}

func deleteUser(c *fiber.Ctx) error {
	db := middleware.GetTenantDB(c)
	id := c.Params("id")

	result := db.Delete(&User{}, id)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": result.Error.Error(),
		})
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// Post Handlers
func getPosts(c *fiber.Ctx) error {
	db := middleware.GetTenantDB(c)

	var posts []Post
	result := db.Preload("User").Find(&posts)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": result.Error.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"tenant": middleware.GetTenant(c),
		"count":  len(posts),
		"posts":  posts,
	})
}

func getPost(c *fiber.Ctx) error {
	db := middleware.GetTenantDB(c)
	id := c.Params("id")

	var post Post
	result := db.Preload("User").First(&post, id)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Post not found",
		})
	}

	return c.JSON(post)
}

func createPost(c *fiber.Ctx) error {
	db := middleware.GetTenantDB(c)

	post := new(Post)
	if err := c.BodyParser(post); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	result := db.Create(post)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": result.Error.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(post)
}

func updatePost(c *fiber.Ctx) error {
	db := middleware.GetTenantDB(c)
	id := c.Params("id")

	var post Post
	result := db.First(&post, id)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Post not found",
		})
	}

	updates := new(Post)
	if err := c.BodyParser(updates); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	db.Model(&post).Updates(updates)

	return c.JSON(post)
}

func deletePost(c *fiber.Ctx) error {
	db := middleware.GetTenantDB(c)
	id := c.Params("id")

	result := db.Delete(&Post{}, id)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": result.Error.Error(),
		})
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Post not found",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
