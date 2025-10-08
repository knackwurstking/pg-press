package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/knackwurstking/pgpress/internal/interfaces"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

// Feeds handles database operations for feed entries
type Feeds struct {
	*BaseService
	broadcaster interfaces.Broadcaster
}

// NewFeeds creates a new Feed instance and initializes the database table
func NewFeeds(db *sql.DB) *Feeds {
	base := NewBaseService(db, "Feeds")

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

	return &Feeds{
		BaseService: base,
	}
}

// SetBroadcaster sets the feed notifier for real-time updates
func (f *Feeds) SetBroadcaster(broadcaster interfaces.Broadcaster) {
	f.LogOperation("Setting broadcaster for real-time updates")
	f.broadcaster = broadcaster
}

// List retrieves all feeds ordered by creation time in descending order
func (f *Feeds) List() ([]*models.Feed, error) {
	start := time.Now()
	f.LogOperation("Listing feeds")

	query := `SELECT id, title, content, user_id, created_at FROM feeds ORDER BY created_at DESC`
	rows, err := f.db.Query(query)
	if err != nil {
		return nil, f.HandleSelectError(err, "feeds")
	}
	defer rows.Close()

	feeds, err := ScanFeedsFromRows(rows)
	if err != nil {
		return nil, err
	}

	f.LogSlowQuery(start, "feed list", 100*time.Millisecond, fmt.Sprintf("%d feeds", len(feeds)))
	return feeds, nil
}

// ListRange retrieves a specific range of feeds with pagination support
func (f *Feeds) ListRange(offset, limit int) ([]*models.Feed, error) {
	if err := f.validatePagination(offset, limit); err != nil {
		return nil, err
	}

	start := time.Now()
	f.LogOperation("Listing feeds with pagination", fmt.Sprintf("offset: %d, limit: %d", offset, limit))

	query := `SELECT id, title, content, user_id, created_at FROM feeds
		ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := f.db.Query(query, limit, offset)
	if err != nil {
		return nil, f.HandleSelectError(err, "feeds")
	}
	defer rows.Close()

	feeds, err := ScanFeedsFromRows(rows)
	if err != nil {
		return nil, err
	}

	f.LogSlowQuery(start, "feed range", 100*time.Millisecond,
		fmt.Sprintf("offset=%d, limit=%d, returned=%d", offset, limit, len(feeds)))
	return feeds, nil
}

// ListByUser retrieves feeds created by a specific user
func (f *Feeds) ListByUser(userID int64, offset, limit int) ([]*models.Feed, error) {
	if err := ValidateID(userID, "user"); err != nil {
		return nil, err
	}

	if err := f.validatePagination(offset, limit); err != nil {
		return nil, err
	}

	start := time.Now()
	f.LogOperation("Listing feeds by user", fmt.Sprintf("userID: %d, offset: %d, limit: %d", userID, offset, limit))

	query := `SELECT id, title, content, user_id, created_at FROM feeds
		WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := f.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, f.HandleSelectError(err, "feeds")
	}
	defer rows.Close()

	feeds, err := ScanFeedsFromRows(rows)
	if err != nil {
		return nil, err
	}

	f.LogSlowQuery(start, "user feeds", 100*time.Millisecond,
		fmt.Sprintf("userID=%d, offset=%d, limit=%d, returned=%d", userID, offset, limit, len(feeds)))
	return feeds, nil
}

// Add creates a new feed entry in the database
func (f *Feeds) Add(feedData *models.Feed) error {
	if err := ValidateFeed(feedData); err != nil {
		return err
	}

	// Call the model's validate method for additional checks
	if err := feedData.Validate(); err != nil {
		return err
	}

	start := time.Now()
	f.LogOperation("Adding feed", fmt.Sprintf("user: %d, title: %s", feedData.UserID, feedData.Title))

	query := `INSERT INTO feeds (title, content, user_id, created_at) VALUES (?, ?, ?, ?)`
	result, err := f.db.Exec(query, feedData.Title, feedData.Content, feedData.UserID, feedData.CreatedAt)
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

	f.LogSlowQuery(start, "feed insert", 50*time.Millisecond, fmt.Sprintf("user: %d", feedData.UserID))
	f.LogOperation("Added feed", fmt.Sprintf("id: %d", id))
	return nil
}

