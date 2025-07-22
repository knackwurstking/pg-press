// Package pgvis user management.
//
// This file provides database operations for managing users,
// including CRUD operations, authentication via API keys, and integration
// with the activity feed system. All database operations use parameterized
// queries to prevent SQL injection attacks.
package pgvis

import (
	"database/sql"
	"fmt"

	"github.com/charmbracelet/log"
)

// SQL queries for user operations
const (
	createUsersTableQuery = `
		CREATE TABLE IF NOT EXISTS users (
			telegram_id INTEGER NOT NULL,
			user_name TEXT NOT NULL,
			api_key TEXT NOT NULL UNIQUE,
			last_feed TEXT NOT NULL,
			PRIMARY KEY("telegram_id")
		);
	`

	selectAllUsersQuery         = `SELECT * FROM users`
	selectUserByTelegramIDQuery = `SELECT * FROM users WHERE telegram_id = ?`
	selectUserByAPIKeyQuery     = `SELECT * FROM users WHERE api_key = ?`
	selectUserExistsQuery       = `SELECT COUNT(*) FROM users WHERE telegram_id = ? OR user_name = ?`
	insertUserQuery             = `INSERT INTO users (telegram_id, user_name, api_key, last_feed) VALUES (?, ?, ?, ?)`
	updateUserQuery             = `UPDATE users SET user_name = ?, api_key = ?, last_feed = ? WHERE telegram_id = ?`
	deleteUserQuery             = `DELETE FROM users WHERE telegram_id = ?`
)

// Users provides database operations for managing user accounts.
// It handles CRUD operations and maintains integration with the activity feed system
// to track user lifecycle events (create, update, delete).
type Users struct {
	db    *sql.DB
	feeds *Feeds
}

// NewUsers creates a new Users instance and initializes the database table.
func NewUsers(db *sql.DB, feeds *Feeds) *Users {
	if _, err := db.Exec(createUsersTableQuery); err != nil {
		panic(NewDatabaseError("create_table", "users",
			"failed to create users table", err))
	}

	return &Users{
		db:    db,
		feeds: feeds,
	}
}

// List retrieves all users from the database.
func (u *Users) List() ([]*User, error) {
	log.Debug("Listing users")

	rows, err := u.db.Query(selectAllUsersQuery)
	if err != nil {
		return nil, NewDatabaseError("select", "users",
			"failed to query users", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user, err := u.scanUser(rows)
		if err != nil {
			return nil, WrapError(err, "failed to scan user")
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, NewDatabaseError("select", "users",
			"error iterating over rows", err)
	}

	return users, nil
}

// Get retrieves a specific user by Telegram ID.
func (u *Users) Get(telegramID int64) (*User, error) {
	log.Debug("Getting user by Telegram ID", telegramID)

	row := u.db.QueryRow(selectUserByTelegramIDQuery, telegramID)

	user, err := u.scanUserRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, NewDatabaseError("select", "users",
			fmt.Sprintf("failed to get user with Telegram ID %d", telegramID), err)
	}

	return user, nil
}

// GetUserFromApiKey retrieves a user by their API key.
func (u *Users) GetUserFromApiKey(apiKey string) (*User, error) {
	log.Debug("Getting user by API key")

	if apiKey == "" {
		return nil, NewValidationError("api_key", "API key cannot be empty", apiKey)
	}

	row := u.db.QueryRow(selectUserByAPIKeyQuery, apiKey)

	user, err := u.scanUserRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, NewDatabaseError("select", "users",
			"failed to get user by API key", err)
	}

	return user, nil
}

