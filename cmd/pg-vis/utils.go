package main

import (
	"database/sql"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"github.com/knackwurstking/pg-vis/internal/database"
)

func openDB(customPath string) (*database.DB, error) {
	path := filepath.Join(configPath, databaseFile)

	if customPath != "" {
		var err error
		path, err = filepath.Abs(customPath)
		if err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	return database.New(db), nil
}
