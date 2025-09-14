package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/knackwurstking/pgpress/internal/interfaces"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

// Feed handles database operations for feed entries
type Feed struct {
	db          *sql.DB
	broadcaster interfaces.Broadcaster
}

// NewFeed creates a new Feed instance and initializes the database table
func NewFeed(db *sql.DB) *Feed {
	//dropQuery := `DROP TABLE IF EXISTS feeds;`
	//if _, err := db.Exec(dropQuery); err != nil {
	//	panic(fmt.Errorf("failed to drop feeds table: %v", err))
	//}

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
	if _, err := db.Exec(query); err != nil {
		panic(fmt.Errorf("failed to create feeds table: %v", err))
	}
	return &Feed{db: db}
}

// SetBroadcaster sets the feed notifier for real-time updates
func (f *Feed) SetBroadcaster(broadcaster interfaces.Broadcaster) {
	logger.DBFeeds().Debug("Setting broadcaster for real-time updates")
	f.broadcaster = broadcaster
}

// List retrieves all feeds ordered by creation time in descending order
func (f *Feed) List() ([]*models.Feed, error) {
	start := time.Now()

	query := `SELECT id, title, content, user_id, created_at FROM feeds ORDER BY created_at DESC`
	rows, err := f.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("select error: feeds: %v", err)
	}
	defer rows.Close()

	feeds, err := f.scanAllRows(rows)
	elapsed := time.Since(start)

	if err != nil {
		return nil, fmt.Errorf("scan error: feeds: %v", err)
	}

	if elapsed > 100*time.Millisecond {
		logger.DBFeeds().Warn("Slow feed list query took %v for %d feeds", elapsed, len(feeds))
	}

	return feeds, nil
}

// ListRange retrieves a specific range of feeds with pagination support
func (f *Feed) ListRange(offset, limit int) ([]*models.Feed, error) {
	start := time.Now()

	if offset < 0 {
		return nil, utils.NewValidationError("offset: must be non-negative")
	}
	if limit <= 0 {
		return nil, utils.NewValidationError("limit: must be positive")
	}
	if limit > 1000 {
		return nil, utils.NewValidationError("limit: must not exceed 1000")
	}

	query := `SELECT id, title, content, user_id, created_at FROM feeds
		ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := f.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("select error: feeds: %v", err)
	}
	defer rows.Close()

	feeds, err := f.scanAllRows(rows)
	elapsed := time.Since(start)

	if err != nil {
		return nil, fmt.Errorf("scan error: feeds: %v", err)
	}

	if elapsed > 100*time.Millisecond {
		logger.DBFeeds().Warn("Slow feed range query took %v (offset=%d, limit=%d, returned=%d)", elapsed, offset, limit, len(feeds))
	}

	return feeds, nil
}

// ListByUser retrieves feeds created by a specific user
func (f *Feed) ListByUser(userID int64, offset, limit int) ([]*models.Feed, error) {
	start := time.Now()

	if userID <= 0 {
		return nil, utils.NewValidationError("user_id: must be positive")
	}
	if offset < 0 {
		return nil, utils.NewValidationError("offset: must be non-negative")
	}
	if limit <= 0 {
		return nil, utils.NewValidationError("limit: must be positive")
	}
	if limit > 1000 {
		return nil, utils.NewValidationError("limit: must not exceed 1000")
	}

	query := `SELECT id, title, content, user_id, created_at FROM feeds
		WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := f.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("select error: feeds: %v", err)
	}
	defer rows.Close()

	feeds, err := f.scanAllRows(rows)
	elapsed := time.Since(start)

	if err != nil {
		return nil, fmt.Errorf("scan error: feeds: %v", err)
	}

	if elapsed > 100*time.Millisecond {
		logger.DBFeeds().Warn("Slow user feeds query took %v (userID=%d, offset=%d, limit=%d, returned=%d)", elapsed, userID, offset, limit, len(feeds))
	}

	return feeds, nil
}

// Add creates a new feed entry in the database
func (f *Feed) Add(feedData *models.Feed) error {
	if feedData == nil {
		return utils.NewValidationError("feed: cannot be nil")
	}

	start := time.Now()

	if err := feedData.Validate(); err != nil {
		return err
	}

	query := `INSERT INTO feeds (title, content, user_id, created_at) VALUES (?, ?, ?, ?)`
	result, err := f.db.Exec(query, feedData.Title, feedData.Content, feedData.UserID, feedData.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert error: feeds: %v", err)
	}

	// Get the generated ID
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("insert error: feeds: %v", err)
	}
	feedData.ID = id

	// Notify about new feed if broadcaster is set
	if f.broadcaster != nil {
		f.broadcaster.Broadcast()
	}

	elapsed := time.Since(start)
	if elapsed > 50*time.Millisecond {
		logger.DBFeeds().Warn("Slow feed insert took %v for user %d", elapsed, feedData.UserID)
	}

	return nil
}

