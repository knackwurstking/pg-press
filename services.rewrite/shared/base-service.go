package shared

import (
	"database/sql"
	"strings"

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
	// NOTE: Manually replace ":table_name" with the actual table name here,
	// currently, there is a syntax issue with sql.Named for table names.
	tableCreationQuery = strings.ReplaceAll(tableCreationQuery, ":table_name", tableName)
	//slog.Debug(fmt.Sprintf("Table creation query: %s\n", tableCreationQuery)) // NOTE: Just for debugging

	_, err := bs.DB.Exec(tableCreationQuery, sql.Named("table_name", tableName))
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}
