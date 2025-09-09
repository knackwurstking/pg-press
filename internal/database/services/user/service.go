// Package database user management.
//
// This file provides database operations for managing users,
// including CRUD operations, authentication via API keys, and integration
// with the activity feed system. All database operations use parameterized
// queries to prevent SQL injection attacks.
package user

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/knackwurstking/pgpress/internal/database/dberror"
	"github.com/knackwurstking/pgpress/internal/database/interfaces"
	"github.com/knackwurstking/pgpress/internal/database/models"
	"github.com/knackwurstking/pgpress/internal/database/services/feed"
	dbutils "github.com/knackwurstking/pgpress/internal/database/utils"
	"github.com/knackwurstking/pgpress/internal/logger"
)

type Service struct {
	db    *sql.DB
	feeds *feed.Service
}

var _ interfaces.DataOperations[*models.User] = (*Service)(nil)

func New(db *sql.DB, feeds *feed.Service) *Service {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			telegram_id INTEGER NOT NULL,
			user_name TEXT NOT NULL,
			api_key TEXT NOT NULL UNIQUE,
			last_feed TEXT NOT NULL,
			PRIMARY KEY("telegram_id")
		);
	`
	if _, err := db.Exec(query); err != nil {
		panic(dberror.NewDatabaseError("create_table", "users",
			"failed to create users table", err))
	}

	return &Service{
		db:    db,
		feeds: feeds,
	}
}

// List retrieves all users from the database.
func (u *Service) List() ([]*models.User, error) {
	logger.DBUsers().Debug("Starting user list query")
	start := time.Now()

	query := `SELECT * FROM users`
	rows, err := u.db.Query(query)
	if err != nil {
		logger.DBUsers().Error("Failed to execute user list query: %v", err)
		return nil, dberror.NewDatabaseError("select", "users",
			"failed to query users", err)
	}
	defer rows.Close()

	var users []*models.User
	userCount := 0
	for rows.Next() {
		user, err := u.scanUser(rows)
		if err != nil {
			logger.DBUsers().Error("Failed to scan user row %d: %v", userCount, err)
			return nil, dberror.WrapError(err, "failed to scan user")
		}
		users = append(users, user)
		userCount++
	}

	if err := rows.Err(); err != nil {
		logger.DBUsers().Error("Row iteration error after scanning %d users: %v", userCount, err)
		return nil, dberror.NewDatabaseError("select", "users",
			"error iterating over rows", err)
	}

	elapsed := time.Since(start)
	logger.DBUsers().Info("Listed %d users in %v", len(users), elapsed)
	if elapsed > 100*time.Millisecond {
		logger.DBUsers().Warn("Slow user list query took %v for %d users", elapsed, len(users))
	}

	return users, nil
}

// Get retrieves a specific user by Telegram ID.
func (u *Service) Get(telegramID int64) (*models.User, error) {
	logger.DBUsers().Debug("Getting user by Telegram ID: %d", telegramID)
	start := time.Now()

	query := `SELECT * FROM users WHERE telegram_id = ?`
	row := u.db.QueryRow(query, telegramID)

	user, err := u.scanUser(row)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.DBUsers().Debug("User not found for Telegram ID: %d", telegramID)
			return nil, dberror.ErrNotFound
		}
		logger.DBUsers().Error("Failed to get user with Telegram ID %d: %v", telegramID, err)
		return nil, dberror.NewDatabaseError("select", "users",
			fmt.Sprintf("failed to get user with Telegram ID %d", telegramID), err)
	}

	elapsed := time.Since(start)
	logger.DBUsers().Debug("Retrieved user %s (ID: %d) in %v", user.Name, telegramID, elapsed)

	return user, nil
}

// Add creates a new user and generates a corresponding activity feed entry.
func (u *Service) Add(user *models.User, actor *models.User) (int64, error) {
	if user == nil {
		logger.DBUsers().Error("Attempted to add nil user")
		return 0, dberror.NewValidationError("user", "user cannot be nil", nil)
	}

	actorInfo := "system"
	if actor != nil {
		actorInfo = fmt.Sprintf("%s (ID: %d)", actor.Name, actor.TelegramID)
	}
	logger.DBUsers().Info("Adding user %s (Telegram ID: %d) by %s", user.Name, user.TelegramID, actorInfo)
	start := time.Now()

	if err := user.Validate(); err != nil {
		logger.DBUsers().Warn("User validation failed for %s (ID: %d): %v", user.Name, user.TelegramID, err)
		return 0, err
	}

	// Check if user already exists
	var count int
	query := `SELECT COUNT(*) FROM users WHERE telegram_id = ?`
	err := u.db.QueryRow(query, user.TelegramID).Scan(&count)
	if err != nil {
		logger.DBUsers().Error("Failed to check user existence for ID %d: %v", user.TelegramID, err)
		return 0, dberror.NewDatabaseError("select", "users",
			"failed to check user existence", err)
	}

	if count > 0 {
		logger.DBUsers().Warn("User already exists with Telegram ID: %d", user.TelegramID)
		return 0, dberror.ErrAlreadyExists
	}

	// Insert the new user
	query = `INSERT INTO users (telegram_id, user_name, api_key, last_feed) VALUES (?, ?, ?, ?)`
	_, err = u.db.Exec(query, user.TelegramID, user.Name, user.ApiKey, user.LastFeed)
	if err != nil {
		logger.DBUsers().Error("Failed to insert user %s (ID: %d): %v", user.Name, user.TelegramID, err)
		return 0, dberror.NewDatabaseError("insert", "users",
			"failed to insert user", err)
	}

	elapsed := time.Since(start)
	logger.DBUsers().Info("Successfully added user %s (ID: %d) in %v", user.Name, user.TelegramID, elapsed)

	// Create feed entry for the new user
	logger.DBUsers().Debug("Creating feed entry for new user %s", user.Name)
	feed := models.NewFeed(
		"Neuer Benutzer",
		fmt.Sprintf("Benutzer %s wurde hinzugefügt.", user.Name),
		user.TelegramID,
	)
	if err := u.feeds.Add(feed); err != nil {
		logger.DBUsers().Error("Failed to create feed entry for new user %s: %v", user.Name, err)
		return user.TelegramID, dberror.WrapError(err, "failed to add feed entry")
	}

	return user.TelegramID, nil
}

// Delete deletes a user by Telegram ID and generates an activity feed entry.
func (u *Service) Delete(telegramID int64, actor *models.User) error {
	actorInfo := "system"
	if actor != nil {
		actorInfo = fmt.Sprintf("%s (ID: %d)", actor.Name, actor.TelegramID)
	}
	logger.DBUsers().Info("Removing user with Telegram ID %d by %s", telegramID, actorInfo)
	start := time.Now()

	// Get the user before deleting for the feed entry and logging
	user, err := u.Get(telegramID)
	if err != nil && err != dberror.ErrNotFound {
		logger.DBUsers().Error("Failed to get user before deletion (ID: %d): %v", telegramID, err)
	}

	query := `DELETE FROM users WHERE telegram_id = ?`
	result, err := u.db.Exec(query, telegramID)
	if err != nil {
		logger.DBUsers().Error("Failed to delete user with Telegram ID %d: %v", telegramID, err)
		return dberror.NewDatabaseError("delete", "users",
			fmt.Sprintf("failed to delete user with Telegram ID %d", telegramID), err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.DBUsers().Error("Failed to get rows affected for user deletion (ID: %d): %v", telegramID, err)
		return dberror.NewDatabaseError("delete", "users",
			"failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		logger.DBUsers().Warn("No user found to delete with Telegram ID: %d", telegramID)
		return dberror.ErrNotFound
	}

	elapsed := time.Since(start)
	userName := "unknown"
	if user != nil {
		userName = user.Name
	}
	logger.DBUsers().Info("Successfully deleted user %s (ID: %d) in %v", userName, telegramID, elapsed)

	// Create feed entry for the removed user
	if user != nil {
		logger.DBUsers().Debug("Creating feed entry for deleted user %s", user.Name)
		feed := models.NewFeed(
			"Benutzer entfernt",
			fmt.Sprintf("Benutzer %s wurde entfernt.", user.Name),
			user.TelegramID,
		)
		if err := u.feeds.Add(feed); err != nil {
			logger.DBUsers().Error("Failed to create feed entry for deleted user %s: %v", user.Name, err)
			return dberror.WrapError(err, "failed to add feed entry")
		}
	}

	return nil
}

// Update modifies an existing user and generates activity feed entries for changes.
func (u *Service) Update(user, actor *models.User) error {
	telegramID := user.TelegramID
	actorInfo := "system"
	if actor != nil {
		actorInfo = fmt.Sprintf("%s (ID: %d)", actor.Name, actor.TelegramID)
	}
	logger.DBUsers().Info("Updating user %d by %s: new_name=%s", telegramID, actorInfo, user.Name)
	start := time.Now()

	if user == nil {
		logger.DBUsers().Error("Attempted to update nil user")
		return dberror.NewValidationError("user", "user cannot be nil", nil)
	}

	if user.Name == "" {
		logger.DBUsers().Warn("Validation failed: empty username for user ID %d", telegramID)
		return dberror.NewValidationError("user_name", "username cannot be empty", user.Name)
	}

	if user.ApiKey == "" {
		logger.DBUsers().Warn("Validation failed: empty API key for user ID %d", telegramID)
		return dberror.NewValidationError("api_key", "API key cannot be empty", user.ApiKey)
	}

	if len(user.ApiKey) < dbutils.MinAPIKeyLength {
		logger.DBUsers().Warn("Validation failed: API key too short for user ID %d (length=%d, required=%d)",
			telegramID, len(user.ApiKey), dbutils.MinAPIKeyLength)
		return dberror.NewValidationError("api_key",
			fmt.Sprintf("API key must be at least %d characters", dbutils.MinAPIKeyLength),
			len(user.ApiKey))
	}

	// Get the current user for comparison
	prevUser, err := u.Get(telegramID)
	if err != nil {
		logger.DBUsers().Error("Failed to get current user data for update (ID: %d): %v", telegramID, err)
		return err
	}

	// Log what's changing
	changes := []string{}
	if prevUser.Name != user.Name {
		changes = append(changes, fmt.Sprintf("name: '%s' -> '%s'", prevUser.Name, user.Name))
	}
	if prevUser.LastFeed != user.LastFeed {
		changes = append(changes, fmt.Sprintf("last_feed: %d -> %d", prevUser.LastFeed, user.LastFeed))
	}
	if len(changes) > 0 {
		logger.DBUsers().Debug("Changes for user %d: %s", telegramID, strings.Join(changes, ", "))
	}

	// Update the user
	query := `UPDATE users SET user_name = ?, api_key = ?, last_feed = ? WHERE telegram_id = ?`
	_, err = u.db.Exec(query, user.Name, user.ApiKey, user.LastFeed, telegramID)
	if err != nil {
		logger.DBUsers().Error("Failed to update user with Telegram ID %d: %v", telegramID, err)
		return dberror.NewDatabaseError("update", "users",
			fmt.Sprintf("failed to update user with Telegram ID %d", telegramID), err)
	}

	elapsed := time.Since(start)
	logger.DBUsers().Info("Successfully updated user %s (ID: %d) in %v", user.Name, telegramID, elapsed)

	// Create feed entry if username changed
	if prevUser.Name != user.Name {
		logger.DBUsers().Info("Username changed for user %d: '%s' -> '%s'", telegramID, prevUser.Name, user.Name)
		feed := models.NewFeed(
			"Benutzername geändert",
			fmt.Sprintf("Benutzer %s hat den Namen zu %s geändert.", prevUser.Name, user.Name),
			user.TelegramID,
		)

		if err := u.feeds.Add(feed); err != nil {
			logger.DBUsers().Error("Failed to create feed entry for username change: %v", err)
			return dberror.WrapError(err, "failed to add feed entry")
		}
	}

	return nil
}

// GetUserFromApiKey retrieves a user by their API key.
func (s *Service) GetUserFromApiKey(apiKey string) (*models.User, error) {
	logger.DBUsers().Debug("Getting user by API key (length: %d)", len(apiKey))
	start := time.Now()

	if apiKey == "" {
		logger.DBUsers().Warn("Empty API key provided for user lookup")
		return nil, dberror.NewValidationError("api_key", "API key cannot be empty", apiKey)
	}

	query := `SELECT * FROM users WHERE api_key = ?`
	row := s.db.QueryRow(query, apiKey)

	user := &models.User{}
	err := row.Scan(&user.TelegramID, &user.Name, &user.ApiKey, &user.LastFeed)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.DBUsers().Debug("No user found for provided API key")
			return nil, dberror.ErrNotFound
		}
		logger.DBUsers().Error("Failed to get user by API key: %v", err)
		return nil, dberror.NewDatabaseError("select", "users",
			"failed to get user by API key", err)
	}

	elapsed := time.Since(start)
	logger.DBUsers().Debug("Retrieved user %s (ID: %d) by API key in %v", user.Name, user.TelegramID, elapsed)

	return user, nil
}

func (u *Service) scanUser(scanner interfaces.Scannable) (*models.User, error) {
	user := &models.User{}
	err := scanner.Scan(&user.TelegramID, &user.Name, &user.ApiKey, &user.LastFeed)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		logger.DBUsers().Error("Failed to scan user row: %v", err)
		return nil, dberror.NewDatabaseError("scan", "users",
			"failed to scan row", err)
	}
	return user, nil
}
