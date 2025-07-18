// NOTE: Cleaned up by AI
package pgvis

import (
	"database/sql"
	"encoding/json"
)

const (
	createFeedsTableQuery = `
		CREATE TABLE IF NOT EXISTS feeds (
			id INTEGER NOT NULL,
			time INTEGER NOT NULL,
			cache BLOB NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`

	selectAllFeedsQuery    = `SELECT id, time, cache FROM feeds ORDER BY id DESC`
	selectFeedsRangeQuery  = `SELECT id, time, cache FROM feeds ORDER BY id DESC LIMIT ? OFFSET ?`
	selectFeedByIDQuery    = `SELECT id, time, cache FROM feeds WHERE id = ?`
	insertFeedQuery        = `INSERT INTO feeds (time, cache) VALUES (?, ?)`
	countFeedsQuery        = `SELECT COUNT(*) FROM feeds`
	deleteFeedsByTimeQuery = `DELETE FROM feeds WHERE time < ?`
	deleteFeedByIDQuery    = `DELETE FROM feeds WHERE id = ?`
)

// Feeds handles database operations for feed entries
type Feeds struct {
	db *sql.DB
}

// NewFeeds creates a new Feeds instance and initializes the database table
func NewFeeds(db *sql.DB) *Feeds {
	if _, err := db.Exec(createFeedsTableQuery); err != nil {
		panic(NewDatabaseError("create table", "feeds", "failed to create feeds table", err))
	}

	return &Feeds{db: db}
}

// List retrieves all feeds ordered by ID in descending order
func (f *Feeds) List() ([]*Feed, error) {
	rows, err := f.db.Query(selectAllFeedsQuery)
	if err != nil {
		return nil, NewDatabaseError("select", "feeds", "failed to query feeds", err)
	}
	defer rows.Close()

	return f.scanAllRows(rows)
}

// ListRange retrieves a specific range of feeds with pagination support
func (f *Feeds) ListRange(offset, limit int) ([]*Feed, error) {
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

	feeds, err := f.scanAllRows(rows)
	if err != nil {
		return nil, WrapErrorf(err, "failed to scan feeds range (offset: %d, limit: %d)", offset, limit)
	}

	return feeds, nil
}

// Add creates a new feed entry in the database
func (f *Feeds) Add(feed *Feed) error {
	if feed == nil {
		return NewValidationError("feed", "cannot be nil", nil)
	}

	if err := feed.Validate(); err != nil {
		return err
	}

	cache, err := json.Marshal(feed.Cache)
	if err != nil {
		return WrapError(err, "failed to marshal feed cache")
	}

	_, err = f.db.Exec(insertFeedQuery, feed.Time, cache)
	if err != nil {
		return NewDatabaseError("insert", "feeds", "failed to insert feed", err)
	}

	return nil
}

// Count returns the total number of feeds in the database
func (f *Feeds) Count() (int, error) {
	var count int
	err := f.db.QueryRow(countFeedsQuery).Scan(&count)
	if err != nil {
		return 0, NewDatabaseError("count", "feeds", "failed to count feeds", err)
	}
	return count, nil
}

// DeleteBefore removes all feeds created before the specified timestamp
func (f *Feeds) DeleteBefore(timestamp int64) (int64, error) {
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
	var cache []byte

	if err := scanFunc(&feed.ID, &feed.Time, &cache); err != nil {
		return nil, err
	}

	if len(cache) > 0 {
		if err := json.Unmarshal(cache, &feed.Cache); err != nil {
			return nil, WrapError(err, "failed to unmarshal cache data")
		}
	}

	return feed, nil
}
