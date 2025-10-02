package main

import (
	"database/sql"
	"path/filepath"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/pkg/logger"

	_ "github.com/mattn/go-sqlite3"
)

func log() *logger.Logger {
	return logger.GetComponentLogger("Server")
}

func clog(component string) *logger.Logger {
	return logger.GetComponentLogger(component)
}

func openDB(customPath string) (*database.DB, error) {
	path := filepath.Join(configPath, databaseFile)
	log().Debug("Database path: %s", path)

	if customPath != "" {
		var err error
		if path, err = filepath.Abs(customPath); err != nil {
			return nil, err
		}
	}

	// Configure SQLite connection string with parameters to prevent locking issues
	connectionString := path + "?_busy_timeout=30000&_journal_mode=WAL&_foreign_keys=on&_synchronous=NORMAL"

	db, err := sql.Open("sqlite3", connectionString)
	if err != nil {
		return nil, err
	}

	// Configure connection pool to prevent resource exhaustion
	db.SetMaxOpenConns(1)    // SQLite works best with single writer
	db.SetMaxIdleConns(1)    // Keep one connection alive
	db.SetConnMaxLifetime(0) // No maximum lifetime

	// Test the connection
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return database.New(db), nil
}
