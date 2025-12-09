package shared

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/knackwurstking/pg-press/errors"
)

type Config struct {
	DriverName       string `json:"driver_name"`
	DatabaseLocation string `json:"database_location"`

	db *sql.DB `json:"-"`
}

func (s *Config) DB() *sql.DB {
	return s.db
}

func (s *Config) Open() *errors.MasterError {
	var err error

	if s.db != nil {
		err = s.Close()
		if err != nil {
			return errors.NewMasterError(err, 0).Wrap("failed to close existing database connection")
		}
	}

	slog.Debug("Opening SQL database connection", "driver", s.DriverName, "location", s.DatabaseLocation)

	// NOTE: Previously used: "?_busy_timeout=30000&_journal_mode=WAL&_foreign_keys=on&_synchronous=NORMAL"
	path := fmt.Sprintf(
		"file:%suserdb.sqlite?cache=shared&mode=rwc&_journal=WAL&_sync=0",
		s.DatabaseLocation,
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
	if s.DB != nil {
		slog.Debug("Closing SQL database connection", "driver", s.DriverName, "location", s.DatabaseLocation)
		return s.db.Close()
	}
	return nil
}
