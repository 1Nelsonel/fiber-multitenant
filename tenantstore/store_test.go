package tenantstore

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"gorm.io/gorm"
)

type TestModel struct {
	ID   uint   `gorm:"primaryKey"`
	Name string
}

func getTestDSN() string {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=multitenant_test port=5432 sslmode=disable"
	}
	return dsn
}

func TestNew(t *testing.T) {
	config := DefaultConfig(getTestDSN())

	store, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	if store == nil {
		t.Fatal("Expected store to be non-nil")
	}

	if store.masterDB == nil {
		t.Fatal("Expected master DB to be non-nil")
	}
}

func TestNewWithInvalidConfig(t *testing.T) {
	_, err := New(nil)
	if err == nil {
		t.Fatal("Expected error when config is nil")
	}
}

func TestGetMasterDB(t *testing.T) {
	config := DefaultConfig(getTestDSN())
	store, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	masterDB := store.GetMasterDB()
	if masterDB == nil {
		t.Fatal("Expected master DB to be non-nil")
	}
}

func TestGetTenantDB(t *testing.T) {
	config := DefaultConfig(getTestDSN())
	config.AutoMigrate = true
	config.Models = []interface{}{&TestModel{}}

	store, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	tenantSchema := fmt.Sprintf("test_tenant_%d", time.Now().Unix())

	// Clean up after test
	defer func() {
		store.masterDB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", tenantSchema))
	}()

	// Get tenant DB
	tenantDB, err := store.GetTenantDB(ctx, tenantSchema)
	if err != nil {
		t.Fatalf("Failed to get tenant DB: %v", err)
	}

	if tenantDB == nil {
		t.Fatal("Expected tenant DB to be non-nil")
	}

	// Verify we can use the database
	var count int64
	tenantDB.Model(&TestModel{}).Count(&count)
	// Should succeed even if count is 0
}

func TestGetTenantDBEmptySchema(t *testing.T) {
	config := DefaultConfig(getTestDSN())
	store, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	_, err = store.GetTenantDB(ctx, "")
	if err == nil {
		t.Fatal("Expected error when tenant schema is empty")
	}
}

func TestGetTenantDBCaching(t *testing.T) {
	config := DefaultConfig(getTestDSN())
	store, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	tenantSchema := fmt.Sprintf("test_tenant_%d", time.Now().Unix())

	// Clean up after test
	defer func() {
		store.masterDB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", tenantSchema))
	}()

	// Get tenant DB first time
	db1, err := store.GetTenantDB(ctx, tenantSchema)
	if err != nil {
		t.Fatalf("Failed to get tenant DB: %v", err)
	}

	// Get tenant DB second time (should be cached)
	db2, err := store.GetTenantDB(ctx, tenantSchema)
	if err != nil {
		t.Fatalf("Failed to get cached tenant DB: %v", err)
	}

	// Should return the same instance
	if fmt.Sprintf("%p", db1) != fmt.Sprintf("%p", db2) {
		t.Fatal("Expected same DB instance for cached connection")
	}
}

func TestRemoveTenantDB(t *testing.T) {
	config := DefaultConfig(getTestDSN())
	store, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	tenantSchema := fmt.Sprintf("test_tenant_%d", time.Now().Unix())

	// Clean up after test
	defer func() {
		store.masterDB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", tenantSchema))
	}()

	// Create tenant DB
	_, err = store.GetTenantDB(ctx, tenantSchema)
	if err != nil {
		t.Fatalf("Failed to get tenant DB: %v", err)
	}

	// Remove tenant DB
	err = store.RemoveTenantDB(tenantSchema)
	if err != nil {
		t.Fatalf("Failed to remove tenant DB: %v", err)
	}

	// Verify it's removed from cache
	store.mu.RLock()
	_, exists := store.tenantDBs[tenantSchema]
	store.mu.RUnlock()

	if exists {
		t.Fatal("Expected tenant DB to be removed from cache")
	}
}

