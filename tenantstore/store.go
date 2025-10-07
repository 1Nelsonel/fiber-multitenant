package tenantstore

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TenantStore manages database connections for multiple tenants with schema isolation
type TenantStore struct {
	masterDB        *gorm.DB
	tenantDBs       map[string]*gorm.DB
	mu              sync.RWMutex
	config          *Config
	healthCheckDone map[string]bool
	healthMu        sync.Mutex
}

// Config holds configuration for tenant store
type Config struct {
	MasterDSN           string
	GetTenantDSN        func(tenantSchema string) string
	AutoMigrate         bool
	Models              []interface{}
	ConnectionTimeout   time.Duration
	HealthCheckInterval time.Duration
	Logger              logger.Interface
}

// DefaultConfig returns a config with sensible defaults
func DefaultConfig(masterDSN string) *Config {
	return &Config{
		MasterDSN: masterDSN,
		GetTenantDSN: func(tenantSchema string) string {
			return masterDSN + fmt.Sprintf("&search_path=%s,public", tenantSchema)
		},
		AutoMigrate:         true,
		Models:              []interface{}{},
		ConnectionTimeout:   10 * time.Second,
		HealthCheckInterval: 5 * time.Minute,
		Logger:              logger.Default.LogMode(logger.Silent),
	}
}

// New creates a new TenantStore instance
func New(config *Config) (*TenantStore, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Open master database connection
	masterDB, err := gorm.Open(postgres.Open(config.MasterDSN), &gorm.Config{
		Logger: config.Logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to master database: %w", err)
	}

	store := &TenantStore{
		masterDB:        masterDB,
		tenantDBs:       make(map[string]*gorm.DB),
		config:          config,
		healthCheckDone: make(map[string]bool),
	}

	return store, nil
}

// GetMasterDB returns the master database connection
func (s *TenantStore) GetMasterDB() *gorm.DB {
	return s.masterDB
}

// GetTenantDB returns a database connection for the specified tenant schema
// It creates the connection if it doesn't exist and performs health checks
func (s *TenantStore) GetTenantDB(ctx context.Context, tenantSchema string) (*gorm.DB, error) {
	if tenantSchema == "" {
		return nil, fmt.Errorf("tenant schema cannot be empty")
	}

	// Check if connection exists
	s.mu.RLock()
	db, exists := s.tenantDBs[tenantSchema]
	s.mu.RUnlock()

	if exists {
		// Perform periodic health check
		s.healthCheckWithInterval(ctx, tenantSchema, db)
		return db, nil
	}

	// Create new connection
	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after acquiring write lock
	if db, exists := s.tenantDBs[tenantSchema]; exists {
		return db, nil
	}

	// Create schema if it doesn't exist
	if err := s.ensureSchema(ctx, tenantSchema); err != nil {
		return nil, fmt.Errorf("failed to ensure schema: %w", err)
	}

	// Get tenant-specific DSN with search_path
	tenantDSN := s.config.GetTenantDSN(tenantSchema)

	// Open tenant database connection
	tenantDB, err := gorm.Open(postgres.Open(tenantDSN), &gorm.Config{
		Logger: s.config.Logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to tenant database: %w", err)
	}

	// Auto-migrate models if enabled
	if s.config.AutoMigrate && len(s.config.Models) > 0 {
		if err := tenantDB.AutoMigrate(s.config.Models...); err != nil {
			return nil, fmt.Errorf("failed to auto-migrate models: %w", err)
		}
	}

	// Store connection
	s.tenantDBs[tenantSchema] = tenantDB
	s.healthCheckDone[tenantSchema] = false

	return tenantDB, nil
}

// ensureSchema creates the schema if it doesn't exist
func (s *TenantStore) ensureSchema(ctx context.Context, schemaName string) error {
	createSchemaSQL := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)
	if err := s.masterDB.WithContext(ctx).Exec(createSchemaSQL).Error; err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}
	return nil
}

// healthCheckWithInterval performs health check with interval control
func (s *TenantStore) healthCheckWithInterval(ctx context.Context, tenantSchema string, db *gorm.DB) {
	s.healthMu.Lock()
	defer s.healthMu.Unlock()

	if s.healthCheckDone[tenantSchema] {
		return
	}

	// Perform health check
	sqlDB, err := db.DB()
	if err == nil {
		if err := sqlDB.PingContext(ctx); err == nil {
			s.healthCheckDone[tenantSchema] = true

			// Reset health check flag after interval
			go func() {
				time.Sleep(s.config.HealthCheckInterval)
				s.healthMu.Lock()
				s.healthCheckDone[tenantSchema] = false
				s.healthMu.Unlock()
			}()
		}
	}
}

// RemoveTenantDB closes and removes a tenant database connection
func (s *TenantStore) RemoveTenantDB(tenantSchema string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	db, exists := s.tenantDBs[tenantSchema]
	if !exists {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	delete(s.tenantDBs, tenantSchema)
	delete(s.healthCheckDone, tenantSchema)

	return nil
}

// Close closes all database connections
func (s *TenantStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var errs []error

	// Close tenant connections
	for schema, db := range s.tenantDBs {
		sqlDB, err := db.DB()
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get DB for %s: %w", schema, err))
			continue
		}

		if err := sqlDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close connection for %s: %w", schema, err))
		}
	}

	// Close master connection
	masterSQLDB, err := s.masterDB.DB()
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to get master DB: %w", err))
	} else {
		if err := masterSQLDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close master connection: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing connections: %v", errs)
	}

	return nil
}

// GetAllTenantSchemas returns a list of all tenant schemas currently in the store
func (s *TenantStore) GetAllTenantSchemas() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	schemas := make([]string, 0, len(s.tenantDBs))
	for schema := range s.tenantDBs {
		schemas = append(schemas, schema)
	}
	return schemas
}