// Count returns the total number of feeds in the database
func (f *Feeds) Count() (int, error) {
	start := time.Now()
	f.LogOperation("Counting feeds")

	var count int
	query := `SELECT COUNT(*) FROM feeds`
	err := f.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, f.HandleSelectError(err, "feeds")
	}

	f.LogSlowQuery(start, "feed count", 50*time.Millisecond, fmt.Sprintf("result: %d", count))
	return count, nil
}

// CountByUser returns the number of feeds created by a specific user
func (f *Feeds) CountByUser(userID int64) (int, error) {
	if err := ValidateID(userID, "user"); err != nil {
		return 0, err
	}

	start := time.Now()
	f.LogOperation("Counting feeds by user", userID)

	var count int
	query := `SELECT COUNT(*) FROM feeds WHERE user_id = ?`
	err := f.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, f.HandleSelectError(err, "feeds")
	}

	f.LogSlowQuery(start, "user feed count", 50*time.Millisecond,
		fmt.Sprintf("userID=%d, result=%d", userID, count))
	return count, nil
}

// DeleteBefore removes all feeds created before the specified timestamp
func (f *Feeds) DeleteBefore(timestamp int64) (int64, error) {
	if err := ValidateTimestamp(timestamp, "timestamp"); err != nil {
		return 0, err
	}

	start := time.Now()
	f.log.Info("Deleting feeds before timestamp %d", timestamp)

	query := `DELETE FROM feeds WHERE created_at < ?`
	result, err := f.db.Exec(query, timestamp)
	if err != nil {
		return 0, f.HandleDeleteError(err, "feeds")
	}

	rowsAffected, err := f.GetRowsAffected(result, "delete feeds before")
	if err != nil {
		return 0, err
	}

	elapsed := time.Since(start)
	f.log.Info("Deleted %d feeds before timestamp %d in %v", rowsAffected, timestamp, elapsed)
	f.LogSlowQuery(start, "feed deletion", 100*time.Millisecond,
		fmt.Sprintf("timestamp=%d, deleted=%d", timestamp, rowsAffected))

	return rowsAffected, nil
}

// Get retrieves a specific feed by ID
func (f *Feeds) Get(id int64) (*models.Feed, error) {
	if err := ValidateID(id, "feed"); err != nil {
		return nil, err
	}

	f.LogOperation("Getting feed", id)

	query := `SELECT id, title, content, user_id, created_at FROM feeds WHERE id = ?`
	row := f.db.QueryRow(query, id)

	feed, err := ScanSingleRow(row, ScanFeed, "feeds")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("feed with id %d not found", id))
		}
		return nil, err
	}

	return feed, nil
}

// Delete removes a specific feed by ID
func (f *Feeds) Delete(id int64) error {
	if err := ValidateID(id, "feed"); err != nil {
		return err
	}

	f.LogOperation("Deleting feed", id)

	query := `DELETE FROM feeds WHERE id = ?`
	result, err := f.db.Exec(query, id)
	if err != nil {
		return f.HandleDeleteError(err, "feeds")
	}

	if err := f.CheckRowsAffected(result, "feed", id); err != nil {
		return err
	}

	f.LogOperation("Deleted feed", id)
	return nil
}

// validatePagination validates pagination parameters
func (f *Feeds) validatePagination(offset, limit int) error {
	if offset < 0 {
		return utils.NewValidationError("offset: must be non-negative")
	}
	if limit <= 0 {
		return utils.NewValidationError("limit: must be positive")
	}
	if limit > 1000 {
		return utils.NewValidationError("limit: must not exceed 1000")
	}
	return nil
}
