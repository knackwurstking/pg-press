package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	// ANSI color codes for output
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

type MigrationScript struct {
	db     *sql.DB
	dbPath string
}

type MigrationStats struct {
	TroubleReportsProcessed int           `json:"trouble_reports_processed"`
	MetalSheetsProcessed    int           `json:"metal_sheets_processed"`
	ToolsProcessed          int           `json:"tools_processed"`
	TotalModsMigrated       int           `json:"total_mods_migrated"`
	Errors                  int           `json:"errors"`
	StartTime               time.Time     `json:"start_time"`
	EndTime                 time.Time     `json:"end_time"`
	Duration                time.Duration `json:"duration"`
}

type User struct {
	TelegramID int64  `json:"telegram_id"`
	UserName   string `json:"user_name"`
	APIKey     string `json:"api_key"`
	LastFeed   int    `json:"last_feed"`
}

type ModificationData struct {
	Action string                 `json:"action"`
	Time   int64                  `json:"time"`
	User   User                   `json:"user"`
	Data   map[string]interface{} `json:"data"`
}

func main() {
	var (
		dbPath  = flag.String("db", "./data.db", "Path to SQLite database file")
		action  = flag.String("action", "full", "Action to perform: full, setup, migrate, verify, cleanup, status")
		force   = flag.Bool("force", false, "Force operation without confirmation")
		dryRun  = flag.Bool("dry-run", false, "Show what would be done without making changes")
		verbose = flag.Bool("v", false, "Verbose output")
		backup  = flag.Bool("backup", true, "Create backup before migration (disable with --backup=false)")
		help    = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	printBanner()

	script := &MigrationScript{dbPath: *dbPath}

	// Validate database path
	if err := script.validateDatabase(); err != nil {
		printError("Database validation failed: %v", err)
		os.Exit(1)
	}

	if err := script.connect(); err != nil {
		printError("Failed to connect to database: %v", err)
		os.Exit(1)
	}
	defer script.db.Close()

	// Create backup unless disabled or in dry-run mode
	if *backup && !*dryRun && (*action == "full" || *action == "setup" || *action == "migrate" || *action == "cleanup") {
		if err := script.createBackup(); err != nil {
			printError("Failed to create backup: %v", err)
			os.Exit(1)
		}
	}

	switch *action {
	case "full":
		if err := script.runFullMigration(*dryRun, *verbose, *force); err != nil {
			printError("Full migration failed: %v", err)
			os.Exit(1)
		}
	case "setup":
		if err := script.setup(*dryRun, *verbose); err != nil {
			printError("Setup failed: %v", err)
			os.Exit(1)
		}
	case "migrate":
		if err := script.migrate(*dryRun, *verbose); err != nil {
			printError("Migration failed: %v", err)
			os.Exit(1)
		}
	case "verify":
		if err := script.verify(*verbose); err != nil {
			printError("Verification failed: %v", err)
			os.Exit(1)
		}
	case "cleanup":
		if err := script.cleanup(*force, *dryRun, *verbose); err != nil {
			printError("Cleanup failed: %v", err)
			os.Exit(1)
		}
	case "status":
		if err := script.status(*verbose); err != nil {
			printError("Status check failed: %v", err)
			os.Exit(1)
		}
	default:
		printError("Unknown action: %s", *action)
		showHelp()
		os.Exit(1)
	}

	printSuccess("Migration script completed successfully!")
}

func showHelp() {
	fmt.Println(colorBold + "PG-Press Migration Tool" + colorReset)
	fmt.Println("=======================")
	fmt.Println()
	fmt.Println("A standalone tool for migrating from old mods columns to the new modifications table.")
	fmt.Println()
	fmt.Println(colorBold + "USAGE:" + colorReset)
	fmt.Println("  go run migrate_mods.go [OPTIONS]")
	fmt.Println()
	fmt.Println(colorBold + "ACTIONS:" + colorReset)
	fmt.Println("  full     Complete migration (setup + migrate + verify) [DEFAULT]")
	fmt.Println("  setup    Set up database schema (add mods columns, create modifications table)")
	fmt.Println("  migrate  Migrate data from mods columns to modifications table")
	fmt.Println("  verify   Verify migration integrity")
	fmt.Println("  cleanup  Remove old mods columns (DESTRUCTIVE!)")
	fmt.Println("  status   Show current migration status")
	fmt.Println()
	fmt.Println(colorBold + "OPTIONS:" + colorReset)
	fmt.Println("  -db string          Database path (default \"./data.db\")")
	fmt.Println("  -action string      Action to perform (default \"full\")")
	fmt.Println("  -force              Force operation without confirmation")
	fmt.Println("  -dry-run            Show what would be done without making changes")
	fmt.Println("  -v                  Verbose output")
	fmt.Println("  -backup             Create backup before migration (default true)")
	fmt.Println("  -help               Show this help message")
	fmt.Println()
	fmt.Println(colorBold + "EXAMPLES:" + colorReset)
	fmt.Println("  go run migrate_mods.go                          # Run complete migration")
	fmt.Println("  go run migrate_mods.go -dry-run                 # See what would happen")
	fmt.Println("  go run migrate_mods.go -action=status           # Check migration status")
	fmt.Println("  go run migrate_mods.go -db=/path/to/db migrate   # Migrate custom database")
	fmt.Println("  go run migrate_mods.go -force -action=cleanup   # Force cleanup")
	fmt.Println()
	fmt.Println(colorBold + "MIGRATION WORKFLOW:" + colorReset)
	fmt.Println("  1. Backup your database first (done automatically)")
	fmt.Println("  2. Run: go run migrate_mods.go -action=full")
	fmt.Println("  3. Test your application thoroughly")
	fmt.Println("  4. Optionally run cleanup: go run migrate_mods.go -action=cleanup")
}

func printBanner() {
	fmt.Println(colorCyan + colorBold + "╔══════════════════════════════════════════════════════╗")
	fmt.Println("║            PG-Press Migration Tool v2.0             ║")
	fmt.Println("║     Standalone Migration from Mods to Modifications ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝" + colorReset)
	fmt.Println()
}

func printSuccess(format string, args ...interface{}) {
	fmt.Printf(colorGreen+"✅ "+format+colorReset+"\n", args...)
}

func printError(format string, args ...interface{}) {
	fmt.Printf(colorRed+"❌ "+format+colorReset+"\n", args...)
}

func printWarning(format string, args ...interface{}) {
	fmt.Printf(colorYellow+"⚠️  "+format+colorReset+"\n", args...)
}

func printInfo(format string, args ...interface{}) {
	fmt.Printf(colorBlue+"ℹ️  "+format+colorReset+"\n", args...)
}

func (m *MigrationScript) validateDatabase() error {
	if m.dbPath == "" {
		return fmt.Errorf("database path cannot be empty")
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(m.dbPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}
	m.dbPath = absPath

	if _, err := os.Stat(m.dbPath); os.IsNotExist(err) {
		return fmt.Errorf("database file does not exist: %s", m.dbPath)
	}

	printInfo("Database path: %s", m.dbPath)
	return nil
}

func (m *MigrationScript) connect() error {
	db, err := sql.Open("sqlite3", m.dbPath+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	m.db = db
	printSuccess("Database connection established")
	return nil
}

func (m *MigrationScript) createBackup() error {
	backupDir := filepath.Dir(m.dbPath) + "/backups"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %v", err)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("data_backup_%s.db", timestamp))

	printInfo("Creating database backup...")

	// Read source file
	sourceData, err := os.ReadFile(m.dbPath)
	if err != nil {
		return fmt.Errorf("failed to read source database: %v", err)
	}

	// Write backup file
	if err := os.WriteFile(backupPath, sourceData, 0644); err != nil {
		return fmt.Errorf("failed to write backup: %v", err)
	}

	printSuccess("Database backed up to: %s", backupPath)
	return nil
}

func (m *MigrationScript) runFullMigration(dryRun, verbose, force bool) error {
	printInfo("Starting full migration process...")

	if dryRun {
		printWarning("DRY RUN MODE - No changes will be made")
	}

	// Step 1: Setup schema
	printInfo("Step 1/3: Setting up database schema...")
	if err := m.setup(dryRun, verbose); err != nil {
		return fmt.Errorf("setup failed: %v", err)
	}

	// Step 2: Migrate data
	printInfo("Step 2/3: Migrating data...")
	if err := m.migrate(dryRun, verbose); err != nil {
		return fmt.Errorf("migration failed: %v", err)
	}

	// Step 3: Verify migration
	printInfo("Step 3/3: Verifying migration...")
	if err := m.verify(verbose); err != nil {
		return fmt.Errorf("verification failed: %v", err)
	}

	if dryRun {
		printWarning("DRY RUN completed - No actual changes were made")
		printInfo("Run without --dry-run to perform the actual migration")
	} else {
		printSuccess("Full migration completed successfully!")
		printWarning("IMPORTANT: Test your application thoroughly before running cleanup!")
		printInfo("To remove old mods columns, run: go run migrate_mods.go -action=cleanup")
	}

	return nil
}

func (m *MigrationScript) setup(dryRun, verbose bool) error {
	printInfo("Setting up database schema...")

	queries := []string{
		// Add mods columns if they don't exist
		`ALTER TABLE trouble_reports ADD COLUMN mods TEXT DEFAULT '[]'`,
		`ALTER TABLE metal_sheets ADD COLUMN mods TEXT DEFAULT '[]'`,
		`ALTER TABLE tools ADD COLUMN mods TEXT DEFAULT '[]'`,

		// Create modifications table
		`CREATE TABLE IF NOT EXISTS modifications (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			entity_type TEXT NOT NULL,
			entity_id INTEGER NOT NULL,
			data BLOB NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(user_id) REFERENCES users(telegram_id) ON DELETE CASCADE
		)`,

		// Create indexes
		`CREATE INDEX IF NOT EXISTS idx_modifications_entity ON modifications(entity_type, entity_id)`,
		`CREATE INDEX IF NOT EXISTS idx_modifications_user ON modifications(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_modifications_created_at ON modifications(created_at)`,

		// Create helpful views
		`CREATE VIEW IF NOT EXISTS migration_status AS
		SELECT
			'trouble_reports' as table_name,
			COUNT(*) as total_records,
			SUM(CASE WHEN mods IS NOT NULL AND json_array_length(mods) > 0 THEN 1 ELSE 0 END) as records_with_mods
		FROM trouble_reports
		UNION ALL
		SELECT
			'metal_sheets' as table_name,
			COUNT(*) as total_records,
			SUM(CASE WHEN mods IS NOT NULL AND json_array_length(mods) > 0 THEN 1 ELSE 0 END) as records_with_mods
		FROM metal_sheets
		UNION ALL
		SELECT
			'tools' as table_name,
			COUNT(*) as total_records,
			SUM(CASE WHEN mods IS NOT NULL AND json_array_length(mods) > 0 THEN 1 ELSE 0 END) as records_with_mods
		FROM tools`,

		`CREATE VIEW IF NOT EXISTS modification_stats AS
		SELECT
			entity_type,
			COUNT(*) as total_modifications,
			COUNT(DISTINCT entity_id) as unique_entities,
			MIN(created_at) as earliest_modification,
			MAX(created_at) as latest_modification
		FROM modifications
		GROUP BY entity_type`,
	}

	for i, query := range queries {
		if verbose || dryRun {
			fmt.Printf("Query %d: %s\n", i+1, strings.TrimSpace(query))
		}

		if dryRun {
			continue
		}

		if _, err := m.db.Exec(query); err != nil {
			// Ignore column already exists errors
			if !isColumnExistsError(err) {
				return fmt.Errorf("failed to execute query %d: %v", i+1, err)
			}
			if verbose {
				printWarning("Column already exists (this is normal): %v", err)
			}
		}
	}

	if dryRun {
		printInfo("Schema setup would execute %d queries", len(queries))
	} else {
		printSuccess("Database schema setup completed")
	}

	return nil
}

func (m *MigrationScript) migrate(dryRun, verbose bool) error {
	printInfo("Starting data migration...")

	if dryRun {
		printInfo("DRY RUN: Migration preview")
	}

	stats := &MigrationStats{StartTime: time.Now()}

	// Migrate each table
	if err := m.migrateTroubleReports(stats, dryRun, verbose); err != nil {
		return fmt.Errorf("failed to migrate trouble reports: %v", err)
	}

	if err := m.migrateMetalSheets(stats, dryRun, verbose); err != nil {
		return fmt.Errorf("failed to migrate metal sheets: %v", err)
	}

	if err := m.migrateTools(stats, dryRun, verbose); err != nil {
		return fmt.Errorf("failed to migrate tools: %v", err)
	}

	stats.EndTime = time.Now()
	stats.Duration = stats.EndTime.Sub(stats.StartTime)
	stats.TotalModsMigrated = stats.TroubleReportsProcessed + stats.MetalSheetsProcessed + stats.ToolsProcessed

	// Print summary
	fmt.Println()
	printInfo("Migration Summary:")
	fmt.Printf("  Trouble Reports processed: %d\n", stats.TroubleReportsProcessed)
	fmt.Printf("  Metal Sheets processed: %d\n", stats.MetalSheetsProcessed)
	fmt.Printf("  Tools processed: %d\n", stats.ToolsProcessed)
	fmt.Printf("  Total entities migrated: %d\n", stats.TotalModsMigrated)
	fmt.Printf("  Errors: %d\n", stats.Errors)
	fmt.Printf("  Duration: %v\n", stats.Duration)

	if stats.Errors > 0 {
		printWarning("Migration completed with %d errors", stats.Errors)
	} else if dryRun {
		printInfo("DRY RUN completed - No changes were made")
	} else {
		printSuccess("Data migration completed successfully!")
	}

	return nil
}

func (m *MigrationScript) migrateTroubleReports(stats *MigrationStats, dryRun, verbose bool) error {
	if verbose {
		printInfo("Migrating trouble report modifications...")
	}

	query := `SELECT id, mods FROM trouble_reports WHERE mods IS NOT NULL AND json_array_length(mods) > 0`

	rows, err := m.db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query trouble reports: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var modsJSON string

		if err := rows.Scan(&id, &modsJSON); err != nil {
			if verbose {
				printError("Error scanning trouble report %d: %v", id, err)
			}
			stats.Errors++
			continue
		}

		if err := m.processMods(id, "trouble_reports", modsJSON, dryRun, verbose); err != nil {
			if verbose {
				printError("Error processing mods for trouble report %d: %v", id, err)
			}
			stats.Errors++
			continue
		}

		stats.TroubleReportsProcessed++
		if verbose {
			printInfo("Processed trouble report %d", id)
		}
	}

	return rows.Err()
}

func (m *MigrationScript) migrateMetalSheets(stats *MigrationStats, dryRun, verbose bool) error {
	if verbose {
		printInfo("Migrating metal sheet modifications...")
	}

	query := `SELECT id, mods FROM metal_sheets WHERE mods IS NOT NULL AND json_array_length(mods) > 0`

	rows, err := m.db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query metal sheets: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var modsJSON string

		if err := rows.Scan(&id, &modsJSON); err != nil {
			if verbose {
				printError("Error scanning metal sheet %d: %v", id, err)
			}
			stats.Errors++
			continue
		}

		if err := m.processMods(id, "metal_sheets", modsJSON, dryRun, verbose); err != nil {
			if verbose {
				printError("Error processing mods for metal sheet %d: %v", id, err)
			}
			stats.Errors++
			continue
		}

		stats.MetalSheetsProcessed++
		if verbose {
			printInfo("Processed metal sheet %d", id)
		}
	}

	return rows.Err()
}

func (m *MigrationScript) migrateTools(stats *MigrationStats, dryRun, verbose bool) error {
	if verbose {
		printInfo("Migrating tool modifications...")
	}

	query := `SELECT id, mods FROM tools WHERE mods IS NOT NULL AND json_array_length(mods) > 0`

	rows, err := m.db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query tools: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var modsJSON string

		if err := rows.Scan(&id, &modsJSON); err != nil {
			if verbose {
				printError("Error scanning tool %d: %v", id, err)
			}
			stats.Errors++
			continue
		}

		if err := m.processMods(id, "tools", modsJSON, dryRun, verbose); err != nil {
			if verbose {
				printError("Error processing mods for tool %d: %v", id, err)
			}
			stats.Errors++
			continue
		}

		stats.ToolsProcessed++
		if verbose {
			printInfo("Processed tool %d", id)
		}
	}

	return rows.Err()
}

