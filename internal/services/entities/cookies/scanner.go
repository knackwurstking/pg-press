package cookies

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/interfaces"
	"github.com/knackwurstking/pgpress/internal/services/shared/scanner"
	"github.com/knackwurstking/pgpress/pkg/models"
)

func scanCookie(scanner interfaces.Scannable) (*models.Cookie, error) {
	cookie := &models.Cookie{}
	err := scanner.Scan(&cookie.UserAgent, &cookie.Value, &cookie.ApiKey, &cookie.LastLogin)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to scan cookie: %v", err)
	}
	return cookie, nil
}

func scanCookiesFromRows(rows *sql.Rows) ([]*models.Cookie, error) {
	return scanner.ScanRows(rows, scanCookie)
}
