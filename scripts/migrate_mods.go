package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
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

type ModificationData struct {
	Action    string                 `json:"action"`
	Timestamp time.Time              `json:"timestamp"`
	UserID    int64                  `json:"user_id"`
	Data      map[string]interface{} `json:"data"`
}

func main() {
	var (
		dbPath  = flag.String("db", "./data.db", "Path to SQLite database file")
		action  = flag.String("action", "migrate", "Action to perform: migrate, verify, cleanup, status")
		force   = flag.Bool("force", false, "Force operation without confirmation")
		dryRun  = flag.Bool("dry-run", false, "Show what would be done without making changes")
		verbose = flag.Bool("v", false, "Verbose output")
	)
	flag.Parse()

	script := &MigrationScript{dbPath: *dbPath}

	if err := script.connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer script.db.Close()

	switch *action {
	case "migrate":
		if err := script.migrate(*dryRun, *verbose); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
	case "verify":
		if err := script.verify(); err != nil {
			log.Fatalf("Verification failed: %v", err)
		}
	case "cleanup":
		if err := script.cleanup(*force, *dryRun); err != nil {
			log.Fatalf("Cleanup failed: %v", err)
		}
	case "status":
		if err := script.status(); err != nil {
			log.Fatalf("Status check failed: %v", err)
		}
	default:
		log.Fatalf("Unknown action: %s. Use: migrate, verify, cleanup, or status", *action)
	}
}

func (m *MigrationScript) connect() error {
	if _, err := os.Stat(m.dbPath); os.IsNotExist(err) {
		return fmt.Errorf("database file does not exist: %s", m.dbPath)
	}

	db, err := sql.Open("sqlite3", m.dbPath+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	m.db = db
	return nil
}

func (m *MigrationScript) migrate(dryRun, verbose bool) error {
	fmt.Println("=== Starting Modification Migration ===")

	stats := &MigrationStats{StartTime: time.Now()}

	// Step 1: Create schema if needed
	if err := m.createSchema(dryRun); err != nil {
		return fmt.Errorf("failed to create schema: %v", err)
	}

	// Step 2: Migrate each table
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

	fmt.Printf("\n=== Migration Complete ===\n")
	fmt.Printf("Trouble Reports processed: %d\n", stats.TroubleReportsProcessed)
	fmt.Printf("Metal Sheets processed: %d\n", stats.MetalSheetsProcessed)
	fmt.Printf("Tools processed: %d\n", stats.ToolsProcessed)
	fmt.Printf("Total modifications migrated: %d\n", stats.TotalModsMigrated)
	fmt.Printf("Errors: %d\n", stats.Errors)
	fmt.Printf("Duration: %v\n", stats.Duration)

	if dryRun {
		fmt.Println("\n🔍 DRY RUN - No changes were made to the database")
	} else {
		fmt.Println("\n✅ Migration completed successfully!")
	}

	return nil
}

func (m *MigrationScript) createSchema(dryRun bool) error {
	fmt.Println("Creating/updating database schema...")

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
	}

	for _, query := range queries {
		if dryRun {
			fmt.Printf("WOULD EXECUTE: %s\n", query)
			continue
		}

		if _, err := m.db.Exec(query); err != nil {
			// Ignore column already exists errors
			if !isColumnExistsError(err) {
				return fmt.Errorf("failed to execute query %s: %v", query, err)
			}
		}
	}

	return nil
}

func (m *MigrationScript) migrateTroubleReports(stats *MigrationStats, dryRun, verbose bool) error {
	fmt.Println("Migrating trouble report modifications...")

	query := `SELECT id, mods FROM trouble_reports WHERE mods IS NOT NULL AND mods != '[]' AND mods != ''`

	rows, err := m.db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query trouble reports: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var modsJSON string

		if err := rows.Scan(&id, &modsJSON); err != nil {
			fmt.Printf("Error scanning trouble report %d: %v\n", id, err)
			stats.Errors++
			continue
		}

		if err := m.processMods(id, "trouble_reports", modsJSON, dryRun, verbose); err != nil {
			fmt.Printf("Error processing mods for trouble report %d: %v\n", id, err)
			stats.Errors++
			continue
		}

		stats.TroubleReportsProcessed++
	}

	return rows.Err()
}

