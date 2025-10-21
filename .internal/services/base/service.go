package base

import (
	"database/sql"
	"fmt"
)

// CreateTable executes a table creation query with error handling

//// GetRowsAffected safely gets the number of rows affected from a result
//func (b *BaseService) GetRowsAffected(result sql.Result, operation string) (int64, error) {
//	rowsAffected, err := result.RowsAffected()
//	if err != nil {
//		return 0, fmt.Errorf("failed to get rows affected for %s: %v", operation, err)
//	}
//	return rowsAffected, nil
//}
//
//// CheckRowsAffected checks if any rows were affected and returns not found error if none
//func (b *BaseService) CheckRowsAffected(result sql.Result, entityName string, id interface{}) error {
//	rowsAffected, err := b.GetRowsAffected(result, entityName)
//	if err != nil {
//		return err
//	}
//	if rowsAffected == 0 {
//		return utils.NewNotFoundError(fmt.Sprintf("%s with ID %v not found", entityName, id))
//	}
//	return nil
//}

// ExecuteInTransaction executes a function within a database transaction
func (b *BaseService) ExecuteInTransaction(fn func(tx *sql.Tx) error) error {
	tx, err := b.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}
