package main

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/logger"

	_ "github.com/mattn/go-sqlite3"
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

// testDBConnection tests the database connection and configuration
func testDBConnection(customPath string) error {
	path := filepath.Join(configPath, databaseFile)
	logger.Server().Debug("Testing database connection: %s", path)

	if customPath != "" {
		var err error
		if path, err = filepath.Abs(customPath); err != nil {
			return err
		}
	}

	// Test connection with same configuration as openDB
	connectionString := path + "?_busy_timeout=30000&_journal_mode=WAL&_foreign_keys=on&_synchronous=NORMAL"

	db, err := sql.Open("sqlite3", connectionString)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	// Test basic connectivity
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	// Test WAL mode is enabled
	var journalMode string
	err = db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	if err != nil {
		return fmt.Errorf("failed to check journal mode: %v", err)
	}

	if journalMode != "wal" {
		return fmt.Errorf("WAL mode not enabled, current mode: %s", journalMode)
	}

	// Test for potential locking by performing a simple transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin test transaction: %v", err)
	}

	// Simple test query
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&count)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to execute test query: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit test transaction: %v", err)
	}

	logger.Server().Info("Database connection test passed - WAL mode enabled, %d tables found", count)
	return nil
}
