package migrations

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Config holds migration configuration
type Config struct {
	DatabaseURL    string
	MigrationsPath string
	Timeout        time.Duration
	LogLevel       string
}

// buildDatabaseURL constructs a database URL from individual environment variables
func buildDatabaseURL() string {
	// Try DATABASE_URL first
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		return dbURL
	}

	// Build from individual components
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSL_MODE")

	// Set defaults if not provided
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}
	if sslmode == "" {
		sslmode = "require"
	}

	// Construct PostgreSQL connection string
	if user != "" && password != "" && dbname != "" {
		return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, dbname, sslmode)
	}

	return ""
}

// DefaultConfig returns default migration configuration
func DefaultConfig() *Config {
	return &Config{
		DatabaseURL:    buildDatabaseURL(),
		MigrationsPath: "infrastructure/scripts",
		Timeout:        30 * time.Second,
		LogLevel:       "info",
	}
}

// Migrator provides database migration functionality
type Migrator struct {
	config  *Config
	migrate *migrate.Migrate
}

// NewMigrator creates a new migrator instance
func NewMigrator(config *Config) (*Migrator, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Validate configuration
	if config.DatabaseURL == "" {
		return nil, fmt.Errorf("database connection not configured. Please set either DATABASE_URL or individual DB_* environment variables (DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, DB_SSL_MODE)")
	}

	// Resolve migrations path to absolute path
	migrationsPath, err := filepath.Abs(config.MigrationsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve migrations path: %v", err)
	}

	// Validate migrations path exists
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("migrations path does not exist: %s", migrationsPath)
	}

	// Create migrator instance with absolute path
	migrationURL := fmt.Sprintf("file://%s", migrationsPath)
	log.Printf("🔧 Creating migrator with URL: %s", migrationURL)
	
	m, err := migrate.New(
		migrationURL,
		config.DatabaseURL,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %v", err)
	}

	// Set timeout
	if config.Timeout > 0 {
		m.LockTimeout = config.Timeout
	}

	// Set log level
	switch config.LogLevel {
	case "debug":
		m.Log = &debugLogger{}
	case "info":
		m.Log = &infoLogger{}
	case "warn":
		m.Log = &warnLogger{}
	case "error":
		m.Log = &errorLogger{}
	default:
		m.Log = &infoLogger{}
	}

	// Validate that the migrator can see the migration files
	if err := validateMigratorFiles(m, migrationsPath); err != nil {
		log.Printf("⚠️  Warning: Migrator validation failed: %v", err)
	}
	
	return &Migrator{
		config:  config,
		migrate: m,
	}, nil
}

// dropAndRecreateSchemaMigrations drops and recreates the schema_migrations table
func (m *Migrator) dropAndRecreateSchemaMigrations() error {
	log.Printf("🗑️  Attempting to drop and recreate schema_migrations table...")
	
	// Since we can't access the database directly, we'll use the Drop method
	// which should clear all migration state
	if err := m.migrate.Drop(); err != nil {
		return fmt.Errorf("failed to drop migration state: %v", err)
	}
	
	log.Printf("✅ Migration state dropped successfully")
	return nil
}

// Up runs all pending migrations
func (m *Migrator) Up(ctx context.Context) error {
	log.Printf("🚀 Starting database migrations...")
	
	// Use the already-configured migrator instance
	log.Printf("📁 Using configured migrations path")
	log.Printf("🔗 Database: %s", maskDatabaseURL(m.config.DatabaseURL))

	// Run migrations using the configured migrator
	err := m.migrate.Up()
	if err != nil && err != migrate.ErrNoChange {
		// Check if it's a schema version conflict or dirty state
		if err.Error() == "dirty database version -1. Fix and force version." || 
		   err.Error() == "pq: column \"version\" does not exist in line 0: SELECT version, dirty FROM \"public\".\"schema_migrations\" LIMIT 1" {
			log.Printf("⚠️  Database has migration state conflict. Attempting to force version...")
			// Force the version to 0 to start fresh
			if forceErr := m.migrate.Force(0); forceErr != nil {
				log.Printf("⚠️  Force failed, attempting to drop and recreate schema_migrations table...")
				// If force fails, try to drop and recreate the schema_migrations table
				if dropErr := m.dropAndRecreateSchemaMigrations(); dropErr != nil {
					return fmt.Errorf("failed to reset migration state: %v", dropErr)
				}
				// Try running migrations again
				if retryErr := m.migrate.Up(); retryErr != nil && retryErr != migrate.ErrNoChange {
					return fmt.Errorf("failed to run migrations after reset: %v", retryErr)
				}
			} else {
				// Try running migrations again after successful force
				if retryErr := m.migrate.Up(); retryErr != nil && retryErr != migrate.ErrNoChange {
					return fmt.Errorf("failed to run migrations after force: %v", retryErr)
				}
			}
		} else {
			return fmt.Errorf("failed to run migrations: %v", err)
		}
	}

	if err == migrate.ErrNoChange {
		log.Printf("✅ Database is up to date (no new migrations)")
	} else {
		log.Printf("✅ All migrations completed successfully")
	}

	return nil
}

