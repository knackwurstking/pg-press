package shared

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/errors"
)

type Config struct {
	DriverName       string `json:"driver_name"`
	DatabaseLocation string `json:"database_location"`

	db *sql.DB `json:"-"`
}

func (s *Config) DB() *sql.DB {
	return s.db
}

func (s *Config) Open(dbName string) *errors.MasterError {
	var err error

	if s.db != nil {
		_ = s.Close()
	}

	if env.Verbose {
		log.Println("Opening database:", s.DriverName, "at", filepath.Join(s.DatabaseLocation, dbName))
	}

	// NOTE: Previously used: "?_busy_timeout=30000&_journal_mode=WAL&_foreign_keys=on&_synchronous=NORMAL"
	path := fmt.Sprintf(
		"file:%s.sqlite?cache=shared&mode=rwc&_journal=WAL&_sync=0",
		filepath.Join(s.DatabaseLocation, dbName),
	)

	s.db, err = sql.Open(s.DriverName, path)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	// Configure connection pool to prevent resource exhaustion
	s.db.SetMaxOpenConns(10)                 // Allow more concurrent connections
	s.db.SetMaxIdleConns(5)                  // Keep some connections alive
	s.db.SetConnMaxLifetime(5 * time.Minute) // Close connections after 5 minutes

	return nil
}

func (s *Config) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
