package services

import (
	"database/sql"
	"fmt"
)

type Scannable interface {
	Scan(dest ...any) error
}

func ScanRows[T any](rows *sql.Rows, scanFunc func(Scannable) (*T, error)) ([]*T, error) {
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

func ScanSingleRow[T any](row *sql.Row, scanFunc func(Scannable) (*T, error)) (*T, error) {
	result, err := scanFunc(row)
	if err != nil {
		return nil, fmt.Errorf("scan failed: %v", err)
	}

	return result, nil
}
