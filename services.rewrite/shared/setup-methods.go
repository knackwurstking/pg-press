package shared

import (
	"database/sql"

	"github.com/knackwurstking/pg-press/errors"
)

func (s *Setup) Open() *errors.MasterError {
	var err error

	if s.DB != nil {
		err = s.Close()
		if err != nil {
			return errors.NewMasterError(err, 0).Wrap("failed to close existing database connection")
		}
	}

	if s.EnableSQL {
		s.DB, err = sql.Open(s.DriverName, s.DataSourceName)
		if err != nil {
			return errors.NewMasterError(err, 0)
		}
	}

	return nil
}

func (s *Setup) Close() error {
	if s.DB != nil {
		return s.DB.Close()
	}
	return nil
}