func TestGetAllTenantSchemas(t *testing.T) {
	config := DefaultConfig(getTestDSN())
	store, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	tenant1 := fmt.Sprintf("test_tenant_1_%d", time.Now().Unix())
	tenant2 := fmt.Sprintf("test_tenant_2_%d", time.Now().Unix())

	// Clean up after test
	defer func() {
		store.masterDB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", tenant1))
		store.masterDB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", tenant2))
	}()

	// Initially should be empty
	schemas := store.GetAllTenantSchemas()
	if len(schemas) != 0 {
		t.Fatalf("Expected 0 schemas, got %d", len(schemas))
	}

	// Create two tenant DBs
	_, err = store.GetTenantDB(ctx, tenant1)
	if err != nil {
		t.Fatalf("Failed to get tenant1 DB: %v", err)
	}

	_, err = store.GetTenantDB(ctx, tenant2)
	if err != nil {
		t.Fatalf("Failed to get tenant2 DB: %v", err)
	}

	// Should have 2 schemas
	schemas = store.GetAllTenantSchemas()
	if len(schemas) != 2 {
		t.Fatalf("Expected 2 schemas, got %d", len(schemas))
	}
}

func TestAutoMigration(t *testing.T) {
	config := DefaultConfig(getTestDSN())
	config.AutoMigrate = true
	config.Models = []interface{}{&TestModel{}}

	store, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	tenantSchema := fmt.Sprintf("test_tenant_%d", time.Now().Unix())

	// Clean up after test
	defer func() {
		store.masterDB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", tenantSchema))
	}()

	// Get tenant DB (should auto-migrate)
	tenantDB, err := store.GetTenantDB(ctx, tenantSchema)
	if err != nil {
		t.Fatalf("Failed to get tenant DB: %v", err)
	}

	// Verify table exists by creating a record
	testModel := TestModel{Name: "Test"}
	result := tenantDB.Create(&testModel)
	if result.Error != nil {
		t.Fatalf("Failed to create test model (table may not exist): %v", result.Error)
	}

	// Verify we can query it back
	var retrieved TestModel
	tenantDB.First(&retrieved, testModel.ID)
	if retrieved.Name != "Test" {
		t.Fatalf("Expected name 'Test', got '%s'", retrieved.Name)
	}
}

func TestSchemaIsolation(t *testing.T) {
	config := DefaultConfig(getTestDSN())
	config.AutoMigrate = true
	config.Models = []interface{}{&TestModel{}}

	store, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	tenant1 := fmt.Sprintf("test_tenant_1_%d", time.Now().Unix())
	tenant2 := fmt.Sprintf("test_tenant_2_%d", time.Now().Unix())

	// Clean up after test
	defer func() {
		store.masterDB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", tenant1))
		store.masterDB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", tenant2))
	}()

	// Get tenant DBs
	db1, err := store.GetTenantDB(ctx, tenant1)
	if err != nil {
		t.Fatalf("Failed to get tenant1 DB: %v", err)
	}

	db2, err := store.GetTenantDB(ctx, tenant2)
	if err != nil {
		t.Fatalf("Failed to get tenant2 DB: %v", err)
	}

	// Create record in tenant1
	model1 := TestModel{Name: "Tenant1 Data"}
	db1.Create(&model1)

	// Create record in tenant2
	model2 := TestModel{Name: "Tenant2 Data"}
	db2.Create(&model2)

	// Verify tenant1 only sees its data
	var tenant1Models []TestModel
	db1.Find(&tenant1Models)
	if len(tenant1Models) != 1 {
		t.Fatalf("Expected 1 record in tenant1, got %d", len(tenant1Models))
	}
	if tenant1Models[0].Name != "Tenant1 Data" {
		t.Fatalf("Expected 'Tenant1 Data', got '%s'", tenant1Models[0].Name)
	}

	// Verify tenant2 only sees its data
	var tenant2Models []TestModel
	db2.Find(&tenant2Models)
	if len(tenant2Models) != 1 {
		t.Fatalf("Expected 1 record in tenant2, got %d", len(tenant2Models))
	}
	if tenant2Models[0].Name != "Tenant2 Data" {
		t.Fatalf("Expected 'Tenant2 Data', got '%s'", tenant2Models[0].Name)
	}
}

func TestClose(t *testing.T) {
	config := DefaultConfig(getTestDSN())
	store, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	ctx := context.Background()
	tenantSchema := fmt.Sprintf("test_tenant_%d", time.Now().Unix())

	// Clean up after test
	defer func() {
		store.masterDB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", tenantSchema))
	}()

	// Create tenant DB
	_, err = store.GetTenantDB(ctx, tenantSchema)
	if err != nil {
		t.Fatalf("Failed to get tenant DB: %v", err)
	}

	// Close store
	err = store.Close()
	if err != nil {
		t.Fatalf("Failed to close store: %v", err)
	}

	// Verify connections are closed by trying to ping
	masterSQLDB, _ := store.masterDB.DB()
	if err := masterSQLDB.Ping(); err == nil {
		t.Fatal("Expected master DB to be closed")
	}
}
