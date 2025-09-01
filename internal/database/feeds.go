package database

import (
	"database/sql"
	"encoding/json"

	"github.com/knackwurstking/pgpress/internal/logger"
)

// Feeds handles database operations for feed entries
type Feeds struct {
	db          *sql.DB
	broadcaster Broadcaster
}

// NewFeeds creates a new Feeds instance and initializes the database table
func NewFeeds(db *sql.DB) *Feeds {
	query := `
		CREATE TABLE IF NOT EXISTS feeds (
			id INTEGER NOT NULL,
			time INTEGER NOT NULL,
			data_type TEXT NOT NULL,
			data BLOB NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`
	if _, err := db.Exec(query); err != nil {
		panic(NewDatabaseError("create table", "feeds", "failed to create feeds table", err))
	}
	return &Feeds{db: db}
}

// SetNotifier sets the feed notifier for real-time updates
func (f *Feeds) SetBroadcaster(notifier Broadcaster) {
	logger.DBFeeds().Debug("Setting broadcaster for real-time updates")
	f.broadcaster = notifier
}

// List retrieves all feeds ordered by ID in descending order
func (f *Feeds) List() ([]*Feed, error) {
	logger.DBFeeds().Info("Listing all feeds")

	query := `SELECT id, time, data_type, data FROM feeds ORDER BY id DESC`
	rows, err := f.db.Query(query)
	if err != nil {
		return nil, NewDatabaseError("select", "feeds", "failed to query feeds", err)
	}
	defer rows.Close()

	return f.scanAllRows(rows)
}

// ListRange retrieves a specific range of feeds with pagination support
func (f *Feeds) ListRange(offset, limit int) ([]*Feed, error) {
	logger.DBFeeds().Info("Listing range of feeds, offset: %d, limit: %d", offset, limit)

	if offset < 0 {
		return nil, NewValidationError("offset", "must be non-negative", offset)
	}
	if limit <= 0 {
		return nil, NewValidationError("limit", "must be positive", limit)
	}
	if limit > 1000 {
		return nil, NewValidationError("limit", "must not exceed 1000", limit)
	}

	query := `SELECT id, time, data_type, data FROM feeds
		ORDER BY id DESC LIMIT ? OFFSET ?`
	rows, err := f.db.Query(query, limit, offset)
	if err != nil {
		return nil, NewDatabaseError("select", "feeds", "failed to query feeds range", err)
	}
	defer rows.Close()

	return f.scanAllRows(rows)
}

// Add creates a new feed entry in the database
func (f *Feeds) Add(feed *Feed) error {
	logger.DBFeeds().Info("Adding feed: %+v", feed)

	if feed == nil {
		logger.DBFeeds().Debug("Validation failed: feed is nil")
		return NewValidationError("feed", "cannot be nil", nil)
	}
	if err := feed.Validate(); err != nil {
		logger.DBFeeds().Debug("Feed validation failed: %v", err)
		return err
	}

	data, err := json.Marshal(feed.Data)
	if err != nil {
		logger.DBFeeds().Error("Failed to marshal feed data: %v", err)
		return WrapError(err, "failed to marshal feed data")
	}

	query := `INSERT INTO feeds (time, data_type, data) VALUES (?, ?, ?)`
	_, err = f.db.Exec(query, feed.Time, feed.DataType, data)
	if err != nil {
		logger.DBFeeds().Error("Failed to insert feed: %v", err)
		return NewDatabaseError("insert", "feeds", "failed to insert feed", err)
	}

	// Notify about new feed if notifier is set
	if f.broadcaster != nil {
		logger.DBFeeds().Debug("Broadcasting new feed notification")
		f.broadcaster.Broadcast()
	}

	logger.DBFeeds().Debug("Successfully added feed with data type: %s", feed.DataType)
	return nil
}

// Count returns the total number of feeds in the database
func (f *Feeds) Count() (int, error) {
	logger.DBFeeds().Debug("Counting feeds")

	var count int
	query := `SELECT COUNT(*) FROM feeds`
	err := f.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, NewDatabaseError("count", "feeds", "failed to count feeds", err)
	}
	return count, nil
}

