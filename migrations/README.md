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
- ✅ **Explicit Path Handling**: No more guessing - specify exact migration paths

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

# Run migrations up (explicit path)
./bin/migrate-tool -command=up -path=infrastructure/scripts

# Check status
./bin/migrate-tool -command=status -path=infrastructure/scripts

# Validate files
./bin/migrate-tool -command=validate -path=infrastructure/scripts
```

## 📋 Available Commands

| Command | Description | Example |
|---------|-------------|---------|
| `up` | Run all pending migrations | `-command=up -path=infrastructure/scripts` |
| `down` | Rollback all migrations | `-command=down -path=infrastructure/scripts` |
| `force` | Force migration version | `-command=force -force-version=5 -path=infrastructure/scripts` |
| `version` | Show current version | `-command=version -path=infrastructure/scripts` |
| `status` | Show migration status | `-command=status -path=infrastructure/scripts` |
| `validate` | Validate migration files | `-command=validate -path=infrastructure/scripts` |

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
| `-path` | **Migrations path (REQUIRED)** | `infrastructure/scripts` |
| `-timeout` | Migration timeout | `30s` |
| `-log-level` | Log level | `info` |
| `-force-version` | Force version (for force command) | `0` |

## 🔧 Service Integration

### 1. Add to Service's go.mod

```go
require (
    github.com/jarakey/jarakey-shared-middleware v1.2.0
)
```

### 2. Use in Service Code

```go
package main

import (
    "context"
    "log"
    
    "github.com/jarakey/jarakey-shared-middleware/migrations"
)

func main() {
    config := migrations.DefaultConfig()
    config.MigrationsPath = "/app/infrastructure/scripts"  // Explicit path
    
    migrator, err := migrations.NewMigrator(config)
    if err != nil {
        log.Fatal(err)
    }
    defer migrator.Close()
    
    if err := migrator.Up(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

## 🐳 Docker Usage

### Explicit Path in Docker

```dockerfile
# Copy migrations to a known location
COPY --from=builder /src/infrastructure/scripts /app/migrations

# Run migrations with explicit path
CMD ["./bin/migrate", "-path", "/app/migrations"]
```

### Kubernetes Job Example

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: user-service-migrations
spec:
  template:
    spec:
      containers:
      - name: migrations
        image: user-service:latest
        command: ["./bin/migrate"]
        args: ["-path", "/app/infrastructure/scripts"]
        volumeMounts:
        - name: migrations
          mountPath: /app/infrastructure/scripts
      volumes:
      - name: migrations
        configMap:
          name: user-service-migrations
```

## 🚨 Important Changes in v1.2.0

- **Explicit Path Required**: Always specify the `-path` flag or set `MigrationsPath` in config
- **No More Auto-Detection**: Removed complex path guessing logic
- **Docker-Friendly**: Works consistently across all container environments
- **Simplified Integration**: Clear and predictable behavior
