package shared

import (
	"database/sql"
	"log/slog"

	"github.com/knackwurstking/pg-press/errors"
)

type Config struct {
	// Contains configuration parameters for the service(s)
	EnableSQL      bool   `json:"enable_sql"`
	DriverName     string `json:"driver_name"`
	DataSourceName string `json:"data_source_name"`

	// DB is the database connection instance if SQL is enabled
	DB *sql.DB `json:"-"`
}

func (s *Config) Open() *errors.MasterError {
	var err error

	if s.DB != nil {
		err = s.Close()
		if err != nil {
			return errors.NewMasterError(err, 0).Wrap("failed to close existing database connection")
		}
	}

	if s.EnableSQL {
		slog.Debug("Opening SQL database connection", "driver", s.DriverName, "data_source", s.DataSourceName)
		s.DB, err = sql.Open(s.DriverName, s.DataSourceName)
		if err != nil {
			return errors.NewMasterError(err, 0)
		}
	}

	return nil
}

func (s *Config) Close() error {
	if s.DB != nil {
		slog.Debug("Closing SQL database connection", "driver", s.DriverName, "data_source", s.DataSourceName)
		return s.DB.Close()
	}
	return nil
}