// DeleteBefore removes all feeds created before the specified timestamp
func (f *Feeds) DeleteBefore(timestamp int64) (int64, error) {
	logger.DBFeeds().Info("Deleting feeds before timestamp: %d", timestamp)

	if timestamp <= 0 {
		logger.DBFeeds().Debug("Validation failed: timestamp must be positive, got %d", timestamp)
		return 0, NewValidationError("timestamp", "must be positive", timestamp)
	}

	query := `DELETE FROM feeds WHERE time < ?`
	result, err := f.db.Exec(query, timestamp)
	if err != nil {
		logger.DBFeeds().Error("Failed to delete feeds by timestamp: %v", err)
		return 0, NewDatabaseError("delete", "feeds", "failed to delete feeds by timestamp", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.DBFeeds().Error("Failed to get rows affected: %v", err)
		return 0, NewDatabaseError("delete", "feeds", "failed to get rows affected", err)
	}

	logger.DBFeeds().Debug("Deleted %d feeds before timestamp %d", rowsAffected, timestamp)
	return rowsAffected, nil
}

// Get retrieves a specific feed by ID
func (f *Feeds) Get(id int) (*Feed, error) {
	logger.DBFeeds().Debug("Getting feed by ID: %d", id)

	if id <= 0 {
		return nil, NewValidationError("id", "must be positive", id)
	}

	query := `SELECT id, time, data_type, data FROM feeds WHERE id = ?`
	row := f.db.QueryRow(query, id)
	feed, err := f.scanFeed(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, NewDatabaseError("select", "feeds", "failed to get feed by ID", err)
	}
	return feed, nil
}

// Delete removes a specific feed by ID
func (f *Feeds) Delete(id int) error {
	logger.DBFeeds().Info("Deleting feed by ID: %d", id)

	if id <= 0 {
		logger.DBFeeds().Debug("Validation failed: id must be positive, got %d", id)
		return NewValidationError("id", "must be positive", id)
	}

	query := `DELETE FROM feeds WHERE id = ?`
	result, err := f.db.Exec(query, id)
	if err != nil {
		logger.DBFeeds().Error("Failed to delete feed with ID %d: %v", id, err)
		return NewDatabaseError("delete", "feeds", "failed to delete feed", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.DBFeeds().Error("Failed to get rows affected: %v", err)
		return NewDatabaseError("delete", "feeds", "failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		logger.DBFeeds().Debug("No feed found with ID %d", id)
		return ErrNotFound
	}

	logger.DBFeeds().Debug("Successfully deleted feed with ID %d", id)
	return nil
}

// scanAllRows scans all rows from a query result into Feed structs
func (f *Feeds) scanAllRows(rows *sql.Rows) ([]*Feed, error) {
	var feeds []*Feed
	for rows.Next() {
		feed, err := f.scanFeed(rows)
		if err != nil {
			logger.DBFeeds().Error("Failed to scan feed row: %v", err)
			return nil, NewDatabaseError("scan", "feeds", "failed to scan feed row", err)
		}
		feeds = append(feeds, feed)
	}

	if err := rows.Err(); err != nil {
		logger.DBFeeds().Error("Error iterating over feed rows: %v", err)
		return nil, NewDatabaseError("scan", "feeds", "error iterating over rows", err)
	}

	logger.DBFeeds().Debug("Scanned %d feeds from query result", len(feeds))
	return feeds, nil
}

func (f *Feeds) scanFeed(scanner scannable) (*Feed, error) {
	feed := &Feed{}
	var data []byte

	if err := scanner.Scan(&feed.ID, &feed.Time, &feed.DataType, &data); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, NewDatabaseError("scan", "feeds", "failed to scan row", err)
	}

	if len(data) > 0 {
		if err := json.Unmarshal(data, &feed.Data); err != nil {
			logger.DBFeeds().Error("Failed to unmarshal feed data for ID %d: %v", feed.ID, err)
			return nil, WrapError(err, "failed to unmarshal feed data")
		}
	}

	return feed, nil
}
