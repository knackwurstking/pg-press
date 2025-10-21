package feeds

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/interfaces"
	"github.com/knackwurstking/pgpress/internal/services/shared/scanner"
	"github.com/knackwurstking/pgpress/pkg/models"
)

func scanFeed(scanner interfaces.Scannable) (*models.Feed, error) {
	feed := &models.Feed{}
	err := scanner.Scan(&feed.ID, &feed.Title, &feed.Content, &feed.UserID, &feed.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to scan feed: %v", err)
	}
	return feed, nil
}

func scanFeedsFromRows(rows *sql.Rows) ([]*models.Feed, error) {
	return scanner.ScanRows(rows, scanFeed)
}
