package middleware

import (
	"context"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Mock TenantStore for testing
type mockTenantStore struct {
	tenants map[string]*gorm.DB
}

func (m *mockTenantStore) GetTenantDB(ctx context.Context, tenantSchema string) (*gorm.DB, error) {
	if db, ok := m.tenants[tenantSchema]; ok {
		return db, nil
	}
	// Return a mock DB (in real tests you'd use a test database)
	return &gorm.DB{}, nil
}

func (m *mockTenantStore) GetMasterDB() *gorm.DB {
	return &gorm.DB{}
}

func TestSubdomainResolver(t *testing.T) {
	tests := []struct {
		name       string
		host       string
		wantTenant string
		wantError  bool
	}{
		{
			name:       "Valid subdomain",
			host:       "tenant1.example.com",
			wantTenant: "tenant1",
			wantError:  false,
		},
		{
			name:       "Valid subdomain with port",
			host:       "tenant2.localhost:3000",
			wantTenant: "tenant2",
			wantError:  false,
		},
		{
			name:       "WWW subdomain (should fail)",
			host:       "www.example.com",
			wantTenant: "",
			wantError:  true,
		},
		{
			name:       "API subdomain (should fail)",
			host:       "api.example.com",
			wantTenant: "",
			wantError:  true,
		},
		{
			name:       "No subdomain",
			host:       "example.com",
			wantTenant: "",
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			app.Get("/test", func(c *fiber.Ctx) error {
				tenant, err := SubdomainResolver(c)

				if tt.wantError {
					if err == nil {
						t.Fatal("Expected error but got none")
					}
					return c.SendStatus(fiber.StatusBadRequest)
				}

				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if tenant != tt.wantTenant {
					t.Fatalf("Expected tenant '%s', got '%s'", tt.wantTenant, tenant)
				}

				return c.SendString(tenant)
			})

			req := httptest.NewRequest("GET", "http://"+tt.host+"/test", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to test: %v", err)
			}

			if tt.wantError {
				if resp.StatusCode != fiber.StatusBadRequest {
					t.Fatalf("Expected status 400, got %d", resp.StatusCode)
				}
			} else {
				if resp.StatusCode != fiber.StatusOK {
					t.Fatalf("Expected status 200, got %d", resp.StatusCode)
				}

				body, _ := io.ReadAll(resp.Body)
				if string(body) != tt.wantTenant {
					t.Fatalf("Expected body '%s', got '%s'", tt.wantTenant, string(body))
				}
			}
		})
	}
}

func TestHeaderResolver(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		resolver := HeaderResolver("X-Tenant-ID")
		tenant, err := resolver(c)

		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}

		return c.SendString(tenant)
	})

	// Test with header
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Tenant-ID", "tenant1")
	resp, _ := app.Test(req)

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "tenant1" {
		t.Fatalf("Expected 'tenant1', got '%s'", string(body))
	}

	// Test without header
	req = httptest.NewRequest("GET", "/test", nil)
	resp, _ = app.Test(req)

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestPathPrefixResolver(t *testing.T) {
	app := fiber.New()

	app.Get("/:tenant/test", func(c *fiber.Ctx) error {
		tenant, err := PathPrefixResolver(c)

		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}

		return c.SendString(tenant)
	})

	req := httptest.NewRequest("GET", "/tenant1/test", nil)
	resp, _ := app.Test(req)

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "tenant1" {
		t.Fatalf("Expected 'tenant1', got '%s'", string(body))
	}
}

func TestQueryParamResolver(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		resolver := QueryParamResolver("tenant")
		tenant, err := resolver(c)

		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}

		return c.SendString(tenant)
	})

	req := httptest.NewRequest("GET", "/test?tenant=tenant1", nil)
	resp, _ := app.Test(req)

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "tenant1" {
		t.Fatalf("Expected 'tenant1', got '%s'", string(body))
	}
}

