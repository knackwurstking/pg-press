package database

import (
	"database/sql"
	"encoding/json"

	"github.com/knackwurstking/pg-vis/internal/logger"
)

// FeedNotifier interface for handling feed update notifications
type FeedNotifier interface {
	NotifyNewFeed()
}

const (
	createFeedsTableQuery = `
		CREATE TABLE IF NOT EXISTS feeds (
			id INTEGER NOT NULL,
			time INTEGER NOT NULL,
			data_type TEXT NOT NULL,
			data BLOB NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`

	selectAllFeedsQuery    = `SELECT id, time, data_type, data FROM feeds ORDER BY id DESC`
	selectFeedsRangeQuery  = `SELECT id, time, data_type, data FROM feeds ORDER BY id DESC LIMIT ? OFFSET ?`
	selectFeedByIDQuery    = `SELECT id, time, data_type, data FROM feeds WHERE id = ?`
	insertFeedQuery        = `INSERT INTO feeds (time, data_type, data) VALUES (?, ?, ?)`
	countFeedsQuery        = `SELECT COUNT(*) FROM feeds`
	deleteFeedsByTimeQuery = `DELETE FROM feeds WHERE time < ?`
	deleteFeedByIDQuery    = `DELETE FROM feeds WHERE id = ?`
)

// Feeds handles database operations for feed entries
type Feeds struct {
	db       *sql.DB
	notifier FeedNotifier
}

// NewFeeds creates a new Feeds instance and initializes the database table
func NewFeeds(db *sql.DB) *Feeds {
	if _, err := db.Exec(createFeedsTableQuery); err != nil {
		panic(NewDatabaseError("create table", "feeds", "failed to create feeds table", err))
	}
	return &Feeds{db: db}
}

// SetNotifier sets the feed notifier for real-time updates
func (f *Feeds) SetNotifier(notifier FeedNotifier) {
	f.notifier = notifier
}

// List retrieves all feeds ordered by ID in descending order
func (f *Feeds) List() ([]*Feed, error) {
	logger.Feed().Info("Listing all feeds")

	rows, err := f.db.Query(selectAllFeedsQuery)
	if err != nil {
		return nil, NewDatabaseError("select", "feeds", "failed to query feeds", err)
	}
	defer rows.Close()

	return f.scanAllRows(rows)
}

// ListRange retrieves a specific range of feeds with pagination support
func (f *Feeds) ListRange(offset, limit int) ([]*Feed, error) {
	logger.Feed().Info("Listing range of feeds, offset: %d, limit: %d", offset, limit)

	if offset < 0 {
		return nil, NewValidationError("offset", "must be non-negative", offset)
	}
	if limit <= 0 {
		return nil, NewValidationError("limit", "must be positive", limit)
	}
	if limit > 1000 {
		return nil, NewValidationError("limit", "must not exceed 1000", limit)
	}

	rows, err := f.db.Query(selectFeedsRangeQuery, limit, offset)
	if err != nil {
		return nil, NewDatabaseError("select", "feeds", "failed to query feeds range", err)
	}
	defer rows.Close()

	return f.scanAllRows(rows)
}

// Add creates a new feed entry in the database
func (f *Feeds) Add(feed *Feed) error {
	logger.Feed().Info("Adding feed: %+v", feed)

	if feed == nil {
		return NewValidationError("feed", "cannot be nil", nil)
	}
	if err := feed.Validate(); err != nil {
		return err
	}

	data, err := json.Marshal(feed.Data)
	if err != nil {
		return WrapError(err, "failed to marshal feed data")
	}

	_, err = f.db.Exec(insertFeedQuery, feed.Time, feed.DataType, data)
	if err != nil {
		return NewDatabaseError("insert", "feeds", "failed to insert feed", err)
	}

	// Notify about new feed if notifier is set
	if f.notifier != nil {
		f.notifier.NotifyNewFeed()
	}

	return nil
}

// Count returns the total number of feeds in the database
func (f *Feeds) Count() (int, error) {
	logger.Feed().Debug("Counting feeds")

	var count int
	err := f.db.QueryRow(countFeedsQuery).Scan(&count)
	if err != nil {
		return 0, NewDatabaseError("count", "feeds", "failed to count feeds", err)
	}
	return count, nil
}

// DeleteBefore removes all feeds created before the specified timestamp
func (f *Feeds) DeleteBefore(timestamp int64) (int64, error) {
	logger.Feed().Info("Deleting feeds before timestamp: %d", timestamp)

	if timestamp <= 0 {
		return 0, NewValidationError("timestamp", "must be positive", timestamp)
	}

	result, err := f.db.Exec(deleteFeedsByTimeQuery, timestamp)
	if err != nil {
		return 0, NewDatabaseError("delete", "feeds", "failed to delete feeds by timestamp", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, NewDatabaseError("delete", "feeds", "failed to get rows affected", err)
	}
	return rowsAffected, nil
}

// Get retrieves a specific feed by ID
func (f *Feeds) Get(id int) (*Feed, error) {
	logger.Feed().Debug("Getting feed by ID: %d", id)

	if id <= 0 {
		return nil, NewValidationError("id", "must be positive", id)
	}

	row := f.db.QueryRow(selectFeedByIDQuery, id)
	feed, err := f.scanFeedRow(row)
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
	logger.Feed().Info("Deleting feed by ID: %d", id)

	if id <= 0 {
		return NewValidationError("id", "must be positive", id)
	}

	result, err := f.db.Exec(deleteFeedByIDQuery, id)
	if err != nil {
		return NewDatabaseError("delete", "feeds", "failed to delete feed", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return NewDatabaseError("delete", "feeds", "failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// scanAllRows scans all rows from a query result into Feed structs
func (f *Feeds) scanAllRows(rows *sql.Rows) ([]*Feed, error) {
	var feeds []*Feed
	for rows.Next() {
		feed, err := f.scanFeed(rows)
		if err != nil {
			return nil, NewDatabaseError("scan", "feeds", "failed to scan feed row", err)
		}
		feeds = append(feeds, feed)
	}

	if err := rows.Err(); err != nil {
		return nil, NewDatabaseError("scan", "feeds", "error iterating over rows", err)
	}
	return feeds, nil
}

// scanFeed scans a single feed from a database row
func (f *Feeds) scanFeed(rows *sql.Rows) (*Feed, error) {
	return f.scanFeedData(rows.Scan)
}

// scanFeedRow scans a single feed from a query row
func (f *Feeds) scanFeedRow(row *sql.Row) (*Feed, error) {
	return f.scanFeedData(row.Scan)
}

// scanFeedData scans feed data using the provided scan function
func (f *Feeds) scanFeedData(scanFunc func(dest ...any) error) (*Feed, error) {
	feed := &Feed{}
	var data []byte

	if err := scanFunc(&feed.ID, &feed.Time, &feed.DataType, &data); err != nil {
		return nil, err
	}

	if len(data) > 0 {
		if err := json.Unmarshal(data, &feed.Data); err != nil {
			return nil, WrapError(err, "failed to unmarshal feed data")
		}
	}
	return feed, nil
}
