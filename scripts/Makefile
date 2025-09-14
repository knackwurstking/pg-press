# Makefile for pg-press migration tools
# Provides convenient targets for building and running migration scripts

.PHONY: help build clean migrate verify cleanup status full backup test install deps

# Default database path (can be overridden)
DB_PATH ?= ../data.db
GO_SCRIPT := migrate_mods.go
SHELL_SCRIPT := ./run_migration.sh
BACKUP_DIR := backups

# Default target
help: ## Show this help message
	@echo "PG-Press Migration Tools"
	@echo "======================="
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
	go build -o migrate_mods $(GO_SCRIPT)
	@echo "✅ Built migrate_mods binary"

# Install dependencies
deps: ## Install Go dependencies
	@echo "Installing Go dependencies..."
	go mod tidy
	@echo "✅ Dependencies installed"

# Clean build artifacts
clean: ## Remove built binaries and temporary files
	@echo "Cleaning up build artifacts..."
	rm -f migrate_mods
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
	$(SHELL_SCRIPT) --db "$(DB_PATH)" status

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
setup: backup ## Set up database schema (add mods columns, create modifications table)
	@echo "Setting up database schema..."
	$(SHELL_SCRIPT) --db "$(DB_PATH)" setup

# Data migration
migrate: backup ## Migrate data from mods columns to modifications table
	@echo "Starting data migration..."
	$(SHELL_SCRIPT) --db "$(DB_PATH)" migrate

# Verify migration
verify: ## Verify migration integrity
	@echo "Verifying migration..."
	$(SHELL_SCRIPT) --db "$(DB_PATH)" verify

# Cleanup old columns (destructive)
cleanup: ## Remove old mods columns (DESTRUCTIVE - use after verification)
	@echo "⚠️  WARNING: This will permanently remove old mods columns!"
	@read -p "Are you sure? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		$(SHELL_SCRIPT) --db "$(DB_PATH)" cleanup; \
	else \
		echo "Cleanup cancelled"; \
	fi

# Force cleanup without confirmation
cleanup-force: ## Force cleanup without confirmation (DANGEROUS)
	@echo "🚨 FORCING CLEANUP WITHOUT CONFIRMATION!"
	$(SHELL_SCRIPT) --db "$(DB_PATH)" --force cleanup

# Complete migration workflow
full: backup ## Run complete migration process (setup + migrate + verify)
	@echo "Starting full migration process..."
	$(SHELL_SCRIPT) --db "$(DB_PATH)" full

# Dry run variants
dry-run: ## Show what full migration would do without making changes
	@echo "DRY RUN: Full migration process"
	$(SHELL_SCRIPT) --db "$(DB_PATH)" --dry-run full

dry-run-migrate: ## Show what migration would do without making changes
	@echo "DRY RUN: Data migration"
	$(SHELL_SCRIPT) --db "$(DB_PATH)" --dry-run migrate

dry-run-cleanup: ## Show what cleanup would do without making changes
	@echo "DRY RUN: Cleanup old columns"
	$(SHELL_SCRIPT) --db "$(DB_PATH)" --dry-run cleanup

# Using Go script directly
go-migrate: build ## Run migration using standalone Go script
	@echo "Running Go migration tool..."
	./migrate_mods --db "$(DB_PATH)" --action migrate

go-verify: build ## Verify migration using standalone Go script
	@echo "Running Go verification tool..."
	./migrate_mods --db "$(DB_PATH)" --action verify

go-cleanup: build ## Cleanup using standalone Go script
	@echo "Running Go cleanup tool..."
	./migrate_mods --db "$(DB_PATH)" --action cleanup

go-status: build ## Show status using standalone Go script
	@echo "Running Go status tool..."
	./migrate_mods --db "$(DB_PATH)" --action status

# Export migration data
export: ## Export migration data to JSON file
	@echo "Exporting migration data..."
	@if command -v pgpress >/dev/null 2>&1; then \
		pgpress migration export --db "$(DB_PATH)" --output migration_export_$$(date +%Y%m%d_%H%M%S).json; \
	else \
		echo "❌ pgpress command not found. Please install pgpress or use manual export."; \
		exit 1; \
	fi