// Add creates a new user and generates a corresponding activity feed entry.
func (u *Users) Add(user *User) error {
	log.Debug("Adding user", user.TelegramID, user.UserName)

	if user == nil {
		return NewValidationError("user", "user cannot be nil", nil)
	}

	if err := user.Validate(); err != nil {
		return err
	}

	// Check if user already exists
	var count int
	err := u.db.QueryRow(selectUserExistsQuery, user.TelegramID, user.UserName).Scan(&count)
	if err != nil {
		return NewDatabaseError("select", "users",
			"failed to check user existence", err)
	}

	if count > 0 {
		return ErrAlreadyExists
	}

	// Insert the new user
	_, err = u.db.Exec(insertUserQuery,
		user.TelegramID, user.UserName, user.ApiKey, user.LastFeed)
	if err != nil {
		return NewDatabaseError("insert", "users",
			"failed to insert user", err)
	}

	// Create feed entry for the new user
	feed := NewFeed(
		FeedTypeUserAdd,
		FeedUserAdd{user.TelegramID, user.UserName},
	)
	if err := u.feeds.Add(feed); err != nil {
		return WrapError(err, "failed to add feed entry")
	}

	return nil
}

// Remove deletes a user by Telegram ID and generates an activity feed entry.
func (u *Users) Remove(telegramID int64) error {
	log.Debug("Removing user", telegramID)

	// Get the user before deleting for the feed entry
	user, _ := u.Get(telegramID)

	result, err := u.db.Exec(deleteUserQuery, telegramID)
	if err != nil {
		return NewDatabaseError("delete", "users",
			fmt.Sprintf("failed to delete user with Telegram ID %d", telegramID), err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return NewDatabaseError("delete", "users",
			"failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	// Create feed entry for the removed user
	if user != nil {
		feed := NewFeed(
			FeedTypeUserRemove,
			FeedUserRemove{user.TelegramID, user.UserName},
		)
		if err := u.feeds.Add(feed); err != nil {
			return WrapError(err, "failed to add feed entry")
		}
	}

	return nil
}

// Update modifies an existing user and generates activity feed entries for changes.
func (u *Users) Update(telegramID int64, user *User) error {
	log.Debug("Update user", user, "telegramID", telegramID)

	if user == nil {
		return NewValidationError("user", "user cannot be nil", nil)
	}

	if user.UserName == "" {
		return NewValidationError("user_name", "username cannot be empty", user.UserName)
	}

	if user.ApiKey == "" {
		return NewValidationError("api_key", "API key cannot be empty", user.ApiKey)
	}

	if len(user.ApiKey) < MinAPIKeyLength {
		return NewValidationError("api_key",
			fmt.Sprintf("API key must be at least %d characters", MinAPIKeyLength),
			len(user.ApiKey))
	}

	// Get the current user for comparison
	prevUser, err := u.Get(telegramID)
	if err != nil {
		return err
	}

	// Update the user
	_, err = u.db.Exec(updateUserQuery,
		user.UserName, user.ApiKey, user.LastFeed, telegramID)
	if err != nil {
		return NewDatabaseError("update", "users",
			fmt.Sprintf("failed to update user with Telegram ID %d", telegramID), err)
	}

	// Create feed entry if username changed
	if prevUser.UserName != user.UserName {
		feed := NewFeed(
			FeedTypeUserNameChange,
			&FeedUserNameChange{
				ID:  user.TelegramID,
				Old: prevUser.UserName,
				New: user.UserName,
			},
		)

		if err := u.feeds.Add(feed); err != nil {
			return WrapError(err, "failed to add feed entry")
		}
	}

	return nil
}

// scanUser scans a user from database rows.
func (u *Users) scanUser(rows *sql.Rows) (*User, error) {
	user := &User{}
	err := rows.Scan(&user.TelegramID, &user.UserName, &user.ApiKey, &user.LastFeed)
	if err != nil {
		return nil, NewDatabaseError("scan", "users",
			"failed to scan row", err)
	}
	return user, nil
}

// scanUserRow scans a user from a single database row.
func (u *Users) scanUserRow(row *sql.Row) (*User, error) {
	user := &User{}
	err := row.Scan(&user.TelegramID, &user.UserName, &user.ApiKey, &user.LastFeed)
	if err != nil {
		return nil, err
	}
	return user, nil
}
