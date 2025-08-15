package migrations

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Config holds migration configuration
type Config struct {
	DatabaseURL    string
	MigrationsPath string
	Timeout        time.Duration
	LogLevel       string
}

// DefaultConfig returns default migration configuration
func DefaultConfig() *Config {
	return &Config{
		DatabaseURL:    os.Getenv("DATABASE_URL"),
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
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	// Create migrator instance
	m, err := migrate.New(
		fmt.Sprintf("file://%s", config.MigrationsPath),
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

	return &Migrator{
		config:  config,
		migrate: m,
	}, nil
}

// Up runs all pending migrations
func (m *Migrator) Up(ctx context.Context) error {
	log.Printf("🚀 Starting database migrations...")
	log.Printf("📁 Migrations path: %s", m.config.MigrationsPath)
	log.Printf("🔗 Database: %s", maskDatabaseURL(m.config.DatabaseURL))

	// Check if migrations path exists
	if _, err := os.Stat(m.config.MigrationsPath); os.IsNotExist(err) {
		return fmt.Errorf("migrations path does not exist: %s", m.config.MigrationsPath)
	}

	// List migration files
	files, err := m.listMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to list migration files: %v", err)
	}

	log.Printf("📋 Found %d migration files", len(files))
	for _, file := range files {
		log.Printf("   📄 %s", filepath.Base(file))
	}

	// Run migrations
	if err := m.migrate.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %v", err)
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
func (m *Migrator) listMigrationFiles() ([]string, error) {
	var files []string

	err := filepath.WalkDir(m.config.MigrationsPath, func(path string, info os.DirEntry, err error) error {
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
