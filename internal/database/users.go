// Package database user management.
//
// This file provides database operations for managing users,
// including CRUD operations, authentication via API keys, and integration
// with the activity feed system. All database operations use parameterized
// queries to prevent SQL injection attacks.
package database

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/dberror"
	"github.com/knackwurstking/pgpress/internal/dbutils"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/models"
)

// Users provides database operations for managing user accounts.
// It handles CRUD operations and maintains integration with the activity feed system
// to track user lifecycle events (create, update, delete).
type Users struct {
	db    *sql.DB
	feeds *Feeds
}

var _ DataOperations[*models.User] = (*Users)(nil)

// NewUsers creates a new Users instance and initializes the database table.
func NewUsers(db *sql.DB, feeds *Feeds) *Users {
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

	return &Users{
		db:    db,
		feeds: feeds,
	}
}

// List retrieves all users from the database.
func (u *Users) List() ([]*models.User, error) {
	logger.DBUsers().Info("Listing all users")

	query := `SELECT * FROM users`
	rows, err := u.db.Query(query)
	if err != nil {
		return nil, dberror.NewDatabaseError("select", "users",
			"failed to query users", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user, err := u.scanUser(rows)
		if err != nil {
			return nil, dberror.WrapError(err, "failed to scan user")
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, dberror.NewDatabaseError("select", "users",
			"error iterating over rows", err)
	}

	return users, nil
}

// Get retrieves a specific user by Telegram ID.
func (u *Users) Get(telegramID int64) (*models.User, error) {
	logger.DBUsers().Debug("Getting user by Telegram ID: %d", telegramID)

	query := `SELECT * FROM users WHERE telegram_id = ?`
	row := u.db.QueryRow(query, telegramID)

	user, err := u.scanUser(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, dberror.ErrNotFound
		}
		return nil, dberror.NewDatabaseError("select", "users",
			fmt.Sprintf("failed to get user with Telegram ID %d", telegramID), err)
	}

	return user, nil
}

// Add creates a new user and generates a corresponding activity feed entry.
func (u *Users) Add(user *models.User, actor *models.User) (int64, error) {
	if user == nil {
		return 0, dberror.NewValidationError("user", "user cannot be nil", nil)
	}

	logger.DBUsers().Info("Adding user: %d, %s", user.TelegramID, user.UserName)

	if err := user.Validate(); err != nil {
		return 0, err
	}

	// Check if user already exists
	var count int
	query := `SELECT COUNT(*) FROM users
		WHERE telegram_id = ?`
	err := u.db.QueryRow(query, user.TelegramID).Scan(&count)
	if err != nil {
		return 0, dberror.NewDatabaseError("select", "users",
			"failed to check user existence", err)
	}

	if count > 0 {
		return 0, dberror.ErrAlreadyExists
	}

	// Insert the new user
	query = `INSERT INTO users
		(telegram_id, user_name, api_key, last_feed) VALUES (?, ?, ?, ?)`
	_, err = u.db.Exec(query,
		user.TelegramID, user.UserName, user.ApiKey, user.LastFeed)
	if err != nil {
		return 0, dberror.NewDatabaseError("insert", "users",
			"failed to insert user", err)
	}

	// Create feed entry for the new user
	feed := models.NewFeed(
		models.FeedTypeUserAdd,
		models.FeedUserAdd{ID: user.TelegramID, Name: user.UserName},
	)
	if err := u.feeds.Add(feed); err != nil {
		return user.TelegramID, dberror.WrapError(err, "failed to add feed entry")
	}

	return user.TelegramID, nil
}

// Delete deletes a user by Telegram ID and generates an activity feed entry.
func (u *Users) Delete(telegramID int64, actor *models.User) error {
	logger.DBUsers().Info("Removing user: %d", telegramID)

	// Get the user before deleting for the feed entry
	user, _ := u.Get(telegramID)

	query := `DELETE FROM users WHERE telegram_id = ?`
	result, err := u.db.Exec(query, telegramID)
	if err != nil {
		return dberror.NewDatabaseError("delete", "users",
			fmt.Sprintf("failed to delete user with Telegram ID %d", telegramID), err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return dberror.NewDatabaseError("delete", "users",
			"failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return dberror.ErrNotFound
	}

	// Create feed entry for the removed user
	if user != nil {
		feed := models.NewFeed(
			models.FeedTypeUserRemove,
			models.FeedUserRemove{ID: user.TelegramID, Name: user.UserName},
		)
		if err := u.feeds.Add(feed); err != nil {
			return dberror.WrapError(err, "failed to add feed entry")
		}
	}

	return nil
}

// Update modifies an existing user and generates activity feed entries for changes.
func (u *Users) Update(user, actor *models.User) error {
	telegramID := user.TelegramID
	logger.DBUsers().Info("Updating user: telegram_id=%d, new_name=%s", telegramID, user.UserName)

	if user == nil {
		return dberror.NewValidationError("user", "user cannot be nil", nil)
	}

	if user.UserName == "" {
		logger.DBUsers().Debug("Validation failed: empty username")
		return dberror.NewValidationError("user_name", "username cannot be empty", user.UserName)
	}

	if user.ApiKey == "" {
		logger.DBUsers().Debug("Validation failed: empty API key")
		return dberror.NewValidationError("api_key", "API key cannot be empty", user.ApiKey)
	}

	if len(user.ApiKey) < dbutils.MinAPIKeyLength {
		logger.DBUsers().Debug("Validation failed: API key too short (length=%d, required=%d)", len(user.ApiKey), dbutils.MinAPIKeyLength)
		return dberror.NewValidationError("api_key",
			fmt.Sprintf("API key must be at least %d characters", dbutils.MinAPIKeyLength),
			len(user.ApiKey))
	}

	// Get the current user for comparison
	prevUser, err := u.Get(telegramID)
	if err != nil {
		return err
	}

	// Update the user
	query := `UPDATE users
		SET user_name = ?, api_key = ?, last_feed = ? WHERE telegram_id = ?`
	_, err = u.db.Exec(query,
		user.UserName, user.ApiKey, user.LastFeed, telegramID)
	if err != nil {
		return dberror.NewDatabaseError("update", "users",
			fmt.Sprintf("failed to update user with Telegram ID %d", telegramID), err)
	}

	// Create feed entry if username changed
	if prevUser.UserName != user.UserName {
		logger.DBUsers().Debug("Username changed from '%s' to '%s'", prevUser.UserName, user.UserName)
		feed := models.NewFeed(
			models.FeedTypeUserNameChange,
			&models.FeedUserNameChange{
				ID:  user.TelegramID,
				Old: prevUser.UserName,
				New: user.UserName,
			},
		)

		if err := u.feeds.Add(feed); err != nil {
			return dberror.WrapError(err, "failed to add feed entry")
		}
	}

	return nil
}

func (u *Users) scanUser(scanner scannable) (*models.User, error) {
	user := &models.User{}
	err := scanner.Scan(&user.TelegramID, &user.UserName, &user.ApiKey, &user.LastFeed)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, dberror.NewDatabaseError("scan", "users",
			"failed to scan row", err)
	}
	return user, nil
}