func (m *MigrationScript) processMods(entityID int64, entityType, modsJSON string, dryRun, verbose bool) error {
	var mods []ModificationData
	if err := json.Unmarshal([]byte(modsJSON), &mods); err != nil {
		return fmt.Errorf("failed to unmarshal mods JSON: %v", err)
	}

	for i, mod := range mods {
		if verbose && len(mods) > 1 {
			fmt.Printf("    Processing mod %d/%d for %s %d: action=%s, user=%d\n",
				i+1, len(mods), entityType, entityID, mod.Action, mod.User.TelegramID)
		}

		if dryRun {
			if verbose {
				fmt.Printf("    WOULD INSERT: entity_type=%s, entity_id=%d, user_id=%d\n",
					entityType, entityID, mod.User.TelegramID)
			}
			continue
		}

		// Convert modification data to JSON
		modDataJSON, err := json.Marshal(mod.Data)
		if err != nil {
			return fmt.Errorf("failed to marshal mod data: %v", err)
		}

		// Convert Unix timestamp (milliseconds) to time.Time
		timestamp := time.Unix(0, mod.Time*int64(time.Millisecond))

		// Insert into modifications table
		query := `
			INSERT INTO modifications (user_id, entity_type, entity_id, data, created_at)
			VALUES (?, ?, ?, ?, ?)
		`
		_, err = m.db.Exec(query, mod.User.TelegramID, entityType, entityID, modDataJSON, timestamp)
		if err != nil {
			return fmt.Errorf("failed to insert modification: %v", err)
		}
	}

	return nil
}