func (m *MigrationScript) migrateMetalSheets(stats *MigrationStats, dryRun, verbose bool) error {
	fmt.Println("Migrating metal sheet modifications...")

	query := `SELECT id, mods FROM metal_sheets WHERE mods IS NOT NULL AND mods != '[]' AND mods != ''`

	rows, err := m.db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query metal sheets: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var modsJSON string

		if err := rows.Scan(&id, &modsJSON); err != nil {
			fmt.Printf("Error scanning metal sheet %d: %v\n", id, err)
			stats.Errors++
			continue
		}

		if err := m.processMods(id, "metal_sheets", modsJSON, dryRun, verbose); err != nil {
			fmt.Printf("Error processing mods for metal sheet %d: %v\n", id, err)
			stats.Errors++
			continue
		}

		stats.MetalSheetsProcessed++
	}

	return rows.Err()
}

func (m *MigrationScript) migrateTools(stats *MigrationStats, dryRun, verbose bool) error {
	fmt.Println("Migrating tool modifications...")

	query := `SELECT id, mods FROM tools WHERE mods IS NOT NULL AND mods != '[]' AND mods != ''`

	rows, err := m.db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query tools: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var modsJSON string

		if err := rows.Scan(&id, &modsJSON); err != nil {
			fmt.Printf("Error scanning tool %d: %v\n", id, err)
			stats.Errors++
			continue
		}

		if err := m.processMods(id, "tools", modsJSON, dryRun, verbose); err != nil {
			fmt.Printf("Error processing mods for tool %d: %v\n", id, err)
			stats.Errors++
			continue
		}

		stats.ToolsProcessed++
	}

	return rows.Err()
}

func (m *MigrationScript) processMods(entityID int64, entityType, modsJSON string, dryRun, verbose bool) error {
	var mods []ModificationData
	if err := json.Unmarshal([]byte(modsJSON), &mods); err != nil {
		return fmt.Errorf("failed to unmarshal mods JSON: %v", err)
	}

	for _, mod := range mods {
		if verbose {
			fmt.Printf("Processing mod for %s %d: action=%s, user=%d\n",
				entityType, entityID, mod.Action, mod.UserID)
		}

		if dryRun {
			fmt.Printf("WOULD INSERT: entity_type=%s, entity_id=%d, user_id=%d\n",
				entityType, entityID, mod.UserID)
			continue
		}

		// Convert modification data to JSON
		modDataJSON, err := json.Marshal(mod.Data)
		if err != nil {
			return fmt.Errorf("failed to marshal mod data: %v", err)
		}

		// Insert into modifications table
		query := `
			INSERT INTO modifications (user_id, entity_type, entity_id, data, created_at)
			VALUES (?, ?, ?, ?, ?)
		`
		_, err = m.db.Exec(query, mod.UserID, entityType, entityID, modDataJSON, mod.Timestamp)
		if err != nil {
			return fmt.Errorf("failed to insert modification: %v", err)
		}
	}

	return nil
}

func (m *MigrationScript) verify() error {
	fmt.Println("=== Verifying Migration ===")

	// Check each table
	tables := []string{"trouble_reports", "metal_sheets", "tools"}
	overallSuccess := true

	for _, table := range tables {
		if err := m.verifyTable(table); err != nil {
			fmt.Printf("❌ Verification failed for %s: %v\n", table, err)
			overallSuccess = false
		} else {
			fmt.Printf("✅ %s verification passed\n", table)
		}
	}

	if overallSuccess {
		fmt.Println("\n✅ Overall verification: PASSED")
	} else {
		fmt.Println("\n❌ Overall verification: FAILED")
		return fmt.Errorf("verification failed for one or more tables")
	}

	return nil
}

