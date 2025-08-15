package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/jarakey/jarakey-shared-middleware/migrations"
)

func main() {
	// Parse command line flags
	var (
		command       = flag.String("command", "up", "Migration command: up, down, force, version, status")
		databaseURL   = flag.String("database", "", "Database URL (overrides DATABASE_URL env var)")
		migrationsPath = flag.String("path", "infrastructure/scripts", "Path to migration files")
		timeout       = flag.Duration("timeout", 30*time.Second, "Migration timeout")
		logLevel      = flag.String("log-level", "info", "Log level: debug, info, warn, error")
		forceVersion  = flag.Int("force-version", 0, "Force migration version (for force command)")
	)
	flag.Parse()

	// Set database URL from flag or environment variable
	dbURL := *databaseURL
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}

	if dbURL == "" {
		log.Fatal("❌ Database URL is required. Set DATABASE_URL environment variable or use -database flag")
	}

	// Create migration configuration
	config := &migrations.Config{
		DatabaseURL:    dbURL,
		MigrationsPath: *migrationsPath,
		Timeout:        *timeout,
		LogLevel:       *logLevel,
	}

	// Create migrator instance
	migrator, err := migrations.NewMigrator(config)
	if err != nil {
		log.Fatalf("❌ Failed to create migrator: %v", err)
	}
	defer migrator.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// Execute command
	switch *command {
	case "up":
		log.Printf("🚀 Running migrations UP...")
		if err := migrator.Up(ctx); err != nil {
			log.Fatalf("❌ Migration failed: %v", err)
		}
		log.Printf("✅ Migrations completed successfully")

	case "down":
		log.Printf("🔄 Rolling back all migrations...")
		if err := migrator.Down(ctx); err != nil {
			log.Fatalf("❌ Rollback failed: %v", err)
		}
		log.Printf("✅ All migrations rolled back")

	case "force":
		if *forceVersion == 0 {
			log.Fatal("❌ Force version is required. Use -force-version flag")
		}
		log.Printf("🔧 Force setting migration version to %d", *forceVersion)
		if err := migrator.Force(*forceVersion); err != nil {
			log.Fatalf("❌ Force version failed: %v", err)
		}
		log.Printf("✅ Migration version set to %d", *forceVersion)

	case "version":
		version, dirty, err := migrator.Version()
		if err != nil {
			log.Fatalf("❌ Failed to get version: %v", err)
		}
		log.Printf("📊 Current migration version: %d", version)
		if dirty {
			log.Printf("⚠️  Database is in dirty state")
		} else {
			log.Printf("✅ Database is clean")
		}

	case "status":
		version, dirty, err := migrator.Version()
		if err != nil {
			log.Printf("📊 Migration status: No migrations applied")
			return
		}
		log.Printf("📊 Migration status:")
		log.Printf("   Version: %d", version)
		log.Printf("   Dirty: %t", dirty)
		log.Printf("   Database: %s", migrations.MaskDatabaseURL(dbURL))

	case "validate":
		log.Printf("🔍 Validating migration files...")
		if err := migrations.ValidateMigrationFiles(*migrationsPath); err != nil {
			log.Fatalf("❌ Validation failed: %v", err)
		}
		log.Printf("✅ Migration files are valid")

	default:
		log.Fatalf("❌ Unknown command: %s. Use: up, down, force, version, status, validate", *command)
	}
}

// MaskDatabaseURL masks sensitive information in database URL for logging
func MaskDatabaseURL(url string) string {
	if len(url) < 10 {
		return "***"
	}
	
	// Show only host and database name, mask credentials
	if len(url) > 20 {
		return url[:20] + "***"
	}
	
	return "***"
}
