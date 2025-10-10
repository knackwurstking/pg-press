package feeds

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/interfaces"
	"github.com/knackwurstking/pgpress/internal/services/base"
	"github.com/knackwurstking/pgpress/internal/services/shared/scanner"
	"github.com/knackwurstking/pgpress/internal/services/shared/validation"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

type Service struct {
	*base.BaseService
	broadcaster interfaces.Broadcaster
}

func NewService(db *sql.DB) *Service {
	base := base.NewBaseService(db, "Feeds")

	query := `
		CREATE TABLE IF NOT EXISTS feeds (
			id INTEGER NOT NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			user_id INTEGER NOT NULL,
			created_at INTEGER NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);

		CREATE INDEX IF NOT EXISTS idx_feeds_created_at ON feeds(created_at);
		CREATE INDEX IF NOT EXISTS idx_feeds_user_id ON feeds(user_id);
	`

	if err := base.CreateTable(query, "feeds"); err != nil {
		panic(err)
	}

	return &Service{
		BaseService: base,
	}
}

func (f *Service) SetBroadcaster(broadcaster interfaces.Broadcaster) {
	f.Log.Debug("Setting broadcaster for real-time updates")
	f.broadcaster = broadcaster
}

func (f *Service) List() ([]*models.Feed, error) {
	f.Log.Debug("Listing feeds")

	query := `SELECT id, title, content, user_id, created_at FROM feeds ORDER BY created_at DESC`
	rows, err := f.DB.Query(query)
	if err != nil {
		return nil, f.HandleSelectError(err, "feeds")
	}
	defer rows.Close()

	feeds, err := scanFeedsFromRows(rows)
	if err != nil {
		return nil, err
	}

	return feeds, nil
}

func (f *Service) ListRange(offset, limit int) ([]*models.Feed, error) {
	f.Log.Debug("Listing feeds with pagination: offset: %d, limit: %d", offset, limit)

	query := `SELECT id, title, content, user_id, created_at FROM feeds
		ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := f.DB.Query(query, limit, offset)
	if err != nil {
		return nil, f.HandleSelectError(err, "feeds")
	}
	defer rows.Close()

	feeds, err := scanFeedsFromRows(rows)
	if err != nil {
		return nil, err
	}

	return feeds, nil
}

func (f *Service) ListByUser(userID int64, offset, limit int) ([]*models.Feed, error) {
	if err := validation.ValidateID(userID, "user"); err != nil {
		return nil, err
	}

	query := `SELECT id, title, content, user_id, created_at FROM feeds
		WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := f.DB.Query(query, userID, limit, offset)
	if err != nil {
		return nil, f.HandleSelectError(err, "feeds")
	}
	defer rows.Close()

	feeds, err := scanFeedsFromRows(rows)
	if err != nil {
		return nil, err
	}

	return feeds, nil
}

func (f *Service) Add(feedData *models.Feed) error {
	if err := validateFeed(feedData); err != nil {
		return err
	}

	// Call the model's validate method for additional checks
	if err := feedData.Validate(); err != nil {
		return err
	}

	query := `INSERT INTO feeds (title, content, user_id, created_at) VALUES (?, ?, ?, ?)`
	result, err := f.DB.Exec(query, feedData.Title, feedData.Content, feedData.UserID, feedData.CreatedAt)
	if err != nil {
		return f.HandleInsertError(err, "feeds")
	}

	// Get the generated ID
	id, err := result.LastInsertId()
	if err != nil {
		return f.HandleInsertError(err, "feeds")
	}
	feedData.ID = id

	// Notify about new feed if broadcaster is set
	if f.broadcaster != nil {
		f.broadcaster.Broadcast()
	}

	f.Log.Debug("Added feed: id: %d", id)
	return nil
}

func (f *Service) Count() (int, error) {
	f.Log.Debug("Counting feeds")

	var count int
	query := `SELECT COUNT(*) FROM feeds`
	err := f.DB.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, f.HandleSelectError(err, "feeds")
	}

	return count, nil
}

func (f *Service) CountByUser(userID int64) (int, error) {
	if err := validation.ValidateID(userID, "user"); err != nil {
		return 0, err
	}

	f.Log.Debug("Counting feeds by user: %d", userID)

	var count int
	query := `SELECT COUNT(*) FROM feeds WHERE user_id = ?`
	err := f.DB.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, f.HandleSelectError(err, "feeds")
	}

	return count, nil
}

func (f *Service) DeleteBefore(timestamp int64) (int64, error) {
	if err := validation.ValidateTimestamp(timestamp, "timestamp"); err != nil {
		return 0, err
	}

	f.Log.Info("Deleting feeds before timestamp %d", timestamp)

	query := `DELETE FROM feeds WHERE created_at < ?`
	result, err := f.DB.Exec(query, timestamp)
	if err != nil {
		return 0, f.HandleDeleteError(err, "feeds")
	}

	rowsAffected, err := f.GetRowsAffected(result, "delete feeds before")
	if err != nil {
		return 0, err
	}

	f.Log.Info("Deleted %d feeds before timestamp %d", rowsAffected, timestamp)

	return rowsAffected, nil
}

func (f *Service) Get(id int64) (*models.Feed, error) {
	if err := validation.ValidateID(id, "feed"); err != nil {
		return nil, err
	}

	f.Log.Debug("Getting feed: %d", id)

	query := `SELECT id, title, content, user_id, created_at FROM feeds WHERE id = ?`
	row := f.DB.QueryRow(query, id)

	feed, err := scanner.ScanSingleRow(row, scanFeed, "feeds")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("feed with id %d not found", id))
		}
		return nil, err
	}

	return feed, nil
}

func (f *Service) Delete(id int64) error {
	if err := validation.ValidateID(id, "feed"); err != nil {
		return err
	}

	f.Log.Debug("Deleting feed: %d", id)

	query := `DELETE FROM feeds WHERE id = ?`
	result, err := f.DB.Exec(query, id)
	if err != nil {
		return f.HandleDeleteError(err, "feeds")
	}

	if err := f.CheckRowsAffected(result, "feed", id); err != nil {
		return err
	}

	f.Log.Debug("Deleted feed: %d", id)
	return nil
}
