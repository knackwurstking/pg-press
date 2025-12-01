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
		log.Fatal("Usage: go run add-indexes-press.go <database-file-path>")
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
		"CREATE INDEX IF NOT EXISTS idx_press_cycles_tool_id ON press_cycles(tool_id)",
		"CREATE INDEX IF NOT EXISTS idx_press_cycles_press_number ON press_cycles(press_number)",
		"CREATE INDEX IF NOT EXISTS idx_press_cycles_date ON press_cycles(date)",
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