# Development and testing targets
test: test-db ## Run basic tests and checks
	@echo "Running migration tests..."
	@echo "✅ Database connection test passed"
	$(SHELL_SCRIPT) --db "$(DB_PATH)" --dry-run status

# Install the migration tools
install: build ## Install migration tools to system PATH
	@echo "Installing migration tools..."
	@if [ -w "/usr/local/bin" ]; then \
		cp migrate_mods /usr/local/bin/; \
		cp $(SHELL_SCRIPT) /usr/local/bin/pgpress-migrate; \
		chmod +x /usr/local/bin/pgpress-migrate; \
		echo "✅ Tools installed to /usr/local/bin/"; \
	else \
		echo "❌ Cannot write to /usr/local/bin. Try: sudo make install"; \
		exit 1; \
	fi

# Uninstall migration tools
uninstall: ## Remove migration tools from system PATH
	@echo "Uninstalling migration tools..."
	rm -f /usr/local/bin/migrate_mods
	rm -f /usr/local/bin/pgpress-migrate
	@echo "✅ Tools uninstalled"

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
	@echo -n "pgpress: "
	@if command -v pgpress >/dev/null 2>&1; then \
		echo "✅ Found"; \
	else \
		echo "⚠️  Not found (optional)"; \
	fi
	@echo -n "Database: "
	@if [ -f "$(DB_PATH)" ]; then \
		echo "✅ Found ($(DB_PATH))"; \
	else \
		echo "❌ Not found ($(DB_PATH))"; \
	fi

# Advanced targets for development
dev-reset: ## Reset development database to pre-migration state (DESTRUCTIVE)
	@echo "⚠️  WARNING: This will reset your development database!"
	@read -p "Are you sure? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		sqlite3 "$(DB_PATH)" "DROP TABLE IF EXISTS modifications;"; \
		sqlite3 "$(DB_PATH)" "DROP VIEW IF EXISTS migration_status;"; \
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

# Show recent migration activity
logs: ## Show recent migration activity and logs
	@echo "Recent Migration Activity"
	@echo "========================"
	@if [ -f "$(DB_PATH)" ] && sqlite3 "$(DB_PATH)" "SELECT name FROM sqlite_master WHERE type='table' AND name='modifications';" | grep -q modifications; then \
		sqlite3 "$(DB_PATH)" "SELECT datetime(created_at) as created, entity_type, entity_id FROM modifications ORDER BY created_at DESC LIMIT 10;"; \
	else \
		echo "No modifications table found"; \
	fi

# Package migration tools
package: build ## Create a distribution package of migration tools
	@echo "Creating migration package..."
	@PACKAGE_NAME="pgpress-migration-tools-$$(date +%Y%m%d)"; \
	mkdir -p "$$PACKAGE_NAME"; \
	cp migrate_mods "$$PACKAGE_NAME/"; \
	cp $(SHELL_SCRIPT) "$$PACKAGE_NAME/"; \
	cp migrate_mods_to_modifications.sql "$$PACKAGE_NAME/"; \
	cp README.md "$$PACKAGE_NAME/"; \
	cp Makefile "$$PACKAGE_NAME/"; \
	tar -czf "$$PACKAGE_NAME.tar.gz" "$$PACKAGE_NAME"; \
	rm -rf "$$PACKAGE_NAME"; \
	echo "✅ Package created: $$PACKAGE_NAME.tar.gz"

# Default variables that can be overridden
VERBOSE ?= false
FORCE ?= false
DRY_RUN ?= false

# Pass variables to shell script
SCRIPT_FLAGS = --db "$(DB_PATH)"
ifeq ($(VERBOSE),true)
	SCRIPT_FLAGS += --verbose
endif
ifeq ($(FORCE),true)
	SCRIPT_FLAGS += --force
endif
ifeq ($(DRY_RUN),true)
	SCRIPT_FLAGS += --dry-run
endif
