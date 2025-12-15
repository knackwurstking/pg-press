package shared

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/logger"
	"github.com/knackwurstking/ui/ui-templ"
)

type Config struct {
	DriverName       string `json:"driver_name"`
	DatabaseLocation string `json:"database_location"`
}

type BaseService struct {
	*Config `json:"config"`

	Log *ui.Logger `json:"-"`

	serviceName string  `json:"-"`
	db          *sql.DB `json:"-"`
}

func NewBaseService(c *Config, serviceName string) *BaseService {
	return &BaseService{
		Config:      c,
		Log:         logger.New("service: " + serviceName),
		serviceName: serviceName,
	}
}

func (s *BaseService) DB() *sql.DB {
	return s.db
}

func (bs *BaseService) Setup(dbName, tableCreationQuery string) *errors.MasterError {
	if bs.Log != nil {
		bs.Log.Debug("Service setup: [name: %s, path: %s]", dbName, bs.DatabaseLocation)
	}

	if bs.db != nil {
		return nil
	}

	var err error
	err = os.MkdirAll(bs.DatabaseLocation, 0700)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	path := fmt.Sprintf(
		"file:%s.sqlite?cache=shared&mode=rwc&_journal=WAL&_sync=0",
		filepath.Join(bs.DatabaseLocation, dbName),
	)
	bs.db, err = sql.Open(bs.DriverName, path)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	// Configure connection pool to prevent resource exhaustion
	bs.db.SetMaxOpenConns(10)                 // Allow more concurrent connections
	bs.db.SetMaxIdleConns(5)                  // Keep some connections alive
	bs.db.SetConnMaxLifetime(5 * time.Minute) // Close connections after 5 minutes

	return bs.createSQLTable(tableCreationQuery)

}

func (bs *BaseService) Close() *errors.MasterError {
	if bs.db != nil {
		err := bs.db.Close()
		if err != nil {
			return errors.NewMasterError(err, 0)
		}
		bs.db = nil
	}
	return nil
}

func (bs *BaseService) createSQLTable(tableCreationQuery string) *errors.MasterError {
	_, err := bs.DB().Exec(tableCreationQuery)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}
