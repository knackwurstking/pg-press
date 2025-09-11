package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// ModificationCLI provides command-line interface for modification service operations
type ModificationCLI struct {
	db            *sql.DB
	modifications *ModificationService
	migration     *ModificationMigration
	users         *User
}

// NewModificationCLI creates a new CLI instance
func NewModificationCLI(db *sql.DB, users *User) *ModificationCLI {
	modifications := NewModificationService(db)
	migration := NewModificationMigration(db, modifications, users)

	return &ModificationCLI{
		db:            db,
		modifications: modifications,
		migration:     migration,
		users:         users,
	}
}

// RunCommand executes the specified command
func (cli *ModificationCLI) RunCommand(command string, args []string) error {
	switch command {
	case "status":
		return cli.showStatus()
	case "migrate":
		return cli.runMigration()
	case "verify":
		return cli.verifyMigration()
	case "cleanup":
		return cli.cleanupOldMods()
	case "stats":
		return cli.showStats()
	case "help":
		cli.showHelp()
		return nil
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

// showStatus displays the current migration status
func (cli *ModificationCLI) showStatus() error {
	fmt.Println("=== Modification System Status ===")

	status, err := cli.migration.GetMigrationStatus()
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	fmt.Printf("Modification table exists: %v\n", status.ModificationTableExists)
	fmt.Printf("Total modifications: %d\n", status.TotalModifications)
	fmt.Printf("Old mods still exist: %v\n", status.OldModsExist)
	fmt.Printf("Migration needed: %v\n", status.MigrationNeeded)

	if status.MigrationNeeded {
		fmt.Println("\n⚠️  Migration is recommended!")
		fmt.Println("Run 'migrate' command to start the migration process.")
	} else if status.TotalModifications > 0 && !status.OldModsExist {
		fmt.Println("\n✅ Migration appears to be complete!")
	} else if status.TotalModifications > 0 && status.OldModsExist {
		fmt.Println("\n⚠️  Both old and new systems detected!")
		fmt.Println("Consider running 'verify' to check migration integrity.")
	}

	return nil
}

// runMigration executes the full migration process
func (cli *ModificationCLI) runMigration() error {
	fmt.Println("=== Starting Migration Process ===")

	// Check if migration is needed
	status, err := cli.migration.GetMigrationStatus()
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if !status.MigrationNeeded && status.TotalModifications > 0 {
		fmt.Println("Migration may have already been completed.")
		fmt.Print("Do you want to continue anyway? (y/N): ")

		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Migration cancelled.")
			return nil
		}
	}

	fmt.Println("Starting migration...")
	startTime := time.Now()

	stats, err := cli.migration.MigrateAll()
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Println("\n=== Migration Complete ===")
	fmt.Printf("Duration: %v\n", time.Since(startTime))
	fmt.Printf("Trouble reports processed: %d\n", stats.TroubleReportsProcessed)
	fmt.Printf("Metal sheets processed: %d\n", stats.MetalSheetsProcessed)
	fmt.Printf("Tools processed: %d\n", stats.ToolsProcessed)
	fmt.Printf("Total modifications migrated: %d\n", stats.TotalModsMigrated)

	if stats.Errors > 0 {
		fmt.Printf("⚠️  Errors encountered: %d\n", stats.Errors)
		fmt.Println("Check logs for details.")
	} else {
		fmt.Println("✅ Migration completed successfully with no errors!")
	}

	// Save migration stats to file
	if err := cli.saveMigrationStats(stats); err != nil {
		fmt.Printf("Warning: Failed to save migration stats: %v\n", err)
	}

	fmt.Println("\nNext steps:")
	fmt.Println("1. Run 'verify' to validate the migration")
	fmt.Println("2. Test your application with the new modification system")
	fmt.Println("3. Once satisfied, run 'cleanup' to remove old mod columns")

	return nil
}

// verifyMigration checks the integrity of the migration
func (cli *ModificationCLI) verifyMigration() error {
	fmt.Println("=== Verifying Migration ===")

	result, err := cli.migration.VerifyMigration()
	if err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	fmt.Printf("Trouble Reports - Old: %d, New: %d, Match: %v\n",
		result.TroubleReports.OldCount, result.TroubleReports.NewCount, result.TroubleReports.Match)
	fmt.Printf("Metal Sheets - Old: %d, New: %d, Match: %v\n",
		result.MetalSheets.OldCount, result.MetalSheets.NewCount, result.MetalSheets.Match)
	fmt.Printf("Tools - Old: %d, New: %d, Match: %v\n",
		result.Tools.OldCount, result.Tools.NewCount, result.Tools.Match)

	if result.OverallMatch {
		fmt.Println("✅ Verification successful! Migration appears to be complete and accurate.")
	} else {
		fmt.Println("⚠️  Verification found discrepancies. Please review the migration.")
	}

	return nil
}

// cleanupOldMods removes old mod columns (destructive operation)
func (cli *ModificationCLI) cleanupOldMods() error {
	fmt.Println("=== Cleanup Old Mod System ===")
	fmt.Println("⚠️  WARNING: This operation is DESTRUCTIVE and cannot be undone!")
	fmt.Println("It will remove the old 'mods' columns from your database tables.")
	fmt.Println()

	// Verify migration first
	fmt.Println("Verifying migration before cleanup...")
	result, err := cli.migration.VerifyMigration()
	if err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	if !result.OverallMatch {
		fmt.Println("❌ Verification failed! Cleanup aborted for safety.")
		fmt.Println("Please resolve migration issues before attempting cleanup.")
		return fmt.Errorf("verification failed, cleanup aborted")
	}

	fmt.Println("✅ Verification passed.")
	fmt.Print("Are you absolutely sure you want to proceed with cleanup? (type 'yes' to confirm): ")

	var response string
	fmt.Scanln(&response)
	if response != "yes" {
		fmt.Println("Cleanup cancelled.")
		return nil
	}

	fmt.Println("Performing cleanup...")
	if err := cli.migration.CleanupOldMods(); err != nil {
		return fmt.Errorf("cleanup failed: %w", err)
	}

	fmt.Println("✅ Cleanup completed successfully!")
	fmt.Println("The old mod system has been removed from your database.")

	return nil
}

// showStats displays various statistics about the modification system
func (cli *ModificationCLI) showStats() error {
	fmt.Println("=== Modification System Statistics ===")

	// Total modifications
	var totalMods int
	err := cli.db.QueryRow("SELECT COUNT(*) FROM modifications").Scan(&totalMods)
	if err != nil {
		return fmt.Errorf("failed to count total modifications: %w", err)
	}
	fmt.Printf("Total modifications: %d\n", totalMods)

	// Modifications by entity type
	entityTypes := []ModificationType{
		ModificationTypeTroubleReport,
		ModificationTypeMetalSheet,
		ModificationTypeTool,
		ModificationTypePressCycle,
		ModificationTypeUser,
	}

	fmt.Println("\nModifications by entity type:")
	for _, entityType := range entityTypes {
		var count int
		err := cli.db.QueryRow(
			"SELECT COUNT(*) FROM modifications WHERE entity_type = ?",
			string(entityType),
		).Scan(&count)
		if err != nil {
			fmt.Printf("  %s: Error (%v)\n", entityType, err)
		} else {
			fmt.Printf("  %s: %d\n", entityType, count)
		}
	}

	// Recent activity (last 7 days)
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	var recentCount int
	err = cli.db.QueryRow(
		"SELECT COUNT(*) FROM modifications WHERE created_at > ?",
		sevenDaysAgo,
	).Scan(&recentCount)
	if err != nil {
		fmt.Printf("Recent activity (7 days): Error (%v)\n", err)
	} else {
		fmt.Printf("Recent activity (7 days): %d\n", recentCount)
	}

	// Most active users
	fmt.Println("\nTop 5 most active users:")
	rows, err := cli.db.Query(`
		SELECT u.user_name, COUNT(*) as mod_count
		FROM modifications m
		JOIN users u ON m.user_id = u.telegram_id
		GROUP BY m.user_id, u.user_name
		ORDER BY mod_count DESC
		LIMIT 5
	`)
	if err != nil {
		fmt.Printf("Error getting user stats: %v\n", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var username string
			var count int
			if err := rows.Scan(&username, &count); err == nil {
				fmt.Printf("  %s: %d modifications\n", username, count)
			}
		}
	}

	return nil
}

// saveMigrationStats saves migration statistics to a JSON file
func (cli *ModificationCLI) saveMigrationStats(stats *MigrationStats) error {
	filename := fmt.Sprintf("migration_stats_%s.json", time.Now().Format("2006-01-02_15-04-05"))

	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write stats file: %w", err)
	}

	fmt.Printf("Migration stats saved to: %s\n", filename)
	return nil
}

