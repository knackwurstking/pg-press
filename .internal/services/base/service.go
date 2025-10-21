package base

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/pkg/utils"
)

// CreateTable executes a table creation query with error handling


// HandleScanError provides consistent error handling for row scanning operations
func (b *BaseService) HandleScanError(err error, entityName string) error {
	if err == sql.ErrNoRows {
		return err
	}
	return fmt.Errorf("scan error: %s: %v", entityName, err)
}

// HandleSelectError provides consistent error handling for select operations
func (b *BaseService) HandleSelectError(err error, entityName string) error {
	return fmt.Errorf("select error: %s: %v", entityName, err)
}

// HandleInsertError provides consistent error handling for insert operations
func (b *BaseService) HandleInsertError(err error, entityName string) error {
	return fmt.Errorf("insert error: %s: %v", entityName, err)
}

// HandleUpdateError provides consistent error handling for update operations
func (b *BaseService) HandleUpdateError(err error, entityName string) error {
	return fmt.Errorf("update error: %s: %v", entityName, err)
}

// HandleDeleteError provides consistent error handling for delete operations
func (b *BaseService) HandleDeleteError(err error, entityName string) error {
	return fmt.Errorf("delete error: %s: %v", entityName, err)
}

// checkExistence checks if a record exists with the given query and parameters
func (b *BaseService) CheckExistence(query string, args ...any) (bool, error) {
	var count int
	err := b.DB.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetRowsAffected safely gets the number of rows affected from a result
func (b *BaseService) GetRowsAffected(result sql.Result, operation string) (int64, error) {
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected for %s: %v", operation, err)
	}
	return rowsAffected, nil
}

// CheckRowsAffected checks if any rows were affected and returns not found error if none
func (b *BaseService) CheckRowsAffected(result sql.Result, entityName string, id interface{}) error {
	rowsAffected, err := b.GetRowsAffected(result, entityName)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return utils.NewNotFoundError(fmt.Sprintf("%s with ID %v not found", entityName, id))
	}
	return nil
}

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
