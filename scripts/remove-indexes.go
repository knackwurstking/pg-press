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
		log.Fatal("Usage: go run main.go <database-path>")
	}

	dbPath := os.Args[1]

	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Check if database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Fatal("Database file does not exist:", dbPath)
	}

	// Get list of all indexes in the database
	indexes, err := getIndexes(db)
	if err != nil {
		log.Fatal("Failed to get indexes:", err)
	}

	fmt.Printf("Found %d indexes to remove\n", len(indexes))

	if len(indexes) == 0 {
		fmt.Println("No indexes found in database")
		return
	}

	// Drop each index
	for _, index := range indexes {
		fmt.Printf("Dropping index: %s\n", index.Name)
		_, err := db.Exec(fmt.Sprintf("DROP INDEX IF EXISTS %s", index.Name))
		if err != nil {
			log.Printf("Failed to drop index %s: %v", index.Name, err)
		} else {
			fmt.Printf("Successfully dropped index: %s\n", index.Name)
		}
	}

	fmt.Println("All indexes have been removed from the database")
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