// showHelp displays available commands
func (cli *ModificationCLI) showHelp() {
	fmt.Println("=== Modification System CLI ===")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println()
	fmt.Println("  status    - Show current migration status")
	fmt.Println("  migrate   - Run the migration from old mod system to new modification system")
	fmt.Println("  verify    - Verify migration integrity by comparing old and new data")
	fmt.Println("  cleanup   - Remove old mod columns (DESTRUCTIVE - use after verification)")
	fmt.Println("  stats     - Display modification system statistics")
	fmt.Println("  help      - Show this help message")
	fmt.Println()
	fmt.Println("Usage examples:")
	fmt.Println("  ./app modification status")
	fmt.Println("  ./app modification migrate")
	fmt.Println("  ./app modification verify")
	fmt.Println("  ./app modification cleanup")
	fmt.Println()
	fmt.Println("Migration workflow:")
	fmt.Println("  1. Run 'status' to check if migration is needed")
	fmt.Println("  2. Run 'migrate' to perform the migration")
	fmt.Println("  3. Run 'verify' to ensure migration was successful")
	fmt.Println("  4. Test your application thoroughly")
	fmt.Println("  5. Run 'cleanup' to remove old mod columns (optional)")
}

// ExampleMain shows how to integrate this CLI into your main application
func ExampleMain() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./app modification <command>")
		fmt.Println("Run './app modification help' for available commands")
		os.Exit(1)
	}

	// Initialize database connection (replace with your actual DB setup)
	db, err := sql.Open("sqlite3", "your_database.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	users := NewUser(db, nil) // Assuming feeds is nil for CLI usage
	cli := NewModificationCLI(db, users)

	command := os.Args[1]
	args := os.Args[2:]

	if err := cli.RunCommand(command, args); err != nil {
		log.Fatalf("Command failed: %v", err)
	}
}
