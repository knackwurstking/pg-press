package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: go run add-indexes-new.go <database-file-path>")
	}

	dbPath := os.Args[1]

	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_notes_linked ON notes(linked)",
		"CREATE INDEX IF NOT EXISTS idx_tool_regenerations_tool_id ON tool_regenerations(tool_id)",
	}

	for _, query := range indexes {
		fmt.Printf("Executing: %s\n", query)
		_, err := db.Exec(query)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Successfully created indexes")
}
