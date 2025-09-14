# Makefile for pg-press migration tools
# Simplified version using only the Go migration script

.PHONY: help build clean migrate verify cleanup status full backup test deps install

# Default database path (can be overridden)
DB_PATH ?= ../data.db
GO_SCRIPT := migrate_mods.go
BINARY := migrate_mods
BACKUP_DIR := backups

# Default target
help: ## Show this help message
	@echo "PG-Press Migration Tools (Go-only)"
	@echo "=================================="
	@echo ""
	@echo "Usage: make [target] [DB_PATH=/path/to/db]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'
	@echo ""
	@echo "Examples:"
	@echo "  make full                    # Run complete migration with default database"
	@echo "  make migrate DB_PATH=/custom/path/data.db  # Migrate custom database"
	@echo "  make dry-run                 # See what would happen without changes"
	@echo "  make backup                  # Create database backup"

# Build the Go migration tool
build: ## Build the Go migration binary
	@echo "Building Go migration tool..."
	go build -o $(BINARY) $(GO_SCRIPT)
	@echo "✅ Built $(BINARY) binary"

# Install dependencies
deps: ## Install Go dependencies
	@echo "Installing Go dependencies..."
	go mod tidy
	@echo "✅ Dependencies installed"

# Clean build artifacts
clean: ## Remove built binaries and temporary files
	@echo "Cleaning up build artifacts..."
	rm -f $(BINARY)
	rm -f migration_stats_*.json
	@echo "✅ Cleanup complete"

# Create backup directory
$(BACKUP_DIR):
	mkdir -p $(BACKUP_DIR)

# Create database backup
backup: $(BACKUP_DIR) ## Create a timestamped backup of the database
	@if [ ! -f "$(DB_PATH)" ]; then \
		echo "❌ Database not found: $(DB_PATH)"; \
		exit 1; \
	fi
	@BACKUP_FILE="$(BACKUP_DIR)/data_backup_$$(date +%Y%m%d_%H%M%S).db"; \
	cp "$(DB_PATH)" "$$BACKUP_FILE"; \
	echo "✅ Database backed up to: $$BACKUP_FILE"

# Database status check
status: ## Show current migration status
	@echo "Checking migration status for: $(DB_PATH)"
	go run $(GO_SCRIPT) -db "$(DB_PATH)" -action status -v

# Test database connection
test-db: ## Test database connection and integrity
	@echo "Testing database connection: $(DB_PATH)"
	@if [ ! -f "$(DB_PATH)" ]; then \
		echo "❌ Database not found: $(DB_PATH)"; \
		exit 1; \
	fi
	@if sqlite3 "$(DB_PATH)" "SELECT 1;" >/dev/null 2>&1; then \
		echo "✅ Database connection successful"; \
	else \
		echo "❌ Database connection failed"; \
		exit 1; \
	fi

# Schema setup
setup: ## Set up database schema (add mods columns, create modifications table)
	@echo "Setting up database schema..."
	go run $(GO_SCRIPT) -db "$(DB_PATH)" -action setup -v

# Data migration
migrate: ## Migrate data from mods columns to modifications table
	@echo "Starting data migration..."
	go run $(GO_SCRIPT) -db "$(DB_PATH)" -action migrate -v

# Verify migration
verify: ## Verify migration integrity
	@echo "Verifying migration..."
	go run $(GO_SCRIPT) -db "$(DB_PATH)" -action verify -v

# Cleanup old columns (destructive)
cleanup: ## Remove old mods columns (DESTRUCTIVE - use after verification)
	@echo "⚠️  WARNING: This will permanently remove old mods columns!"
	@read -p "Are you sure? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		go run $(GO_SCRIPT) -db "$(DB_PATH)" -action cleanup -v; \
	else \
		echo "Cleanup cancelled"; \
	fi

# Force cleanup without confirmation
cleanup-force: ## Force cleanup without confirmation (DANGEROUS)
	@echo "🚨 FORCING CLEANUP WITHOUT CONFIRMATION!"
	go run $(GO_SCRIPT) -db "$(DB_PATH)" -action cleanup -force -v

