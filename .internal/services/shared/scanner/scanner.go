package scanner

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/interfaces"
)



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
