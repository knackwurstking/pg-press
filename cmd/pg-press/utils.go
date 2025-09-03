package main

import (
	"database/sql"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"github.com/knackwurstking/pgpress/internal/database/core"
	"github.com/knackwurstking/pgpress/internal/logger"
)

func openDB(customPath string) (*database.DB, error) {
	path := filepath.Join(configPath, databaseFile)
	logger.Server().Debug("Database path: %s", path)

	if customPath != "" {
		var err error
		if path, err = filepath.Abs(customPath); err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	return database.New(db), nil
}