func (m *MigrationScript) verifyTable(table string) error {
	// Count records with mods
	var oldCount int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE mods IS NOT NULL AND mods != '[]' AND mods != ''", table)
	if err := m.db.QueryRow(query).Scan(&oldCount); err != nil {
		return fmt.Errorf("failed to count old mods: %v", err)
	}

	// Count corresponding modifications
	var newCount int
	query = "SELECT COUNT(DISTINCT entity_id) FROM modifications WHERE entity_type = ?"
	if err := m.db.QueryRow(query, table).Scan(&newCount); err != nil {
		return fmt.Errorf("failed to count new modifications: %v", err)
	}

	fmt.Printf("  %s: %d records with mods, %d entities in modifications\n", table, oldCount, newCount)

	if oldCount != newCount {
		return fmt.Errorf("count mismatch: expected %d, got %d", oldCount, newCount)
	}

	return nil
}

func (m *MigrationScript) cleanup(force, dryRun bool) error {
	fmt.Println("=== Cleanup Old Mods Columns ===")
	fmt.Println("⚠️  WARNING: This operation is DESTRUCTIVE and cannot be undone!")

	if !force && !dryRun {
		fmt.Print("Are you sure you want to proceed? (yes/no): ")
		var response string
		fmt.Scanln(&response)
		if response != "yes" {
			fmt.Println("Operation cancelled.")
			return nil
		}

		// Run verification first
		if err := m.verify(); err != nil {
			return fmt.Errorf("verification failed, aborting cleanup: %v", err)
		}
	}

	queries := []string{
		"ALTER TABLE trouble_reports DROP COLUMN mods",
		"ALTER TABLE metal_sheets DROP COLUMN mods",
		"ALTER TABLE tools DROP COLUMN mods",
	}

	for _, query := range queries {
		if dryRun {
			fmt.Printf("WOULD EXECUTE: %s\n", query)
			continue
		}

		if _, err := m.db.Exec(query); err != nil {
			fmt.Printf("⚠️  Failed to drop column (this is expected in older SQLite): %v\n", err)
			fmt.Println("   Consider upgrading SQLite or manually recreating tables without mods columns")
		} else {
			fmt.Printf("✅ Successfully executed: %s\n", query)
		}
	}

	if dryRun {
		fmt.Println("\n🔍 DRY RUN - No changes were made to the database")
	} else {
		fmt.Println("\n✅ Cleanup completed!")
	}

	return nil
}

func (m *MigrationScript) status() error {
	fmt.Println("=== Migration Status ===")

	// Check if modifications table exists
	var modTableExists bool
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name='modifications'`
	var name string
	err := m.db.QueryRow(query).Scan(&name)
	modTableExists = err == nil

	fmt.Printf("Modifications table exists: %v\n", modTableExists)

	if modTableExists {
		var totalMods int
		m.db.QueryRow("SELECT COUNT(*) FROM modifications").Scan(&totalMods)
		fmt.Printf("Total modifications: %d\n", totalMods)
	}

	// Check each table for mods columns
	tables := []string{"trouble_reports", "metal_sheets", "tools"}
	for _, table := range tables {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE mods IS NOT NULL AND mods != '[]' AND mods != ''", table)
		if err := m.db.QueryRow(query).Scan(&count); err != nil {
			fmt.Printf("%s: Error checking mods - %v\n", table, err)
		} else {
			fmt.Printf("%s: %d records with mods\n", table, count)
		}
	}

	if modTableExists {
		// Show modification stats by entity type
		fmt.Println("\nModification statistics:")
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

	return nil
}

func isColumnExistsError(err error) bool {
	return err != nil && (err.Error() == "duplicate column name: mods" ||
		err.Error() == "table trouble_reports has no column named mods" ||
		err.Error() == "table metal_sheets has no column named mods" ||
		err.Error() == "table tools has no column named mods")
}
