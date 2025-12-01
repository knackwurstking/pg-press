package main

import (
	"database/sql"
	"fmt"
	"os"
	"slices"

	_ "github.com/mattn/go-sqlite3"
)

const (
	GenericExitCode  = 1
	NotFoundExitCode = 2
	DatabaseExitCode = 3

	// ANSI color codes
	Red = "\033[31m"
	//Green   = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	//Magenta = "\033[35m"
	//Cyan    = "\033[36m"
	//White   = "\033[37m"
	Reset = "\033[0m"
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
		fmt.Fprintf(os.Stderr, "%sUsage: go run main.go [-v/--verbose] <database-path>%s\n", Blue, Reset)
		os.Exit(GenericExitCode)
	}

	dbPath := os.Args[1]

	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sFailed to open database: %v%s\n", Red, err, Reset)
		os.Exit(DatabaseExitCode)
	}
	defer db.Close()

	// Check if database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "%sDatabase file does not exist: %s%s\n", Red, dbPath, Reset)
		os.Exit(DatabaseExitCode)
	}

	// Get list of all indexes in the database
	indexes, err := getIndexes(db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sFailed to get indexes: %v%s\n", Red, err, Reset)
		os.Exit(DatabaseExitCode)
	}

	lindexes := len(indexes)
	if lindexes == 0 {
		fmt.Fprintf(os.Stderr, "%sNo indexes found in database%s\n", Yellow, Reset)
		os.Exit(NotFoundExitCode)
	}
	if verbose {
		fmt.Printf("Found %d indexes to remove\n", lindexes)
	}

	// Drop each index
	for _, index := range indexes {
		if verbose {
			fmt.Printf("Dropping index \"%s\" from \"%s\"\n", index.Name, index.Table)
		}

		_, err := db.Exec(fmt.Sprintf("DROP INDEX IF EXISTS %s", index.Name))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%sFailed to drop index \"%s\" from table \"%s\": %v%s\n", Red, index.Name, index.Table, err, Reset)
			continue
		}

		if verbose {
			fmt.Printf("Successfully dropped index \"%s\" from table \"%s\"\n", index.Name, index.Table)
		}
	}

	if verbose {
		fmt.Println("All indexes have been removed from the database")
	}
}

type IndexInfo struct {
	Name  string
	Table string
}

func getIndexes(db *sql.DB) ([]IndexInfo, error) {
	rows, err := db.Query(`
		SELECT name, tbl_name 
		FROM sqlite_master 
		WHERE type = 'index' AND name NOT LIKE 'sqlite_%'
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexes []IndexInfo
	for rows.Next() {
		var index IndexInfo
		if err := rows.Scan(&index.Name, &index.Table); err != nil {
			return nil, err
		}
		indexes = append(indexes, index)
	}

	return indexes, rows.Err()
}
