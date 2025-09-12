package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/knackwurstking/pgpress/internal/services"

	"github.com/SuperPaintman/nice/cli"
)

// NOTE: This command will become obsolete if version 1.0.0 is finished.
func migrationStatusCommand() cli.Command {
	return cli.Command{
		Name:  "status",
		Usage: cli.Usage("Show current migration status"),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := cli.String(cmd, "db",
				cli.WithShort("d"),
				cli.Usage("Custom database path"),
				cli.Optional,
			)

			return func(cmd *cli.Command) error {
				db, err := openDB(*customDBPath)
				if err != nil {
					return err
				}
				defer db.GetDB().Close()

				users := services.NewUser(db.GetDB(), db.Feeds)
				migrationCLI := services.NewModificationCLI(db.GetDB(), users)

				return migrationCLI.RunCommand("status", []string{})
			}
		}),
	}
}

func migrationRunCommand() cli.Command {
	return cli.Command{
		Name:  "run",
		Usage: cli.Usage("Run the migration from old mod system to new modification system"),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := cli.String(cmd, "db",
				cli.WithShort("d"),
				cli.Usage("Custom database path"),
				cli.Optional,
			)

			return func(cmd *cli.Command) error {
				db, err := openDB(*customDBPath)
				if err != nil {
					return err
				}
				defer db.GetDB().Close()

				users := services.NewUser(db.GetDB(), db.Feeds)
				migrationCLI := services.NewModificationCLI(db.GetDB(), users)

				return migrationCLI.RunCommand("migrate", []string{})
			}
		}),
	}
}

func migrationVerifyCommand() cli.Command {
	return cli.Command{
		Name:  "verify",
		Usage: cli.Usage("Verify migration integrity by comparing old and new data"),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := cli.String(cmd, "db",
				cli.WithShort("d"),
				cli.Usage("Custom database path"),
				cli.Optional,
			)

			return func(cmd *cli.Command) error {
				db, err := openDB(*customDBPath)
				if err != nil {
					return err
				}
				defer db.GetDB().Close()

				users := services.NewUser(db.GetDB(), db.Feeds)
				migrationCLI := services.NewModificationCLI(db.GetDB(), users)

				return migrationCLI.RunCommand("verify", []string{})
			}
		}),
	}
}

func migrationCleanupCommand() cli.Command {
	return cli.Command{
		Name:  "cleanup",
		Usage: cli.Usage("Remove old mod columns (DESTRUCTIVE - use after verification)"),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := cli.String(cmd, "db",
				cli.WithShort("d"),
				cli.Usage("Custom database path"),
				cli.Optional,
			)

			force := cli.Bool(cmd, "force",
				cli.WithShort("f"),
				cli.Usage("Force cleanup without interactive confirmation"),
				cli.Optional,
			)

			return func(cmd *cli.Command) error {
				db, err := openDB(*customDBPath)
				if err != nil {
					return err
				}
				defer db.GetDB().Close()

				users := services.NewUser(db.GetDB(), db.Feeds)
				migrationCLI := services.NewModificationCLI(db.GetDB(), users)

				if *force {
					fmt.Println("⚠️  WARNING: Force cleanup requested!")
					fmt.Println("This will skip safety checks and proceed with cleanup.")
					return migrationCLI.RunCommand("cleanup", []string{"--force"})
				}

				return migrationCLI.RunCommand("cleanup", []string{})
			}
		}),
	}
}

func migrationStatsCommand() cli.Command {
	return cli.Command{
		Name:  "stats",
		Usage: cli.Usage("Display modification system statistics"),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := cli.String(cmd, "db",
				cli.WithShort("d"),
				cli.Usage("Custom database path"),
				cli.Optional,
			)

			output := cli.String(cmd, "output",
				cli.WithShort("o"),
				cli.Usage("Output format (json|text)"),
				cli.Optional,
			)
			*output = "text"

			return func(cmd *cli.Command) error {
				db, err := openDB(*customDBPath)
				if err != nil {
					return err
				}
				defer db.GetDB().Close()

				users := services.NewUser(db.GetDB(), db.Feeds)
				migrationCLI := services.NewModificationCLI(db.GetDB(), users)

				if *output == "json" {
					// For JSON output, we need to capture the stats and output as JSON
					// This is a simplified approach - in practice you might want to modify
					// the CLI to return structured data
					fmt.Println("JSON output format not yet implemented - using text format")
				}

				return migrationCLI.RunCommand("stats", []string{})
			}
		}),
	}
}

