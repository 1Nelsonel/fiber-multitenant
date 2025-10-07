package middleware

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// TenantStore interface defines methods for managing tenant database connections
type TenantStore interface {
	GetTenantDB(ctx context.Context, tenantSchema string) (*gorm.DB, error)
	GetMasterDB() *gorm.DB
}

// Config holds middleware configuration
type Config struct {
	// Resolver function to extract tenant from request
	Resolver TenantResolver

	// TenantStore manages database connections
	Store TenantStore

	// Optional: Custom error handler
	ErrorHandler func(c *fiber.Ctx, err error) error

	// Optional: Skip middleware for certain paths
	Skip func(c *fiber.Ctx) bool

	// ContextKey for storing tenant in fiber context (defaults to "tenant")
	ContextKey string

	// DBContextKey for storing tenant DB in fiber context (defaults to "tenant_db")
	DBContextKey string

	// Optional: Callback after tenant is resolved successfully
	OnTenantResolved func(c *fiber.Ctx, tenant string) error
}

// ConfigDefault is the default config
var ConfigDefault = Config{
	Resolver:     SubdomainResolver,
	ContextKey:   "tenant",
	DBContextKey: "tenant_db",
	ErrorHandler: func(c *fiber.Ctx, err error) error {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "tenant_resolution_failed",
			"message": err.Error(),
		})
	},
}

// New creates a new tenant middleware handler
func New(config ...Config) fiber.Handler {
	// Set default config
	cfg := ConfigDefault

	// Override config if provided
	if len(config) > 0 {
		cfg = config[0]

		// Set defaults for optional fields
		if cfg.Resolver == nil {
			cfg.Resolver = ConfigDefault.Resolver
		}
		if cfg.ContextKey == "" {
			cfg.ContextKey = ConfigDefault.ContextKey
		}
		if cfg.DBContextKey == "" {
			cfg.DBContextKey = ConfigDefault.DBContextKey
		}
		if cfg.ErrorHandler == nil {
			cfg.ErrorHandler = ConfigDefault.ErrorHandler
		}
	}

	// Validate required config
	if cfg.Store == nil {
		panic("TenantStore is required")
	}

	return func(c *fiber.Ctx) error {
		// Skip middleware if Skip function returns true
		if cfg.Skip != nil && cfg.Skip(c) {
			return c.Next()
		}

		// Resolve tenant from request
		tenant, err := cfg.Resolver(c)
		if err != nil {
			return cfg.ErrorHandler(c, err)
		}

		// Store tenant in context
		c.Locals(cfg.ContextKey, tenant)

		// Get tenant database connection
		tenantDB, err := cfg.Store.GetTenantDB(c.Context(), tenant)
		if err != nil {
			return cfg.ErrorHandler(c, err)
		}

		// Store tenant DB in context
		c.Locals(cfg.DBContextKey, tenantDB)

		// Call optional callback
		if cfg.OnTenantResolved != nil {
			if err := cfg.OnTenantResolved(c, tenant); err != nil {
				return cfg.ErrorHandler(c, err)
			}
		}

		return c.Next()
	}
}

// GetTenant retrieves the tenant identifier from fiber context
func GetTenant(c *fiber.Ctx, contextKey ...string) string {
	key := "tenant"
	if len(contextKey) > 0 && contextKey[0] != "" {
		key = contextKey[0]
	}

	tenant, ok := c.Locals(key).(string)
	if !ok {
		return ""
	}
	return tenant
}

// GetTenantDB retrieves the tenant database from fiber context
func GetTenantDB(c *fiber.Ctx, contextKey ...string) *gorm.DB {
	key := "tenant_db"
	if len(contextKey) > 0 && contextKey[0] != "" {
		key = contextKey[0]
	}

	db, ok := c.Locals(key).(*gorm.DB)
	if !ok {
		return nil
	}
	return db
}

// MustGetTenant retrieves tenant and panics if not found (use in routes after middleware)
func MustGetTenant(c *fiber.Ctx, contextKey ...string) string {
	tenant := GetTenant(c, contextKey...)
	if tenant == "" {
		panic("tenant not found in context")
	}
	return tenant
}

// MustGetTenantDB retrieves tenant DB and panics if not found (use in routes after middleware)
func MustGetTenantDB(c *fiber.Ctx, contextKey ...string) *gorm.DB {
	db := GetTenantDB(c, contextKey...)
	if db == nil {
		panic("tenant database not found in context")
	}
	return db
}
