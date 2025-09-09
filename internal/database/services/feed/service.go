package feed

import (
	"database/sql"

	"github.com/knackwurstking/pgpress/internal/database/dberror"
	"github.com/knackwurstking/pgpress/internal/database/interfaces"
	"github.com/knackwurstking/pgpress/internal/database/models/feed"
	"github.com/knackwurstking/pgpress/internal/logger"
)

// Service handles database operations for feed entries
type Service struct {
	db          *sql.DB
	broadcaster interfaces.Broadcaster
}

// New creates a new Service instance and initializes the database table
func New(db *sql.DB) *Service {
	//dropQuery := `DROP TABLE IF EXISTS feeds;`
	//if _, err := db.Exec(dropQuery); err != nil {
	//	panic(fmt.Errorf("failed to drop feeds table: %w", err))
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
		panic(dberror.NewDatabaseError(
			"create table",
			"feeds",
			"failed to create feeds table",
			err,
		))
	}
	return &Service{db: db}
}

// SetBroadcaster sets the feed notifier for real-time updates
func (s *Service) SetBroadcaster(broadcaster interfaces.Broadcaster) {
	logger.DBFeeds().Debug("Setting broadcaster for real-time updates")
	s.broadcaster = broadcaster
}

// List retrieves all feeds ordered by creation time in descending order
func (s *Service) List() ([]*feed.Feed, error) {
	logger.DBFeeds().Info("Listing all feeds")

	query := `SELECT id, title, content, user_id, created_at FROM feeds ORDER BY created_at DESC`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, dberror.NewDatabaseError("select", "feeds", "failed to query feeds", err)
	}
	defer rows.Close()

	return s.scanAllRows(rows)
}

// ListRange retrieves a specific range of feeds with pagination support
func (s *Service) ListRange(offset, limit int) ([]*feed.Feed, error) {
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

	query := `SELECT id, title, content, user_id, created_at FROM feeds
		ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, dberror.NewDatabaseError("select", "feeds", "failed to query feeds range", err)
	}
	defer rows.Close()

	return s.scanAllRows(rows)
}

// ListByUser retrieves feeds created by a specific user
func (s *Service) ListByUser(userID int64, offset, limit int) ([]*feed.Feed, error) {
	logger.DBFeeds().Info("Listing feeds for user %d, offset: %d, limit: %d", userID, offset, limit)

	if userID <= 0 {
		return nil, dberror.NewValidationError("user_id", "must be positive", userID)
	}
	if offset < 0 {
		return nil, dberror.NewValidationError("offset", "must be non-negative", offset)
	}
	if limit <= 0 {
		return nil, dberror.NewValidationError("limit", "must be positive", limit)
	}
	if limit > 1000 {
		return nil, dberror.NewValidationError("limit", "must not exceed 1000", limit)
	}

	query := `SELECT id, title, content, user_id, created_at FROM feeds
		WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := s.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, dberror.NewDatabaseError("select", "feeds", "failed to query feeds by user", err)
	}
	defer rows.Close()

	return s.scanAllRows(rows)
}

// Add creates a new feed entry in the database
func (s *Service) Add(feedData *feed.Feed) error {
	logger.DBFeeds().Info("Adding feed: %+v", feedData)

	if feedData == nil {
		logger.DBFeeds().Debug("Validation failed: feed is nil")
		return dberror.NewValidationError("feed", "cannot be nil", nil)
	}
	if err := feedData.Validate(); err != nil {
		logger.DBFeeds().Debug("Feed validation failed: %v", err)
		return err
	}

	query := `INSERT INTO feeds (title, content, user_id, created_at) VALUES (?, ?, ?, ?)`
	result, err := s.db.Exec(query, feedData.Title, feedData.Content, feedData.UserID, feedData.CreatedAt)
	if err != nil {
		logger.DBFeeds().Error("Failed to insert feed: %v", err)
		return dberror.NewDatabaseError("insert", "feeds", "failed to insert feed", err)
	}

	// Get the generated ID
	id, err := result.LastInsertId()
	if err != nil {
		logger.DBFeeds().Error("Failed to get last insert ID: %v", err)
		return dberror.NewDatabaseError("insert", "feeds", "failed to get insert ID", err)
	}
	feedData.ID = id

	// Notify about new feed if broadcaster is set
	if s.broadcaster != nil {
		logger.DBFeeds().Debug("Broadcasting new feed notification")
		s.broadcaster.Broadcast()
	}

	logger.DBFeeds().Debug("Successfully added feed with ID: %d", feedData.ID)
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

// CountByUser returns the number of feeds created by a specific user
func (s *Service) CountByUser(userID int64) (int, error) {
	logger.DBFeeds().Debug("Counting feeds for user %d", userID)

	if userID <= 0 {
		return 0, dberror.NewValidationError("user_id", "must be positive", userID)
	}

	var count int
	query := `SELECT COUNT(*) FROM feeds WHERE user_id = ?`
	err := s.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, dberror.NewDatabaseError("count", "feeds", "failed to count feeds by user", err)
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

	query := `DELETE FROM feeds WHERE created_at < ?`
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
func (s *Service) Get(id int64) (*feed.Feed, error) {
	logger.DBFeeds().Debug("Getting feed by ID: %d", id)

	if id <= 0 {
		return nil, dberror.NewValidationError("id", "must be positive", id)
	}

	query := `SELECT id, title, content, user_id, created_at FROM feeds WHERE id = ?`
	row := s.db.QueryRow(query, id)
	feedData, err := s.scanFeed(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, dberror.ErrNotFound
		}
		return nil, dberror.NewDatabaseError("select", "feeds", "failed to get feed by ID", err)
	}
	return feedData, nil
}

// Delete removes a specific feed by ID
func (s *Service) Delete(id int64) error {
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
func (s *Service) scanAllRows(rows *sql.Rows) ([]*feed.Feed, error) {
	var feeds []*feed.Feed
	for rows.Next() {
		feedData, err := s.scanFeed(rows)
		if err != nil {
			logger.DBFeeds().Error("Failed to scan feed row: %v", err)
			return nil, dberror.NewDatabaseError("scan", "feeds", "failed to scan feed row", err)
		}
		feeds = append(feeds, feedData)
	}

	if err := rows.Err(); err != nil {
		logger.DBFeeds().Error("Error iterating over feed rows: %v", err)
		return nil, dberror.NewDatabaseError("scan", "feeds", "error iterating over rows", err)
	}

	logger.DBFeeds().Debug("Scanned %d feeds from query result", len(feeds))
	return feeds, nil
}

func (s *Service) scanFeed(scanner interfaces.Scannable) (*feed.Feed, error) {
	feedData := &feed.Feed{}

	if err := scanner.Scan(&feedData.ID, &feedData.Title, &feedData.Content, &feedData.UserID, &feedData.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, dberror.NewDatabaseError("scan", "feeds", "failed to scan row", err)
	}

	return feedData, nil
}
