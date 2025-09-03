// TODO: Need to fix this type to fit the `interfaces.DataOperations[*models.Cookie]` type
package feed

import (
	"database/sql"
	"encoding/json"

	"github.com/knackwurstking/pgpress/internal/database/errors"
	"github.com/knackwurstking/pgpress/internal/database/interfaces"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/database/models"
)

// Service handles database operations for feed entries
type Service struct {
	db          *sql.DB
	broadcaster interfaces.Broadcaster
}

// New creates a new Service instance and initializes the database table
func New(db *sql.DB) *Service {
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
		panic(dberror.NewDatabaseError(
			"create table",
			"feeds",
			"failed to create feeds table",
			err,
		))
	}
	return &Service{db: db}
}

// SetNotifier sets the feed notifier for real-time updates
func (s *Service) SetBroadcaster(notifier interfaces.Broadcaster) {
	logger.DBFeeds().Debug("Setting broadcaster for real-time updates")
	s.broadcaster = notifier
}

// List retrieves all feeds ordered by ID in descending order
func (s *Service) List() ([]*models.Feed, error) {
	logger.DBFeeds().Info("Listing all feeds")

	query := `SELECT id, time, data_type, data FROM feeds ORDER BY id DESC`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, dberror.NewDatabaseError("select", "feeds", "failed to query feeds", err)
	}
	defer rows.Close()

	return s.scanAllRows(rows)
}

// ListRange retrieves a specific range of feeds with pagination support
func (s *Service) ListRange(offset, limit int) ([]*models.Feed, error) {
	logger.DBFeeds().Info("Listing range of feeds, offset: %d, limit: %d", offset, limit)

	if offset < 0 {
		return nil, dberror.NewValidationError("offset", "must be non-negative", offset)
	}
	if limit <= 0 {
		return nil, dberror.NewValidationError("limit", "must be positive", limit)
	}
	if limit > 1000 {
		return nil, dberror.NewValidationError("limit", "must not exceed 1000", limit)
	}

	query := `SELECT id, time, data_type, data FROM feeds
		ORDER BY id DESC LIMIT ? OFFSET ?`
	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, dberror.NewDatabaseError("select", "feeds", "failed to query feeds range", err)
	}
	defer rows.Close()

	return s.scanAllRows(rows)
}

// Add creates a new feed entry in the database
func (s *Service) Add(feed *models.Feed) error {
	logger.DBFeeds().Info("Adding feed: %+v", feed)

	if feed == nil {
		logger.DBFeeds().Debug("Validation failed: feed is nil")
		return dberror.NewValidationError("feed", "cannot be nil", nil)
	}
	if err := feed.Validate(); err != nil {
		logger.DBFeeds().Debug("Feed validation failed: %v", err)
		return err
	}

	data, err := json.Marshal(feed.Data)
	if err != nil {
		logger.DBFeeds().Error("Failed to marshal feed data: %v", err)
		return dberror.WrapError(err, "failed to marshal feed data")
	}

	query := `INSERT INTO feeds (time, data_type, data) VALUES (?, ?, ?)`
	_, err = s.db.Exec(query, feed.Time, feed.DataType, data)
	if err != nil {
		logger.DBFeeds().Error("Failed to insert feed: %v", err)
		return dberror.NewDatabaseError("insert", "feeds", "failed to insert feed", err)
	}

	// Notify about new feed if notifier is set
	if s.broadcaster != nil {
		logger.DBFeeds().Debug("Broadcasting new feed notification")
		s.broadcaster.Broadcast()
	}

	logger.DBFeeds().Debug("Successfully added feed with data type: %s", feed.DataType)
	return nil
}

// Count returns the total number of feeds in the database
func (s *Service) Count() (int, error) {
	logger.DBFeeds().Debug("Counting feeds")

	var count int
	query := `SELECT COUNT(*) FROM feeds`
	err := s.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, dberror.NewDatabaseError("count", "feeds", "failed to count feeds", err)
	}
	return count, nil
}