func TestChainResolvers(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		resolver := ChainResolvers(
			HeaderResolver("X-Tenant-ID"),
			SubdomainResolver,
			QueryParamResolver("tenant"),
		)

		tenant, err := resolver(c)

		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}

		return c.SendString(tenant)
	})

	// Test header (priority 1)
	req := httptest.NewRequest("GET", "http://tenant2.localhost:3000/test?tenant=tenant3", nil)
	req.Header.Set("X-Tenant-ID", "tenant1")
	resp, _ := app.Test(req)

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "tenant1" {
		t.Fatalf("Expected 'tenant1' from header, got '%s'", string(body))
	}

	// Test subdomain (priority 2, no header)
	req = httptest.NewRequest("GET", "http://tenant2.localhost:3000/test?tenant=tenant3", nil)
	resp, _ = app.Test(req)

	body, _ = io.ReadAll(resp.Body)
	if string(body) != "tenant2" {
		t.Fatalf("Expected 'tenant2' from subdomain, got '%s'", string(body))
	}

	// Test query param (priority 3, no header or subdomain)
	req = httptest.NewRequest("GET", "http://localhost:3000/test?tenant=tenant3", nil)
	resp, _ = app.Test(req)

	body, _ = io.ReadAll(resp.Body)
	if string(body) != "tenant3" {
		t.Fatalf("Expected 'tenant3' from query, got '%s'", string(body))
	}
}

func TestMiddlewareNew(t *testing.T) {
	mockStore := &mockTenantStore{
		tenants: make(map[string]*gorm.DB),
	}

	app := fiber.New()

	app.Use(New(Config{
		Store:    mockStore,
		Resolver: SubdomainResolver,
	}))

	app.Get("/test", func(c *fiber.Ctx) error {
		tenant := GetTenant(c)
		db := GetTenantDB(c)

		if tenant == "" {
			return c.Status(fiber.StatusInternalServerError).SendString("No tenant")
		}

		if db == nil {
			return c.Status(fiber.StatusInternalServerError).SendString("No DB")
		}

		return c.JSON(fiber.Map{
			"tenant": tenant,
			"hasDB":  db != nil,
		})
	})

	req := httptest.NewRequest("GET", "http://tenant1.localhost:3000/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d: %s", resp.StatusCode, string(body))
	}
}

func TestGetTenant(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		c.Locals("tenant", "test-tenant")
		tenant := GetTenant(c)

		if tenant != "test-tenant" {
			t.Fatalf("Expected 'test-tenant', got '%s'", tenant)
		}

		return c.SendString(tenant)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	app.Test(req)
}

func TestGetTenantCustomKey(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		c.Locals("custom_tenant", "test-tenant")
		tenant := GetTenant(c, "custom_tenant")

		if tenant != "test-tenant" {
			t.Fatalf("Expected 'test-tenant', got '%s'", tenant)
		}

		return c.SendString(tenant)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	app.Test(req)
}

func TestMustGetTenant(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		c.Locals("tenant", "test-tenant")

		defer func() {
			if r := recover(); r != nil {
				t.Fatal("Should not panic when tenant exists")
			}
		}()

		tenant := MustGetTenant(c)
		if tenant != "test-tenant" {
			t.Fatalf("Expected 'test-tenant', got '%s'", tenant)
		}

		return c.SendString(tenant)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	app.Test(req)
}

func TestMustGetTenantPanic(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("Should panic when tenant doesn't exist")
			}
		}()

		MustGetTenant(c)
		return nil
	})

	req := httptest.NewRequest("GET", "/test", nil)
	app.Test(req)
}

func TestSkipMiddleware(t *testing.T) {
	mockStore := &mockTenantStore{
		tenants: make(map[string]*gorm.DB),
	}

	app := fiber.New()

	app.Use(New(Config{
		Store:    mockStore,
		Resolver: SubdomainResolver,
		Skip: func(c *fiber.Ctx) bool {
			return c.Path() == "/health"
		},
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		// Should not have tenant in context
		tenant := GetTenant(c)
		if tenant != "" {
			t.Fatal("Expected no tenant for skipped path")
		}
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/health", nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}
}
