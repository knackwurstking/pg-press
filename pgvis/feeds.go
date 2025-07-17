package pgvis

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

const (
	// SQL queries for feeds table
	createFeedsTableQuery = `
		CREATE TABLE IF NOT EXISTS feeds (
			id INTEGER NOT NULL,
			time INTEGER NOT NULL,
			main TEXT NOT NULL,
			cache BLOB NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`

	selectAllFeedsQuery    = `SELECT id, time, main, cache FROM feeds ORDER BY id DESC`
	selectFeedsRangeQuery  = `SELECT id, time, main, cache FROM feeds ORDER BY id DESC LIMIT ? OFFSET ?`
	insertFeedQuery        = `INSERT INTO feeds (time, main, cache) VALUES (?, ?, ?)`
	countFeedsQuery        = `SELECT COUNT(*) FROM feeds`
	deleteFeedsByTimeQuery = `DELETE FROM feeds WHERE time < ?`
)

// Feeds handles database operations for feed entries
type Feeds struct {
	db *sql.DB
}

// NewFeeds creates a new Feeds instance and initializes the database table
func NewFeeds(db *sql.DB) *Feeds {
	if _, err := db.Exec(createFeedsTableQuery); err != nil {
		panic(fmt.Errorf("failed to create feeds table: %w", err))
	}

	return &Feeds{
		db: db,
	}
}

// List retrieves all feeds ordered by ID in descending order
func (f *Feeds) List() ([]*Feed, error) {
	rows, err := f.db.Query(selectAllFeedsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query feeds: %w", err)
	}
	defer rows.Close()

	feeds, err := f.scanAllRows(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan feeds: %w", err)
	}

	return feeds, nil
}

// ListRange retrieves a specific range of feeds with pagination support
func (f *Feeds) ListRange(offset, limit int) ([]*Feed, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 10 // Default limit
	}

	rows, err := f.db.Query(selectFeedsRangeQuery, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query feeds range (offset: %d, limit: %d): %w", offset, limit, err)
	}
	defer rows.Close()

	feeds, err := f.scanAllRows(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan feeds range: %w", err)
	}

	return feeds, nil
}

// Add creates a new feed entry in the database
func (f *Feeds) Add(feed *Feed) error {
	if feed == nil {
		return fmt.Errorf("feed cannot be nil")
	}

	cache, err := json.Marshal(feed.Cache)
	if err != nil {
		return fmt.Errorf("failed to marshal feed cache: %w", err)
	}

	_, err = f.db.Exec(insertFeedQuery, feed.Time, feed.Main, cache)
	if err != nil {
		return fmt.Errorf("failed to insert feed: %w", err)
	}

	return nil
}

// Count returns the total number of feeds in the database
func (f *Feeds) Count() (int, error) {
	var count int
	err := f.db.QueryRow(countFeedsQuery).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count feeds: %w", err)
	}
	return count, nil
}

// DeleteBefore removes all feeds created before the specified timestamp
func (f *Feeds) DeleteBefore(timestamp int64) (int64, error) {
	result, err := f.db.Exec(deleteFeedsByTimeQuery, timestamp)
	if err != nil {
		return 0, fmt.Errorf("failed to delete feeds before timestamp %d: %w", timestamp, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// scanAllRows scans all rows from a query result into Feed structs
func (f *Feeds) scanAllRows(rows *sql.Rows) ([]*Feed, error) {
	var feeds []*Feed

	for rows.Next() {
		feed, err := f.scanFeed(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan feed row: %w", err)
		}
		feeds = append(feeds, feed)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return feeds, nil
}

// scanFeed scans a single feed from a database row
func (f *Feeds) scanFeed(rows *sql.Rows) (*Feed, error) {
	feed := &Feed{}
	var cache []byte

	if err := rows.Scan(&feed.ID, &feed.Time, &feed.Main, &cache); err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	if len(cache) > 0 {
		if err := json.Unmarshal(cache, &feed.Cache); err != nil {
			return nil, fmt.Errorf("failed to unmarshal cache data: %w", err)
		}
	}

	return feed, nil
}