# Complete migration workflow
full: ## Run complete migration process (setup + migrate + verify)
	@echo "Starting full migration process..."
	go run $(GO_SCRIPT) -db "$(DB_PATH)" -action full -v

# Dry run variants
dry-run: ## Show what full migration would do without making changes
	@echo "DRY RUN: Full migration process"
	go run $(GO_SCRIPT) -db "$(DB_PATH)" -action full -dry-run -v

dry-run-migrate: ## Show what migration would do without making changes
	@echo "DRY RUN: Data migration"
	go run $(GO_SCRIPT) -db "$(DB_PATH)" -action migrate -dry-run -v

dry-run-cleanup: ## Show what cleanup would do without making changes
	@echo "DRY RUN: Cleanup old columns"
	go run $(GO_SCRIPT) -db "$(DB_PATH)" -action cleanup -dry-run -v

# Using built binary instead of go run
bin-migrate: build ## Run migration using built binary
	@echo "Running migration with built binary..."
	./$(BINARY) -db "$(DB_PATH)" -action migrate -v

bin-full: build ## Run full migration using built binary
	@echo "Running full migration with built binary..."
	./$(BINARY) -db "$(DB_PATH)" -action full -v

bin-status: build ## Show status using built binary
	@echo "Running status check with built binary..."
	./$(BINARY) -db "$(DB_PATH)" -action status -v

# Development and testing targets
test: test-db ## Run basic tests and checks
	@echo "Running migration tests..."
	@echo "✅ Database connection test passed"
	go run $(GO_SCRIPT) -db "$(DB_PATH)" -action status

# Install the migration tool
install: build ## Install migration tool to system PATH
	@echo "Installing migration tool..."
	@if [ -w "/usr/local/bin" ]; then \
		cp $(BINARY) /usr/local/bin/pgpress-migrate; \
		echo "✅ Tool installed to /usr/local/bin/pgpress-migrate"; \
		echo "You can now run: pgpress-migrate -db /path/to/db -action full"; \
	else \
		echo "❌ Cannot write to /usr/local/bin. Try: sudo make install"; \
		exit 1; \
	fi

# Uninstall migration tool
uninstall: ## Remove migration tool from system PATH
	@echo "Uninstalling migration tool..."
	rm -f /usr/local/bin/pgpress-migrate
	@echo "✅ Tool uninstalled"

# Show database statistics
stats: ## Show database and migration statistics
	@echo "Database Statistics"
	@echo "=================="
	@echo "Database path: $(DB_PATH)"
	@if [ -f "$(DB_PATH)" ]; then \
		echo "Database size: $$(du -h "$(DB_PATH)" | cut -f1)"; \
		echo ""; \
		sqlite3 "$(DB_PATH)" "SELECT 'Total tables: ' || COUNT(*) FROM sqlite_master WHERE type='table';"; \
		if sqlite3 "$(DB_PATH)" "SELECT name FROM sqlite_master WHERE type='table' AND name='modifications';" | grep -q modifications; then \
			echo "Modifications table exists: Yes"; \
			sqlite3 "$(DB_PATH)" "SELECT 'Total modifications: ' || COUNT(*) FROM modifications;"; \
			sqlite3 "$(DB_PATH)" "SELECT 'Unique entities: ' || COUNT(DISTINCT entity_id || '_' || entity_type) FROM modifications;"; \
		else \
			echo "Modifications table exists: No"; \
		fi; \
	else \
		echo "❌ Database file not found"; \
	fi