func (m *MigrationScript) verify(verbose bool) error {
	printInfo("Verifying migration integrity...")

	tables := []string{"trouble_reports", "metal_sheets", "tools"}
	overallSuccess := true
	totalOldMods := 0
	totalNewMods := 0

	for _, table := range tables {
		oldCount, newCount, err := m.verifyTable(table, verbose)
		if err != nil {
			printError("Verification failed for %s: %v", table, err)
			overallSuccess = false
			continue
		}

		totalOldMods += oldCount
		totalNewMods += newCount

		if oldCount == newCount {
			printSuccess("%s: %d records verified ✓", table, newCount)
		} else {
			printWarning("%s: Mismatch! Old: %d, New: %d", table, oldCount, newCount)
			overallSuccess = false
		}
	}

	fmt.Println()
	fmt.Printf("Total verification: Old mods: %d, New modifications: %d\n", totalOldMods, totalNewMods)

	if overallSuccess && totalNewMods > 0 {
		printSuccess("Verification successful! Migration appears complete and accurate.")
		return nil
	} else if totalOldMods == 0 && totalNewMods == 0 {
		printWarning("No modifications found in either old or new system")
		printInfo("This could indicate: no data to migrate, migration not needed, or cleanup already completed")
		return nil
	} else {
		printError("Verification failed - found discrepancies in migration data")
		return fmt.Errorf("verification failed")
	}
}

