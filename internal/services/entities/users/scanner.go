package users

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/interfaces"
	"github.com/knackwurstking/pgpress/internal/services/shared/scanner"
	"github.com/knackwurstking/pgpress/pkg/models"
)

func scanUser(scanner interfaces.Scannable) (*models.User, error) {
	user := &models.User{}
	err := scanner.Scan(&user.TelegramID, &user.Name, &user.ApiKey, &user.LastFeed)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to scan user: %v", err)
	}
	return user, nil
}

func scanUsersFromRows(rows *sql.Rows) ([]*models.User, error) {
	return scanner.ScanRows(rows, scanUser)
}
