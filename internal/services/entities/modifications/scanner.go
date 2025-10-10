package modifications

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/interfaces"
	"github.com/knackwurstking/pgpress/internal/services/shared/scanner"
	"github.com/knackwurstking/pgpress/pkg/models"
)

func modificationScanner(scanner interfaces.Scannable) (*models.Modification[any], error) {
	mod := &models.Modification[any]{}
	err := scanner.Scan(&mod.ID, &mod.UserID, &mod.Data, &mod.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to scan modification: %v", err)
	}
	return mod, nil
}

func scanModificationsFromRows(rows *sql.Rows) ([]*models.Modification[any], error) {
	return scanner.ScanRows(rows, modificationScanner)
}
