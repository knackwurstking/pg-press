package main

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"github.com/knackwurstking/pg-vis/pkg/pgvis"
)

func getDB(customDBPath *string) (*pgvis.DB, error) {
	dbPath := filepath.Join(configPath, databaseFile)

	if customDBPath != nil {
		var err error
		dbPath, err = filepath.Abs(*customDBPath)
		if err != nil {
			return nil, err
		}
	}

	return openDB(dbPath)
}

func openDB(path string) (*pgvis.DB, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	return pgvis.NewDB(db), nil
}
