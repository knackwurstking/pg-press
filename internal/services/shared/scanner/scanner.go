package scanner

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/interfaces"
)

// ScanSingleRow is a helper for scanning a single row with error handling
func ScanSingleRow[T any](row *sql.Row, scanFunc func(interfaces.Scannable) (*T, error), entityName string) (*T, error) {
	result, err := scanFunc(row)
	if err != nil {
		return nil, handleScanError(err, entityName)
	}
	return result, nil
}

// ScanRows provides a generic way to scan multiple rows
func ScanRows[T any](rows *sql.Rows, scanFunc func(interfaces.Scannable) (*T, error)) ([]*T, error) {
	var results []*T

	for rows.Next() {
		item, err := scanFunc(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %v", err)
	}

	return results, nil
}

// ScanIntoMap scans rows into a map by ID for efficient lookup
func ScanIntoMap[T any, K comparable](rows *sql.Rows, scanFunc func(interfaces.Scannable) (*T, error), keyFunc func(*T) K) (map[K]*T, error) {
	resultMap := make(map[K]*T)

	for rows.Next() {
		item, err := scanFunc(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		key := keyFunc(item)
		resultMap[key] = item
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %v", err)
	}

	return resultMap, nil
}

func handleScanError(err error, entityName string) error {
	if err == sql.ErrNoRows {
		return err
	}
	return fmt.Errorf("scan error: %s: %v", entityName, err)
}
