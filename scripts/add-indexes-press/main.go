package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"slices"

	"github.com/knackwurstking/pg-press/scripts/constants"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	verbose := false

	switch len(os.Args) {
	case 3:
		// Check for verbose flag
		for i, a := range os.Args {
			if a == "-v" || a == "--verbose" {
				verbose = true
				os.Args = slices.Delete(os.Args, i, i+1)
				break
			}
		}
	case 2:
	default:
		fmt.Fprintf(os.Stderr, "%sUsage: go run %s [-v/--verbose] <database-path>%s\n", constants.Blue, os.Args[0], constants.Reset)
		os.Exit(constants.GenericExitCode)
	}

	dbPath := os.Args[1]

	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sFailed to open database: %v%s\n", constants.Red, err, constants.Reset)
		os.Exit(constants.DatabaseExitCode)
	}
	defer db.Close()

	// Create indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_notes_linked ON notes(linked)",
		"CREATE INDEX IF NOT EXISTS idx_tool_regenerations_tool_id ON tool_regenerations(tool_id)",
		"CREATE INDEX IF NOT EXISTS idx_press_cycles_tool_id ON press_cycles(tool_id)",
		"CREATE INDEX IF NOT EXISTS idx_press_cycles_press_number ON press_cycles(press_number)",
		"CREATE INDEX IF NOT EXISTS idx_press_cycles_date ON press_cycles(date)",
	}

	for _, query := range indexes {
		if verbose {
			fmt.Fprintf(os.Stderr, "Executing: %s\n", query)
		}
		_, err := db.Exec(query)
		if err != nil {
			log.Fatal(err)
		}
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "%sSuccessfully created indexes%s\n", constants.Green, constants.Reset)
	}
}
