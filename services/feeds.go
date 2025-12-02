package services

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

const TableNameFeeds = "feeds"

type Feeds struct {
	*Base
	broadcaster Broadcaster
}

func NewFeeds(r *Registry) *Feeds {
	base := NewBase(r)

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %[1]s (
			id INTEGER NOT NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			user_id INTEGER NOT NULL,
			created_at INTEGER NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`, TableNameFeeds)

	if _, err := base.DB.Exec(query); err != nil {
		panic(errors.Wrap(err, "create %s table", TableNameFeeds))
	}

	return &Feeds{
		Base: base,
	}
}

func (f *Feeds) SetBroadcaster(broadcaster Broadcaster) {
	slog.Info("Setting broadcaster for real-time updates")
	f.broadcaster = broadcaster
}

func (f *Feeds) List() ([]*models.Feed, error) {
	slog.Info("Listing feeds")

	query := fmt.Sprintf(
		`SELECT id, title, content, user_id, created_at FROM %s ORDER BY created_at DESC`,
		TableNameFeeds,
	)

	rows, err := f.DB.Query(query)
	if err != nil {
		return nil, f.GetSelectError(err)
	}
	defer rows.Close()

	return ScanRows(rows, scanFeed)
}

func (f *Feeds) ListRange(offset, limit int) ([]*models.Feed, error) {
	slog.Info("Listing feeds with pagination", "offset", offset, "limit", limit)

	query := fmt.Sprintf(
		`SELECT id, title, content, user_id, created_at
		FROM %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`,
		TableNameFeeds,
	)

	rows, err := f.DB.Query(query, limit, offset)
	if err != nil {
		return nil, f.GetSelectError(err)
	}
	defer rows.Close()

	return ScanRows(rows, scanFeed)
}

func (f *Feeds) ListByUser(userID int64, offset, limit int) ([]*models.Feed, error) {
	slog.Info("Listing feeds for user", "telegram_id", userID, "offset", offset, "limit", limit)

	query := fmt.Sprintf(
		`SELECT id, title, content, user_id, created_at
		FROM %s
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`,
		TableNameFeeds,
	)

	rows, err := f.DB.Query(query, userID, limit, offset)
	if err != nil {
		return nil, f.GetSelectError(err)
	}
	defer rows.Close()

	return ScanRows(rows, scanFeed)
}

func (f *Feeds) Get(id models.FeedID) (*models.Feed, error) {
	slog.Info("Getting feed", "id", id)

	query := fmt.Sprintf(
		`SELECT id, title, content, user_id, created_at FROM %s WHERE id = ?`,
		TableNameFeeds,
	)

	row := f.DB.QueryRow(query, id)
	feed, err := ScanSingleRow(row, scanFeed)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError(fmt.Sprintf("feed with id %d not found", id))
		}
		return nil, f.GetSelectError(err)
	}

	return feed, nil
}

func (f *Feeds) Add(feed *models.Feed) error {
	slog.Info("Adding feed", "feed", feed)

	if err := feed.Validate(); err != nil {
		return err
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (title, content, user_id, created_at) VALUES (?, ?, ?, ?)`,
		TableNameFeeds,
	)

	result, err := f.DB.Exec(query, feed.Title, feed.Content, feed.UserID, feed.CreatedAt)
	if err != nil {
		return f.GetInsertError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return f.GetInsertError(err)
	}
	feed.ID = models.FeedID(id)

	if f.broadcaster != nil {
		f.broadcaster.Broadcast()
	}

	return nil
}

// AddSimple creates a new feed with automatic timestamp and broadcasts the update
func (f *Feeds) AddSimple(title, content string, userID models.TelegramID) (*models.Feed, error) {
	slog.Info("Adding simple feed", "title", title, "user_id", userID)

	// Create feed with automatic timestamp
	feed := models.NewFeed(title, content, userID)

	// Validate the feed
	if err := feed.Validate(); err != nil {
		return nil, err
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (title, content, user_id, created_at) VALUES (?, ?, ?, ?)`,
		TableNameFeeds,
	)

	result, err := f.DB.Exec(query, feed.Title, feed.Content, feed.UserID, feed.CreatedAt)
	if err != nil {
		return nil, f.GetInsertError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, f.GetInsertError(err)
	}
	feed.ID = models.FeedID(id)

	// Broadcast update if broadcaster is set
	if f.broadcaster != nil {
		f.broadcaster.Broadcast()
	}

	return feed, nil
}

func (f *Feeds) Delete(id models.FeedID) error {
	slog.Info("Deleting feed", "id", id)

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, TableNameFeeds)
	_, err := f.DB.Exec(query, id)
	if err != nil {
		return f.GetDeleteError(err)
	}

	return nil
}

func (f *Feeds) DeleteBefore(timestamp int64) (int, error) {
	slog.Info("Deleting feeds before timestamp", "timestamp", timestamp)

	query := fmt.Sprintf(`DELETE FROM %s WHERE created_at < ?`, TableNameFeeds)
	result, err := f.DB.Exec(query, timestamp)
	if err != nil {
		return 0, f.GetDeleteError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, f.GetDeleteError(err)
	}

	return int(rowsAffected), nil
}

func (f *Feeds) Count() (int, error) {
	slog.Info("Counting feeds")

	count, err := f.QueryCount(fmt.Sprintf(`SELECT COUNT(*) FROM %s`, TableNameFeeds))
	if err != nil {
		return 0, f.GetSelectError(err)
	}

	return count, nil
}

func (f *Feeds) CountByUser(userID int64) (int, error) {
	slog.Info("Counting feeds by user", "telegram_id", userID)

	count, err := f.QueryCount(
		fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE user_id = ?`, TableNameFeeds),
		userID,
	)
	if err != nil {
		return 0, f.GetSelectError(err)
	}

	return count, nil
}

func scanFeed(scanner Scannable) (*models.Feed, error) {
	feed := &models.Feed{}
	err := scanner.Scan(&feed.ID, &feed.Title, &feed.Content, &feed.UserID, &feed.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("scan feed: %v", err)
	}
	return feed, nil
}