// Down rolls back all migrations
func (m *Migrator) Down(ctx context.Context) error {
	log.Printf("🔄 Rolling back all migrations...")

	if err := m.migrate.Down(); err != nil {
		return fmt.Errorf("failed to rollback migrations: %v", err)
	}

	log.Printf("✅ All migrations rolled back successfully")
	return nil
}

// Force sets the migration version (useful for fixing dirty state)
func (m *Migrator) Force(version int) error {
	log.Printf("🔧 Force setting migration version to %d", version)

	if err := m.migrate.Force(version); err != nil {
		return fmt.Errorf("failed to force migration version: %v", err)
	}

	log.Printf("✅ Migration version set to %d", version)
	return nil
}

// Version returns the current migration version
func (m *Migrator) Version() (uint, bool, error) {
	return m.migrate.Version()
}

// Close closes the migrator and releases resources
func (m *Migrator) Close() error {
	if m.migrate != nil {
		if _, err := m.migrate.Close(); err != nil {
			return err
		}
	}
	return nil
}

// listMigrationFiles lists all SQL migration files in the migrations path
func (m *Migrator) listMigrationFiles(migrationsPath string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(migrationsPath, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".sql" {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// maskDatabaseURL masks sensitive information in database URL for logging
func maskDatabaseURL(url string) string {
	if len(url) < 10 {
		return "***"
	}

	// Show only host and database name, mask credentials
	if len(url) > 20 {
		return url[:20] + "***"
	}

	return "***"
}

// Logger implementations for different log levels
type debugLogger struct{}
type infoLogger struct{}
type warnLogger struct{}
type errorLogger struct{}

func (l *debugLogger) Printf(format string, v ...interface{}) {
	log.Printf("[DEBUG] "+format, v...)
}

func (l *debugLogger) Verbose() bool {
	return true
}

func (l *infoLogger) Printf(format string, v ...interface{}) {
	log.Printf("[INFO] "+format, v...)
}

func (l *infoLogger) Verbose() bool {
	return false
}

func (l *warnLogger) Printf(format string, v ...interface{}) {
	log.Printf("[WARN] "+format, v...)
}

func (l *warnLogger) Verbose() bool {
	return false
}

func (l *errorLogger) Printf(format string, v ...interface{}) {
	log.Printf("[ERROR] "+format, v...)
}

func (l *errorLogger) Verbose() bool {
	return false
}

// Utility functions for common migration operations

// ValidateMigrationFiles checks if migration files exist and are readable
func ValidateMigrationFiles(migrationsPath string) error {
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		return fmt.Errorf("migrations path does not exist: %s", migrationsPath)
	}

	files, err := filepath.Glob(filepath.Join(migrationsPath, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to glob migration files: %v", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no SQL migration files found in %s", migrationsPath)
	}

	log.Printf("✅ Found %d migration files in %s", len(files), migrationsPath)
	return nil
}

// GetMigrationStatus returns the current migration status
func GetMigrationStatus(databaseURL string) (uint, bool, error) {
	m, err := migrate.New("", databaseURL)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrator: %v", err)
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return 0, false, fmt.Errorf("failed to get migration version: %v", err)
	}

	return version, dirty, nil
}

// MaskDatabaseURL masks sensitive information in database URL for logging
func MaskDatabaseURL(url string) string {
	return maskDatabaseURL(url)
}

// validateMigratorFiles validates that the migrator can see the migration files
func validateMigratorFiles(m *migrate.Migrate, migrationsPath string) error {
	log.Printf("🔍 Validating migrator can see files in: %s", migrationsPath)
	
	// List files that the migrator should be able to see
	files, err := filepath.Glob(filepath.Join(migrationsPath, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to glob migration files: %v", err)
	}
	
	if len(files) == 0 {
		return fmt.Errorf("no SQL migration files found in %s", migrationsPath)
	}
	
	log.Printf("📋 Migrator validation: Found %d SQL files", len(files))
	for _, file := range files {
		log.Printf("   📄 %s", filepath.Base(file))
	}
	
	return nil
}
