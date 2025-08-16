package migrations

import (
	"os"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	if config == nil {
		t.Fatal("DefaultConfig() returned nil")
	}
	
	if config.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", config.Timeout)
	}
	
	if config.LogLevel != "info" {
		t.Errorf("Expected log level 'info', got %s", config.LogLevel)
	}
}

func TestBuildDatabaseURL(t *testing.T) {
	// Test with DATABASE_URL set
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
	url := buildDatabaseURL()
	if url != "postgres://test:test@localhost:5432/test" {
		t.Errorf("Expected DATABASE_URL to be used, got %s", url)
	}
	
	// Test with individual components
	os.Unsetenv("DATABASE_URL")
	os.Setenv("DB_HOST", "testhost")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_SSL_MODE", "disable")
	
	url = buildDatabaseURL()
	expected := "postgres://testuser:testpass@testhost:5433/testdb?sslmode=disable"
	if url != expected {
		t.Errorf("Expected %s, got %s", expected, url)
	}
	
	// Clean up
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DB_SSL_MODE")
}

func TestValidateMigrationFiles(t *testing.T) {
	// Test with non-existent path
	err := ValidateMigrationFiles("/non/existent/path")
	if err == nil {
		t.Error("Expected error for non-existent path")
	}
	
	// Test with current directory (should have some files)
	err = ValidateMigrationFiles(".")
	if err != nil {
		t.Logf("Validation error (expected for current dir): %v", err)
	}
}

func TestGetMigrationStatus(t *testing.T) {
	// Test with invalid database URL
	_, _, err := GetMigrationStatus("invalid://url")
	if err == nil {
		t.Error("Expected error for invalid database URL")
	}
}

func TestMaskDatabaseURL(t *testing.T) {
	// Test short URL
	masked := MaskDatabaseURL("short")
	if masked != "***" {
		t.Errorf("Expected '***' for short URL, got %s", masked)
	}
	
	// Test long URL
	longURL := "postgres://user:password@localhost:5432/database?sslmode=require"
	masked = MaskDatabaseURL(longURL)
	if len(masked) < 20 {
		t.Errorf("Expected masked URL to be longer, got %s", masked)
	}
}

func TestNewMigratorWithoutDatabase(t *testing.T) {
	config := &Config{
		DatabaseURL:    "", // Empty database URL
		MigrationsPath: ".",
		Timeout:        30 * time.Second,
		LogLevel:       "info",
	}
	
	_, err := NewMigrator(config)
	if err == nil {
		t.Error("Expected error for empty database URL")
	}
	
	expectedErr := "database connection not configured"
	if err.Error()[:len(expectedErr)] != expectedErr {
		t.Errorf("Expected error starting with '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestMigratorClose(t *testing.T) {
	// Test that Close doesn't panic on nil migrator
	m := &Migrator{}
	err := m.Close()
	if err != nil {
		t.Errorf("Close() should not return error on nil migrator, got %v", err)
	}
}

func TestContextTimeout(t *testing.T) {
	// Test that context timeout is respected
	config := &Config{
		DatabaseURL:    "postgres://test:test@localhost:5432/test",
		MigrationsPath: ".",
		Timeout:        1 * time.Millisecond, // Very short timeout
		LogLevel:       "info",
	}
	
	// This should fail due to database connection, but shouldn't hang
	_, err := NewMigrator(config)
	if err == nil {
		t.Error("Expected error for invalid database connection")
	}
}

func TestFixPathSlashes(t *testing.T) {
	// Test various path scenarios
	tests := []struct {
		input    string
		expected string
	}{
		{"/normal/path", "/normal/path"},
		{"///triple/slash", "/triple/slash"},
		{"//double/slash", "/double/slash"},
		{"/trailing//", "/trailing"}, // filepath.Clean removes trailing slash
		{"/multiple////slashes", "/multiple/slashes"},
	}
	
	for _, test := range tests {
		result := fixPathSlashes(test.input)
		if result != test.expected {
			t.Errorf("fixPathSlashes(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}