func (m *MigrationScript) verifyTable(table string, verbose bool) (int, int, error) {
	// Count records with mods
	var oldCount int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE mods IS NOT NULL AND json_array_length(mods) > 0", table)
	if err := m.db.QueryRow(query).Scan(&oldCount); err != nil {
		// Check if error is due to missing column (cleanup completed)
		if strings.Contains(err.Error(), "no such column: mods") {
			if verbose {
				printInfo("%s: Old mods column not found (cleanup completed)", table)
			}
			oldCount = 0
		} else {
			return 0, 0, fmt.Errorf("failed to count old mods: %v", err)
		}
	}

	// Count corresponding modifications
	var newCount int
	query = "SELECT COUNT(DISTINCT entity_id) FROM modifications WHERE entity_type = ?"
	if err := m.db.QueryRow(query, table).Scan(&newCount); err != nil {
		return 0, 0, fmt.Errorf("failed to count new modifications: %v", err)
	}

	if verbose {
		fmt.Printf("  %s: %d old mods → %d new modifications\n", table, oldCount, newCount)
	}

	return oldCount, newCount, nil
}

func (m *MigrationScript) cleanup(force, dryRun, verbose bool) error {
	printWarning("CLEANUP: This operation will permanently remove old mods columns!")
	printWarning("This is DESTRUCTIVE and cannot be undone!")

	if dryRun {
		printInfo("DRY RUN: Showing what would be cleaned up")
	}

	if !force && !dryRun {
		// Run verification first
		printInfo("Running verification before cleanup...")
		if err := m.verify(verbose); err != nil {
			printError("Verification failed! Cleanup aborted for safety.")
			return fmt.Errorf("verification failed, cleanup aborted: %v", err)
		}
		printSuccess("Verification passed")

		// Ask for confirmation
		fmt.Print("\nAre you absolutely sure you want to remove the old mods columns? (type 'yes' to confirm): ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %v", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "yes" {
			printInfo("Cleanup cancelled by user")
			return nil
		}
	} else if force {
		printWarning("FORCE mode enabled - skipping verification and confirmation")
	}

	queries := []string{
		"ALTER TABLE trouble_reports DROP COLUMN mods",
		"ALTER TABLE metal_sheets DROP COLUMN mods",
		"ALTER TABLE tools DROP COLUMN mods",
	}

	printInfo("Removing old mods columns...")

	for i, query := range queries {
		if verbose || dryRun {
			fmt.Printf("Query %d: %s\n", i+1, query)
		}

		if dryRun {
			continue
		}

		if _, err := m.db.Exec(query); err != nil {
			printWarning("Failed to drop column (expected in older SQLite versions): %v", err)
			printInfo("Consider upgrading SQLite or manually recreating tables without mods columns")
		} else {
			if verbose {
				printSuccess("Successfully executed: %s", query)
			}
		}
	}

	// Drop migration status view since it depends on mods columns
	if !dryRun {
		m.db.Exec("DROP VIEW IF EXISTS migration_status")
	}

	if dryRun {
		printInfo("DRY RUN: Would execute %d cleanup queries", len(queries))
	} else {
		printSuccess("Cleanup completed! Old mods columns have been removed.")
	}

	return nil
}

func (m *MigrationScript) status(verbose bool) error {
	printInfo("Checking migration status...")
	fmt.Println()

	// Database info
	stat, err := os.Stat(m.dbPath)
	if err == nil {
		size := float64(stat.Size()) / 1024 / 1024 // Convert to MB
		fmt.Printf("Database: %s (%.2f MB)\n", m.dbPath, size)
	}

	// Check if modifications table exists
	var modTableExists bool
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name='modifications'`
	var name string
	err = m.db.QueryRow(query).Scan(&name)
	modTableExists = err == nil

	fmt.Printf("Modifications table exists: %v\n", modTableExists)

	var totalMods int
	if modTableExists {
		m.db.QueryRow("SELECT COUNT(*) FROM modifications").Scan(&totalMods)
		fmt.Printf("Total modifications: %d\n", totalMods)

		if totalMods > 0 && verbose {
			// Show modification stats by entity type
			fmt.Println("\nModifications by entity type:")
			query = `
				SELECT entity_type, COUNT(*) as count, COUNT(DISTINCT entity_id) as unique_entities
				FROM modifications
				GROUP BY entity_type
			`
			rows, err := m.db.Query(query)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var entityType string
					var count, uniqueEntities int
					if rows.Scan(&entityType, &count, &uniqueEntities) == nil {
						fmt.Printf("  %s: %d modifications across %d entities\n", entityType, count, uniqueEntities)
					}
				}
			}
		}
	}

	// Check each table for mods columns and data
	fmt.Println()
	tables := []string{"trouble_reports", "metal_sheets", "tools"}
	totalOldMods := 0
	cleanupCompleted := false

	for _, table := range tables {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE mods IS NOT NULL AND json_array_length(mods) > 0", table)
		err := m.db.QueryRow(query).Scan(&count)

		if err != nil {
			if strings.Contains(err.Error(), "no such column: mods") {
				fmt.Printf("%s: Old mods column not found (cleanup completed)\n", table)
				cleanupCompleted = true
			} else {
				fmt.Printf("%s: Error checking mods - %v\n", table, err)
			}
		} else {
			fmt.Printf("%s: %d records with old mods\n", table, count)
			totalOldMods += count
		}
	}

	fmt.Println()

	// Provide status summary and recommendations
	if cleanupCompleted {
		printSuccess("Migration cleanup completed - old mods columns removed")
		if modTableExists && totalMods > 0 {
			printInfo("New modification system is active with %d modifications", totalMods)
		}
	} else if modTableExists && totalMods > 0 && totalOldMods == 0 {
		printSuccess("Migration completed successfully!")
		printInfo("Ready for cleanup. Run: go run migrate_mods.go -action=cleanup")
	} else if modTableExists && totalMods > 0 && totalOldMods > 0 {
		printWarning("Both old and new systems detected!")
		printInfo("Run verification: go run migrate_mods.go -action=verify")
	} else if !modTableExists || totalMods == 0 {
		if totalOldMods > 0 {
			printInfo("Migration needed - found %d records with old mods", totalOldMods)
			printInfo("Run: go run migrate_mods.go -action=full")
		} else {
			printInfo("No migration needed - no data found in old or new systems")
		}
	}

	return nil
}

func isColumnExistsError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "duplicate column name: mods") ||
		strings.Contains(errStr, "table trouble_reports has no column named mods") ||
		strings.Contains(errStr, "table metal_sheets has no column named mods") ||
		strings.Contains(errStr, "table tools has no column named mods")
}