# Check migration prerequisites
check: ## Check migration prerequisites
	@echo "Checking migration prerequisites..."
	@echo "================================="
	@echo -n "SQLite3: "
	@if command -v sqlite3 >/dev/null 2>&1; then \
		echo "✅ Found ($$(sqlite3 --version | cut -d' ' -f1))"; \
	else \
		echo "❌ Not found"; \
	fi
	@echo -n "Go: "
	@if command -v go >/dev/null 2>&1; then \
		echo "✅ Found ($$(go version | cut -d' ' -f3))"; \
	else \
		echo "❌ Not found"; \
	fi
	@echo -n "Database: "
	@if [ -f "$(DB_PATH)" ]; then \
		echo "✅ Found ($(DB_PATH))"; \
	else \
		echo "❌ Not found ($(DB_PATH))"; \
	fi

# Show recent migration activity
logs: ## Show recent migration activity and logs
	@echo "Recent Migration Activity"
	@echo "========================"
	@if [ -f "$(DB_PATH)" ] && sqlite3 "$(DB_PATH)" "SELECT name FROM sqlite_master WHERE type='table' AND name='modifications';" | grep -q modifications; then \
		sqlite3 "$(DB_PATH)" "SELECT datetime(created_at) as created, entity_type, entity_id FROM modifications ORDER BY created_at DESC LIMIT 10;"; \
	else \
		echo "No modifications table found"; \
	fi

# Package migration tool
package: build ## Create a distribution package of migration tool
	@echo "Creating migration package..."
	@PACKAGE_NAME="pgpress-migration-tool-$$(date +%Y%m%d)"; \
	mkdir -p "$$PACKAGE_NAME"; \
	cp $(BINARY) "$$PACKAGE_NAME/"; \
	cp $(GO_SCRIPT) "$$PACKAGE_NAME/"; \
	cp Makefile "$$PACKAGE_NAME/"; \
	cp README.md "$$PACKAGE_NAME/" 2>/dev/null || true; \
	tar -czf "$$PACKAGE_NAME.tar.gz" "$$PACKAGE_NAME"; \
	rm -rf "$$PACKAGE_NAME"; \
	echo "✅ Package created: $$PACKAGE_NAME.tar.gz"

# Show help for the Go tool
go-help: ## Show help for the Go migration tool
	@go run $(GO_SCRIPT) -help

# Quick commands for common operations
quick-migrate: ## Quick migration with minimal output
	@go run $(GO_SCRIPT) -db "$(DB_PATH)" -action full

quick-status: ## Quick status check
	@go run $(GO_SCRIPT) -db "$(DB_PATH)" -action status

quick-verify: ## Quick verification
	@go run $(GO_SCRIPT) -db "$(DB_PATH)" -action verify

# Advanced targets for development
dev-reset: ## Reset database to pre-migration state (DESTRUCTIVE)
	@echo "⚠️  WARNING: This will reset your database!"
	@read -p "Are you sure? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		sqlite3 "$(DB_PATH)" "DROP TABLE IF EXISTS modifications;"; \
		sqlite3 "$(DB_PATH)" "DROP VIEW IF EXISTS modification_stats;"; \
		echo "✅ Database reset to pre-migration state"; \
	else \
		echo "Reset cancelled"; \
	fi

# Create sample data for testing
dev-sample: ## Create sample data for testing migration
	@echo "Creating sample modification data..."
	@sqlite3 "$(DB_PATH)" "UPDATE trouble_reports SET mods = '[{\"action\":\"create\",\"timestamp\":\"2024-01-01T10:00:00Z\",\"user_id\":1,\"data\":{\"title\":\"Sample\"}}]' WHERE id = 1 AND (mods IS NULL OR mods = '[]');"
	@echo "✅ Sample data created"

# Performance test
perf-test: ## Run performance test on migration
	@echo "Running migration performance test..."
	@time go run $(GO_SCRIPT) -db "$(DB_PATH)" -action migrate -dry-run

# Show tool version/info
info: ## Show tool information
	@echo "PG-Press Migration Tool"
	@echo "======================"
	@echo "Go script: $(GO_SCRIPT)"
	@echo "Binary: $(BINARY)"
	@echo "Database: $(DB_PATH)"
	@echo "Go version: $$(go version)"
	@echo ""
	@echo "Available actions: full, setup, migrate, verify, cleanup, status"
