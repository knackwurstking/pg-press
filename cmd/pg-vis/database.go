package main

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"github.com/knackwurstking/pg-vis/pkg/pgvis"
)

func openDB(customPath *string) (*pgvis.DB, error) {
	path := filepath.Join(configPath, databaseFile)

	if customPath != nil {
		var err error
		path, err = filepath.Abs(*customPath)
		if err != nil {
			return nil, err
		}
	}

	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	return pgvis.NewDB(db), nil
}
