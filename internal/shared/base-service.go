package shared

import (
	"database/sql"

	"github.com/knackwurstking/pg-press/internal/errors"
)

type BaseService struct {
	*Config
}

func (bs *BaseService) Setup(dbName, tableCreationQuery string) *errors.MasterError {
	merr := bs.Open(dbName)
	if merr != nil {
		return merr
	}

	return bs.createSQLTable(tableCreationQuery)
}

func (s *BaseService) Close() *errors.MasterError {
	err := s.Config.Close()
	if err != nil {
		return errors.NewMasterError(err, 0)
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
