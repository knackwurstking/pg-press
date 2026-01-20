// Package db provides database access and management functionality for the PG Press application.
//
// It handles connections to multiple SQLite databases (tools, presses, notes, users, reports)
// and manages the creation and maintenance of database tables and records.
//
// Database Overview:
// - tools: Stores information about tools, including their physical properties and cycles
// - presses: Manages press machines and their configuration in slots
// - notes: Stores user notes associated with various entities
// - users: Manages user accounts and authentication
// - reports: Handles trouble reports and related data
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/logger"
)

// Scannable interface defines the required scanning method for database rows.
type Scannable interface {
	Scan(dest ...any) error
}

// Database connection variables for different modules.
var (
	dbTool    *sql.DB
	dbPress   *sql.DB
	dbNote    *sql.DB
	dbUser    *sql.DB
	dbReports *sql.DB

	log = logger.New("db")
)

// Open initializes and opens all database connections.
//
// It creates the necessary directories for database files, configures SQLite
// connection parameters for optimal performance, and initializes all required tables.
//
// Parameters:
//   - path: The directory path where database files should be stored
//   - allowCreate: Whether to create new database files if they don't exist
//
// Returns:
//   - error: An error if any database connection or initialization fails
func Open(path string, allowCreate bool) error {
	if err := os.MkdirAll(path, 0700); err != nil {
		return fmt.Errorf("failed to create database directory: %v", err)
	}

	mode := "rw"
	if allowCreate {
		mode = "rwc"
	}

	wg := &sync.WaitGroup{}
	names := []string{"tool", "press", "note", "user", "reports"}
	chErr := make(chan error, len(names))
	for _, name := range names {
		log.Debug("Opening %s database at %s", name, path)

		wg.Go(func() {
			path := fmt.Sprintf(
				"file:%s.sqlite?cache=shared&mode=%s&journal=WAL&synchronous=1",
				filepath.Join(path, name), mode,
			)
			db, err := sql.Open("sqlite3", path)
			if err != nil {
				chErr <- fmt.Errorf("failed to open %s database: %v", name, err)
				return
			}

			// Configure connection pool to prevent resource exhaustion
			db.SetMaxOpenConns(10)                 // Allow more concurrent connections
			db.SetMaxIdleConns(5)                  // Keep some connections alive
			db.SetConnMaxLifetime(5 * time.Minute) // Close connections after 5 minutes

			switch name {
			case "tool":
				dbTool = db
				if err = createTable(db, sqlCreateMetalSheetsTable); err != nil {
					chErr <- errors.Wrap(err, "failed to create metal_sheets table")
					return
				}
				if err = createTable(db, sqlCreateToolRegenerationsTable); err != nil {
					chErr <- errors.Wrap(err, "failed to create tool_regenerations table")
					return
				}
				if err = createTable(db, sqlCreateToolsTable); err != nil {
					chErr <- errors.Wrap(err, "failed to create tools table")
					return
				}

			case "press":
				dbPress = db
				if err = createTable(db, sqlCreateCyclesTable); err != nil {
					chErr <- errors.Wrap(err, "failed to create cycles table")
					return
				}
				if err = createTable(db, sqlCreatePressesTable); err != nil {
					chErr <- errors.Wrap(err, "failed to create presses table")
					return
				}
				if err = createTable(db, sqlCreatePressRegenerationsTable); err != nil {
					chErr <- errors.Wrap(err, "failed to create press_regenerations table")
					return
				}

			case "note":
				dbNote = db
				if err := createTable(db, sqlCreateNotesTable); err != nil {
					chErr <- errors.Wrap(err, "failed to create notes table")
					return
				}

			case "user":
				dbUser = db
				if err := createTable(db, sqlCreateCookiesTable); err != nil {
					chErr <- errors.Wrap(err, "failed to create cookies table")
					return
				}
				if err := createTable(db, sqlCreateUsersTable); err != nil {
					chErr <- errors.Wrap(err, "failed to create users table")
					return
				}

			case "reports":
				dbReports = db
				if err := createTable(db, sqlCreateTroubleReportsTable); err != nil {
					chErr <- errors.Wrap(err, "failed to create trouble_reports table")
					return
				}
			}

			chErr <- nil
		})
	}
	wg.Wait()
	close(chErr)

	var errs []string
	for e := range chErr {
		if e != nil {
			errs = append(errs, e.Error())
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}
	return nil
}

// Close closes all open database connections.
func Close() {
	for _, db := range []*sql.DB{dbTool, dbPress, dbNote, dbUser} {
		if db != nil {
			db.Close()
		}
	}
}

// createTable executes a SQL statement to create a database table.
//
// Parameters:
//   - db: The database connection to use
//   - query: The SQL CREATE TABLE statement to execute
//
// Returns:
//   - error: An error if the table creation fails
func createTable(db *sql.DB, query string) error {
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}
	return nil
}