// Count returns the total number of feeds in the database
func (f *Feed) Count() (int, error) {
	start := time.Now()

	var count int
	query := `SELECT COUNT(*) FROM feeds`
	err := f.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count error: feeds: %v", err)
	}

	elapsed := time.Since(start)
	if elapsed > 50*time.Millisecond {
		logger.DBFeeds().Warn("Slow feed count query took %v (result: %d)", elapsed, count)
	}

	return count, nil
}

// CountByUser returns the number of feeds created by a specific user
func (f *Feed) CountByUser(userID int64) (int, error) {
	start := time.Now()

	if userID <= 0 {
		return 0, utils.NewValidationError("user_id: must be positive")
	}

	var count int
	query := `SELECT COUNT(*) FROM feeds WHERE user_id = ?`
	err := f.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count error: feeds: %v", err)
	}

	elapsed := time.Since(start)
	if elapsed > 50*time.Millisecond {
		logger.DBFeeds().Warn("Slow user feed count query took %v (userID=%d, result=%d)", elapsed, userID, count)
	}

	return count, nil
}

// DeleteBefore removes all feeds created before the specified timestamp
func (f *Feed) DeleteBefore(timestamp int64) (int64, error) {
	start := time.Now()

	if timestamp <= 0 {
		return 0, utils.NewValidationError("timestamp: must be positive")
	}

	query := `DELETE FROM feeds WHERE created_at < ?`
	result, err := f.db.Exec(query, timestamp)
	if err != nil {
		return 0, fmt.Errorf("delete error: feeds: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("delete error: feeds: %v", err)
	}

	elapsed := time.Since(start)
	logger.DBFeeds().Info("Deleted %d feeds before timestamp %d in %v", rowsAffected, timestamp, elapsed)

	if elapsed > 100*time.Millisecond {
		logger.DBFeeds().Warn("Slow feed deletion took %v (timestamp=%d, deleted=%d)", elapsed, timestamp, rowsAffected)
	}

	return rowsAffected, nil
}

// Get retrieves a specific feed by ID
func (f *Feed) Get(id int64) (*models.Feed, error) {
	if id <= 0 {
		return nil, utils.NewValidationError("id: must be positive")
	}

	query := `SELECT id, title, content, user_id, created_at FROM feeds WHERE id = ?`
	row := f.db.QueryRow(query, id)
	feedData, err := f.scanFeed(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("feed with id %d not found", id))
		}
		return nil, fmt.Errorf("select error: feeds: %v", err)
	}

	return feedData, nil
}

// Delete removes a specific feed by ID
func (f *Feed) Delete(id int64) error {
	if id <= 0 {
		return utils.NewValidationError("id: must be positive")
	}

	query := `DELETE FROM feeds WHERE id = ?`
	result, err := f.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("delete error: feeds: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete error: feeds: %v", err)
	}

	if rowsAffected == 0 {
		return utils.NewNotFoundError(fmt.Sprintf("feed with id %d not found", id))
	}
	return nil
}

// scanAllRows scans all rows from a query result into Feed structs
func (f *Feed) scanAllRows(rows *sql.Rows) ([]*models.Feed, error) {
	var feeds []*models.Feed
	scanStart := time.Now()

	for rows.Next() {
		feedData, err := f.scanFeed(rows)
		if err != nil {
			return nil, fmt.Errorf("scan error: feeds: %v", err)
		}
		feeds = append(feeds, feedData)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan error: feeds: %v", err)
	}

	scanElapsed := time.Since(scanStart)
	if scanElapsed > 50*time.Millisecond {
		logger.DBFeeds().Warn("Slow feed row scanning took %v for %d rows", scanElapsed, len(feeds))
	}

	return feeds, nil
}

func (f *Feed) scanFeed(scanner interfaces.Scannable) (*models.Feed, error) {
	feedData := &models.Feed{}

	if err := scanner.Scan(&feedData.ID, &feedData.Title, &feedData.Content, &feedData.UserID, &feedData.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to scan row: %v", err)
	}

	return feedData, nil
}