func migrationExportCommand() cli.Command {
	return cli.Command{
		Name:  "export",
		Usage: cli.Usage("Export migration data to a file"),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := cli.String(cmd, "db",
				cli.WithShort("d"),
				cli.Usage("Custom database path"),
				cli.Optional,
			)

			outputFile := cli.String(cmd, "output",
				cli.WithShort("o"),
				cli.Usage("Output file path"),
			)
			*outputFile = "migration_export.json"

			entityType := cli.String(cmd, "entity",
				cli.WithShort("e"),
				cli.Usage("Entity type to export (trouble_reports|metal_sheets|tools|all)"),
				cli.Optional,
			)
			*entityType = "all"

			return func(cmd *cli.Command) error {
				db, err := openDB(*customDBPath)
				if err != nil {
					return err
				}
				defer db.GetDB().Close()

				fmt.Printf("Exporting migration data to %s...\n", *outputFile)

				// Query modifications from database
				query := `
					SELECT id, user_id, entity_type, entity_id, data, created_at
					FROM modifications
				`
				args := []interface{}{}

				if *entityType != "all" {
					query += " WHERE entity_type = ?"
					args = append(args, *entityType)
				}

				query += " ORDER BY created_at DESC"

				rows, err := db.GetDB().Query(query, args...)
				if err != nil {
					return fmt.Errorf("failed to query modifications: %w", err)
				}
				defer rows.Close()

				var modifications []map[string]interface{}
				for rows.Next() {
					var id, userID, entityID int64
					var entityType, data string
					var createdAt string

					if err := rows.Scan(&id, &userID, &entityType, &entityID, &data, &createdAt); err != nil {
						fmt.Fprintf(os.Stderr, "Error scanning row: %v\n", err)
						continue
					}

					// Parse JSON data
					var jsonData interface{}
					if err := json.Unmarshal([]byte(data), &jsonData); err != nil {
						fmt.Fprintf(os.Stderr, "Error parsing JSON data for modification %d: %v\n", id, err)
						continue
					}

					modification := map[string]interface{}{
						"id":          id,
						"user_id":     userID,
						"entity_type": entityType,
						"entity_id":   entityID,
						"data":        jsonData,
						"created_at":  createdAt,
					}

					modifications = append(modifications, modification)
				}

				if err = rows.Err(); err != nil {
					return fmt.Errorf("error iterating rows: %w", err)
				}

				// Create export data structure
				exportData := map[string]interface{}{
					"exported_at":   time.Now().Format("2006-01-02 15:04:05"),
					"entity_type":   *entityType,
					"total_count":   len(modifications),
					"modifications": modifications,
				}

				// Write to file
				data, err := json.MarshalIndent(exportData, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal export data: %w", err)
				}

				if err := os.WriteFile(*outputFile, data, 0644); err != nil {
					return fmt.Errorf("failed to write export file: %w", err)
				}

				fmt.Printf("Successfully exported %d modifications to %s\n", len(modifications), *outputFile)
				return nil
			}
		}),
	}
}

func migrationTestCommand() cli.Command {
	return cli.Command{
		Name:  "test-db",
		Usage: cli.Usage("Test database connection and configuration"),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := cli.String(cmd, "db",
				cli.WithShort("d"),
				cli.Usage("Custom database path"),
				cli.Optional,
			)

			return func(cmd *cli.Command) error {
				fmt.Println("=== Testing Database Connection ===")

				if err := testDBConnection(*customDBPath); err != nil {
					fmt.Printf("❌ Database connection test failed: %v\n", err)
					fmt.Println("\nTroubleshooting tips:")
					fmt.Println("1. Ensure the database file exists and is not corrupted")
					fmt.Println("2. Check file permissions on the database file")
					fmt.Println("3. Make sure no other processes are using the database")
					fmt.Println("4. Try running with a different database path using --db flag")
					return err
				}

				fmt.Println("✅ Database connection test passed!")
				fmt.Println("The database is properly configured and accessible.")
				return nil
			}
		}),
	}
}

func migrationHelpCommand() cli.Command {
	return cli.Command{
		Name:  "help",
		Usage: cli.Usage("Show detailed help for migration commands"),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			return func(cmd *cli.Command) error {
				fmt.Println("=== Modification System Migration Help ===")
				fmt.Println()
				fmt.Println("The modification system migration helps you transition from the old")
				fmt.Println("'mods' column-based system to the new centralized modification service.")
				fmt.Println()
				fmt.Println("Available commands:")
				fmt.Println()
				fmt.Println("  status    - Show current migration status")
				fmt.Println("            Shows whether migration is needed and current state")
				fmt.Println()
				fmt.Println("  test-db   - Test database connection and configuration")
				fmt.Println("            Verifies database connectivity and WAL mode setup")
				fmt.Println()
				fmt.Println("  run       - Execute the migration process")
				fmt.Println("            Migrates data from old 'mods' columns to new system")
				fmt.Println("            Options: --force to skip confirmation")
				fmt.Println()
				fmt.Println("  verify    - Verify migration integrity")
				fmt.Println("            Compares old and new data to ensure accuracy")
				fmt.Println()
				fmt.Println("  stats     - Display modification system statistics")
				fmt.Println("            Shows counts and activity across the system")
				fmt.Println()
				fmt.Println("  export    - Export migration data to JSON file")
				fmt.Println("            Options: --entity to filter by type, --output for filename")
				fmt.Println()
				fmt.Println("  cleanup   - Remove old mod columns (DESTRUCTIVE)")
				fmt.Println("            Only run after successful migration and verification!")
				fmt.Println("            Options: --force to skip verification and confirmation")
				fmt.Println()
				fmt.Println("Recommended workflow:")
				fmt.Println("  1. pgpress migration test-db   # Test database connection")
				fmt.Println("  2. pgpress migration status    # Check if migration is needed")
				fmt.Println("  3. pgpress migration run       # Perform the migration")
				fmt.Println("  4. pgpress migration verify    # Verify migration success")
				fmt.Println("  5. pgpress migration stats     # Review statistics")
				fmt.Println("  6. Test your application thoroughly")
				fmt.Println("  7. pgpress migration cleanup   # Remove old columns (optional)")
				fmt.Println()
				fmt.Println("Safety notes:")
				fmt.Println("  - Always backup your database before running migrations")
				fmt.Println("  - The 'cleanup' command is destructive and cannot be undone")
				fmt.Println("  - Test thoroughly before running cleanup")
				fmt.Println("  - Migration can be run multiple times safely")
				fmt.Println()

				return nil
			}
		}),
	}
}
