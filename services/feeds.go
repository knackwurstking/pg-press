package services

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/errors"
	"github.com/knackwurstking/pgpress/logger"
	"github.com/knackwurstking/pgpress/models"
)

const TableNameFeeds = "feeds"

type Feeds struct {
	*Base
	broadcaster Broadcaster
}

func NewFeeds(r *Registry) *Feeds {
	base := NewBase(r, logger.NewComponentLogger("Service: Feeds"))

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %[1]s (
			id INTEGER NOT NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			user_id INTEGER NOT NULL,
			created_at INTEGER NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);

		CREATE INDEX IF NOT EXISTS idx_%[1]s_created_at ON %[1]s(created_at);
		CREATE INDEX IF NOT EXISTS idx_%[1]s_user_id ON %[1]s(user_id);
	`, TableNameFeeds)

	if err := base.CreateTable(query, "feeds"); err != nil {
		panic(err)
	}

	return &Feeds{
		Base: base,
	}
}

func (f *Feeds) SetBroadcaster(broadcaster Broadcaster) {
	f.Log.Debug("Setting broadcaster for real-time updates")
	f.broadcaster = broadcaster
}

func (f *Feeds) List() ([]*models.Feed, error) {
	f.Log.Debug("Listing feeds")

	query := `SELECT id, title, content, user_id, created_at FROM feeds ORDER BY created_at DESC`
	rows, err := f.DB.Query(query)
	if err != nil {
		return nil, f.GetSelectError(err)
	}
	defer rows.Close()

	feeds, err := ScanRows(rows, scanFeed)
	if err != nil {
		return nil, err
	}

	return feeds, nil
}

func (f *Feeds) ListRange(offset, limit int) ([]*models.Feed, error) {
	f.Log.Debug("Listing feeds with pagination: offset: %d, limit: %d", offset, limit)

	query := `
		SELECT
			id, title, content, user_id, created_at
		FROM
			feeds
		ORDER BY
			created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := f.DB.Query(query, limit, offset)
	if err != nil {
		return nil, f.GetSelectError(err)
	}
	defer rows.Close()

	feeds, err := ScanRows(rows, scanFeed)
	if err != nil {
		return nil, err
	}

	return feeds, nil
}

func (f *Feeds) ListByUser(userID int64, offset, limit int) ([]*models.Feed, error) {
	f.Log.Debug("Listing feeds by user: userID: %d, offset: %d, limit: %d",
		userID, offset, limit)

	query := `
		SELECT
			id, title, content, user_id, created_at
		FROM
			feeds
		WHERE
			user_id = ?
		ORDER BY
			created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := f.DB.Query(query, userID, limit, offset)
	if err != nil {
		return nil, f.GetSelectError(err)
	}
	defer rows.Close()

	feeds, err := ScanRows(rows, scanFeed)
	if err != nil {
		return nil, err
	}

	return feeds, nil
}

func (f *Feeds) Add(feed *models.Feed) error {
	f.Log.Debug("Adding feed: %#v", feed)

	if err := feed.Validate(); err != nil {
		return err
	}

	query := `INSERT INTO feeds (title, content, user_id, created_at) VALUES (?, ?, ?, ?)`
	result, err := f.DB.Exec(query, feed.Title, feed.Content, feed.UserID, feed.CreatedAt)
	if err != nil {
		return f.GetInsertError(err)
	}

	// Get the generated ID
	id, err := result.LastInsertId()
	if err != nil {
		return f.GetInsertError(err)
	}
	feed.ID = id

	// Notify about new feed if broadcaster is set
	if f.broadcaster != nil {
		f.broadcaster.Broadcast()
	}

	return nil
}

func (f *Feeds) Count() (int, error) {
	f.Log.Debug("Counting feeds")

	count, err := f.QueryCount(`SELECT COUNT(*) FROM feeds`)
	if err != nil {
		return 0, f.GetSelectError(err)
	}

	return count, nil
}

func (f *Feeds) CountByUser(userID int64) (int, error) {
	f.Log.Debug("Counting feeds by user: %d", userID)

	count, err := f.QueryCount(`SELECT COUNT(*) FROM feeds WHERE user_id = ?`, userID)
	if err != nil {
		return 0, f.GetSelectError(err)
	}

	return count, nil
}

func (f *Feeds) DeleteBefore(timestamp int64) error {
	f.Log.Debug("Deleting feeds before timestamp %d", timestamp)

	query := `DELETE FROM feeds WHERE created_at < ?`
	_, err := f.DB.Exec(query, timestamp)
	if err != nil {
		return f.GetDeleteError(err)
	}

	return nil
}

func (f *Feeds) Get(id int64) (*models.Feed, error) {
	f.Log.Debug("Getting feed: %d", id)

	query := `SELECT id, title, content, user_id, created_at FROM feeds WHERE id = ?`
	row := f.DB.QueryRow(query, id)

	feed, err := ScanSingleRow(row, scanFeed)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError(fmt.Sprintf(
				"feed with id %d not found", id))
		}

		return nil, f.GetSelectError(err)
	}

	return feed, nil
}

func (f *Feeds) Delete(id int64) error {
	f.Log.Debug("Deleting feed: %d", id)

	query := `DELETE FROM feeds WHERE id = ?`
	_, err := f.DB.Exec(query, id)
	if err != nil {
		return f.GetDeleteError(err)
	}

	return nil
}

func scanFeed(scanner Scannable) (*models.Feed, error) {
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
