package shared

import (
	"database/sql"

	"github.com/knackwurstking/pg-press/errors"
)

type BaseService struct {
	*Config
}

func (bs *BaseService) Setup(tableName, tableCreationQuery string) *errors.MasterError {
	merr := bs.Open()
	if merr != nil {
		return merr
	}

	return bs.createSQLTable(tableName, tableCreationQuery)
}

func (s *BaseService) Close() *errors.MasterError {
	err := s.Config.Close()
	if err != nil {
		return errors.NewMasterError(err, 0)
	}
	return nil
}

func (bs *BaseService) createSQLTable(tableName, tableCreationQuery string) *errors.MasterError {
	_, err := bs.DB().Exec(tableCreationQuery, sql.Named("table_name", tableName))
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}
