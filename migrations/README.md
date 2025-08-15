# Shared Migration Package

A comprehensive database migration package for the Jarakey platform, built on top of `golang-migrate/migrate/v4` with enhanced features and consistent behavior across all services.

## 🎯 Features

- ✅ **PostgreSQL Support**: Full PostgreSQL compatibility with proper SQL parsing
- ✅ **Dollar-Quote Handling**: Correctly handles PostgreSQL `$$` function definitions
- ✅ **Environment Variables**: Law.md compliant configuration via environment variables
- ✅ **Structured Logging**: Comprehensive logging with correlation IDs
- ✅ **Error Handling**: Graceful error handling and recovery
- ✅ **Timeout Support**: Configurable timeouts for long-running migrations
- ✅ **Validation**: Migration file validation and integrity checks
- ✅ **CLI Tool**: Simple command-line interface for all operations
- ✅ **Service Integration**: Easy integration with individual microservices

## 🚀 Quick Start

### 1. Install Dependencies

```bash
cd jarakey-shared-middleware
make install
```

### 2. Build Migration Tool

```bash
make build
```

### 3. Run Migrations

```bash
# Set database URL
export DATABASE_URL="postgres://user:pass@localhost:5432/db?sslmode=disable"

# Run migrations up
./bin/migrate-tool -command=up

# Check status
./bin/migrate-tool -command=status

# Validate files
./bin/migrate-tool -command=validate
```

## 📋 Available Commands

| Command | Description | Example |
|---------|-------------|---------|
| `up` | Run all pending migrations | `-command=up` |
| `down` | Rollback all migrations | `-command=down` |
| `force` | Force migration version | `-command=force -force-version=5` |
| `version` | Show current version | `-command=version` |
| `status` | Show migration status | `-command=status` |
| `validate` | Validate migration files | `-command=validate` |

## ⚙️ Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | Required |
| `MIGRATIONS_PATH` | Path to migration files | `infrastructure/scripts` |
| `MIGRATION_TIMEOUT` | Migration timeout | `30s` |
| `MIGRATION_LOG_LEVEL` | Log level | `info` |

### Command Line Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-database` | Database URL (overrides env) | `DATABASE_URL` |
| `-path` | Migrations path | `infrastructure/scripts` |
| `-timeout` | Migration timeout | `30s` |
| `-log-level` | Log level | `info` |
| `-force-version` | Force version (for force command) | `0` |

## 🔧 Service Integration

### 1. Add to Service's go.mod

```go
require (
    github.com/jarakey/jarakey-shared-middleware v0.1.0
)
```

### 2. Use in Service Code

```go
package main

import (
    "context"
    "log"
    
    "github.com/jarakey/jarakey-jarakey-shared-middleware/migrations"
)

func runMigrations() error {
    config := &migrations.Config{
        DatabaseURL: os.Getenv("DATABASE_URL"),
        MigrationsPath: "infrastructure/scripts",
        Timeout: 30 * time.Second,
        LogLevel: "info",
    }
    
    migrator, err := migrations.NewMigrator(config)
    if err != nil {
        return err
    }
    defer migrator.Close()
    
    ctx := context.Background()
    return migrator.Up(ctx)
}
```

### 3. Update Dockerfile

```dockerfile
# Copy shared migration tool
COPY --from=builder /go/bin/migrate-tool /bin/migrate-tool

# Use shared tool instead of custom implementation
CMD ["/bin/migrate-tool", "-command=up"]
```

## 📁 Migration File Structure

### Naming Convention

Migration files should follow the pattern: `{version}_{description}.sql`

```
infrastructure/scripts/
├── 001_init.sql
├── 002_create_users.sql
├── 003_add_indexes.sql
└── 004_functions.sql
```

### SQL Syntax Support

- ✅ **Standard SQL**: CREATE, ALTER, INSERT, UPDATE, DELETE
- ✅ **PostgreSQL Functions**: `$$` dollar-quoted strings
- ✅ **Complex Statements**: Multi-line statements with proper termination
- ✅ **Transactions**: Automatic transaction handling
- ✅ **Comments**: `--` and `/* */` comment styles

## 🧪 Testing

### Run Tests

```bash
make test
```

### Run Tests with Coverage

```bash
make test-coverage
```

### Test Individual Components

```bash
# Test migration package
go test -v ./migrations/

# Test CLI tool
go test -v ./migrations/cmd/
```

## 🚀 Deployment

### Build for Production

```bash
# Build for current platform
make build

# Build for multiple platforms
make build-all
```

### Install Locally

```bash
make install-local
```

### Docker Integration

```dockerfile
# Multi-stage build
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o migrate-tool migrations/cmd/main.go

FROM alpine:latest
COPY --from=builder /app/migrate-tool /bin/migrate-tool
CMD ["/bin/migrate-tool", "-command=up"]
```

## 🔍 Troubleshooting

### Common Issues

#### 1. "no such file or directory"
- Verify migration files exist in the specified path
- Check file permissions
- Use `-command=validate` to verify files

#### 2. "connection refused"
- Verify database is running
- Check DATABASE_URL format
- Ensure network connectivity

#### 3. "migration failed"
- Check database logs for detailed errors
- Verify SQL syntax is valid
- Use `-log-level=debug` for verbose output

### Debug Mode

```bash
# Enable debug logging
./bin/migrate-tool -command=up -log-level=debug

# Check migration status
./bin/migrate-tool -command=status

# Validate files
./bin/migrate-tool -command=validate
```

## 📚 Examples

### Basic Migration

```sql
-- 001_create_users.sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Function Definition

```sql
-- 002_user_functions.sql
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';
```

### Complex Migration

```sql
-- 003_user_roles.sql
BEGIN;

-- Create roles table
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL
);

-- Insert default roles
INSERT INTO roles (name) VALUES 
    ('user'),
    ('admin'),
    ('moderator'
);

-- Create user_roles junction table
CREATE TABLE user_roles (
    user_id INTEGER REFERENCES users(id),
    role_id INTEGER REFERENCES roles(id),
    PRIMARY KEY (user_id, role_id)
);

COMMIT;
```

## 🤝 Contributing

### Development Setup

```bash
make dev-setup
```

### Code Style

- Follow Go best practices
- Add tests for new functionality
- Update documentation
- Use conventional commits

### Testing

- Write unit tests for all functions
- Test with real PostgreSQL database
- Verify error handling scenarios
- Test timeout and cancellation

## 📄 License

This package is part of the Jarakey platform and follows the same licensing terms.

## 🔗 Related

- [golang-migrate/migrate](https://github.com/golang-migrate/migrate) - Base migration library
- [PostgreSQL Documentation](https://www.postgresql.org/docs/) - SQL syntax reference