// DeleteBefore removes all feeds created before the specified timestamp
func (s *Service) DeleteBefore(timestamp int64) (int64, error) {
	logger.DBFeeds().Info("Deleting feeds before timestamp: %d", timestamp)

	if timestamp <= 0 {
		logger.DBFeeds().Debug("Validation failed: timestamp must be positive, got %d", timestamp)
		return 0, dberror.NewValidationError("timestamp", "must be positive", timestamp)
	}

	query := `DELETE FROM feeds WHERE time < ?`
	result, err := s.db.Exec(query, timestamp)
	if err != nil {
		logger.DBFeeds().Error("Failed to delete feeds by timestamp: %v", err)
		return 0, dberror.NewDatabaseError("delete", "feeds", "failed to delete feeds by timestamp", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.DBFeeds().Error("Failed to get rows affected: %v", err)
		return 0, dberror.NewDatabaseError("delete", "feeds", "failed to get rows affected", err)
	}

	logger.DBFeeds().Debug("Deleted %d feeds before timestamp %d", rowsAffected, timestamp)
	return rowsAffected, nil
}

// Get retrieves a specific feed by ID
func (s *Service) Get(id int) (*models.Feed, error) {
	logger.DBFeeds().Debug("Getting feed by ID: %d", id)

	if id <= 0 {
		return nil, dberror.NewValidationError("id", "must be positive", id)
	}

	query := `SELECT id, time, data_type, data FROM feeds WHERE id = ?`
	row := s.db.QueryRow(query, id)
	feed, err := s.scanFeed(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, dberror.ErrNotFound
		}
		return nil, dberror.NewDatabaseError("select", "feeds", "failed to get feed by ID", err)
	}
	return feed, nil
}

// Delete removes a specific feed by ID
func (s *Service) Delete(id int) error {
	logger.DBFeeds().Info("Deleting feed by ID: %d", id)

	if id <= 0 {
		logger.DBFeeds().Debug("Validation failed: id must be positive, got %d", id)
		return dberror.NewValidationError("id", "must be positive", id)
	}

	query := `DELETE FROM feeds WHERE id = ?`
	result, err := s.db.Exec(query, id)
	if err != nil {
		logger.DBFeeds().Error("Failed to delete feed with ID %d: %v", id, err)
		return dberror.NewDatabaseError("delete", "feeds", "failed to delete feed", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.DBFeeds().Error("Failed to get rows affected: %v", err)
		return dberror.NewDatabaseError("delete", "feeds", "failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		logger.DBFeeds().Debug("No feed found with ID %d", id)
		return dberror.ErrNotFound
	}

	logger.DBFeeds().Debug("Successfully deleted feed with ID %d", id)
	return nil
}

// scanAllRows scans all rows from a query result into Feed structs
func (s *Service) scanAllRows(rows *sql.Rows) ([]*models.Feed, error) {
	var feeds []*models.Feed
	for rows.Next() {
		feed, err := s.scanFeed(rows)
		if err != nil {
			logger.DBFeeds().Error("Failed to scan feed row: %v", err)
			return nil, dberror.NewDatabaseError("scan", "feeds", "failed to scan feed row", err)
		}
		feeds = append(feeds, feed)
	}

	if err := rows.Err(); err != nil {
		logger.DBFeeds().Error("Error iterating over feed rows: %v", err)
		return nil, dberror.NewDatabaseError("scan", "feeds", "error iterating over rows", err)
	}

	logger.DBFeeds().Debug("Scanned %d feeds from query result", len(feeds))
	return feeds, nil
}

func (s *Service) scanFeed(scanner interfaces.Scannable) (*models.Feed, error) {
	feed := &models.Feed{}
	var data []byte

	if err := scanner.Scan(&feed.ID, &feed.Time, &feed.DataType, &data); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, dberror.NewDatabaseError("scan", "feeds", "failed to scan row", err)
	}

	if len(data) > 0 {
		if err := json.Unmarshal(data, &feed.Data); err != nil {
			logger.DBFeeds().Error("Failed to unmarshal feed data for ID %d: %v", feed.ID, err)
			return nil, dberror.WrapError(err, "failed to unmarshal feed data")
		}
	}

	return feed, nil
}
