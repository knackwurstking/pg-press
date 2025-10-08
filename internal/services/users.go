// Package database user management.
//
// This file provides database operations for managing users,
// including CRUD operations, authentication via API keys, and integration
// with the activity feed system. All database operations use parameterized
// queries to prevent SQL injection attacks.
package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

type Users struct {
	*BaseService
}

func NewUsers(db *sql.DB) *Users {
	base := NewBaseService(db, "Users")

	query := `
		CREATE TABLE IF NOT EXISTS users (
			telegram_id INTEGER NOT NULL,
			user_name TEXT NOT NULL,
			api_key TEXT NOT NULL UNIQUE,
			last_feed TEXT NOT NULL,
			PRIMARY KEY("telegram_id")
		);
	`

	if err := base.CreateTable(query, "users"); err != nil {
		panic(err)
	}

	return &Users{
		BaseService: base,
	}
}

// List retrieves all users from the database.
func (u *Users) List() ([]*models.User, error) {
	start := time.Now()
	u.LogOperation("Listing users")

	query := `SELECT * FROM users`
	rows, err := u.db.Query(query)
	if err != nil {
		return nil, u.HandleSelectError(err, "users")
	}
	defer rows.Close()

	users, err := ScanUsersFromRows(rows)
	if err != nil {
		return nil, err
	}

	u.LogSlowQuery(start, "user list", 100*time.Millisecond, fmt.Sprintf("%d users", len(users)))

	return users, nil
}

// Get retrieves a specific user by Telegram ID.
func (u *Users) Get(telegramID int64) (*models.User, error) {
	u.LogOperation("Getting user", telegramID)

	query := `SELECT * FROM users WHERE telegram_id = ?`
	row := u.db.QueryRow(query, telegramID)

	user, err := ScanSingleRow(row, ScanUser, "users")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("user with Telegram ID %d not found", telegramID))
		}
		return nil, err
	}

	return user, nil
}

// Add creates a new user and generates a corresponding activity feed entry.
func (u *Users) Add(user *models.User) (int64, error) {
	if err := ValidateUser(user); err != nil {
		return 0, err
	}

	u.log.Info("Adding user %s (Telegram ID: %d)", user.Name, user.TelegramID)

	// Check if user already exists
	exists, err := u.CheckExistence(`SELECT COUNT(*) FROM users WHERE telegram_id = ?`, user.TelegramID)
	if err != nil {
		return 0, u.HandleSelectError(err, "users")
	}

	if exists {
		return 0, utils.NewAlreadyExistsError(fmt.Sprintf("User with Telegram ID %d already exists", user.TelegramID))
	}

	// Insert the new user
	query := `INSERT INTO users (telegram_id, user_name, api_key, last_feed) VALUES (?, ?, ?, ?)`
	_, err = u.db.Exec(query, user.TelegramID, user.Name, user.ApiKey, user.LastFeed)
	if err != nil {
		return 0, u.HandleInsertError(err, "users")
	}

	return user.TelegramID, nil
}

// Delete deletes a user by Telegram ID and generates an activity feed entry.
func (u *Users) Delete(telegramID int64) error {
	u.log.Info("Removing user %d", telegramID)

	// Get the user before deleting for validation
	if _, err := u.Get(telegramID); utils.IsNotFoundError(err) {
		return err
	} else if err != nil {
		u.log.Error("Failed to get user before deletion (ID: %d): %v", telegramID, err)
	}

	query := `DELETE FROM users WHERE telegram_id = ?`
	result, err := u.db.Exec(query, telegramID)
	if err != nil {
		return u.HandleDeleteError(err, "users")
	}

	return u.CheckRowsAffected(result, "user", telegramID)
}

// Update modifies an existing user and generates activity feed entries for changes.
func (u *Users) Update(user *models.User) error {
	if err := ValidateUser(user); err != nil {
		return err
	}

	telegramID := user.TelegramID
	u.log.Info("Updating user %d: new_name=%s", telegramID, user.Name)

	// Update the user
	query := `UPDATE users SET user_name = ?, api_key = ?, last_feed = ? WHERE telegram_id = ?`
	result, err := u.db.Exec(query, user.Name, user.ApiKey, user.LastFeed, telegramID)
	if err != nil {
		return u.HandleUpdateError(err, "users")
	}

	return u.CheckRowsAffected(result, "user", telegramID)
}

// GetUserFromApiKey retrieves a user by their API key.
func (u *Users) GetUserFromApiKey(apiKey string) (*models.User, error) {
	start := time.Now()

	if err := ValidateAPIKey(apiKey); err != nil {
		return nil, err
	}

	u.LogOperation("Getting user by API key")

	query := `SELECT * FROM users WHERE api_key = ?`
	row := u.db.QueryRow(query, apiKey)

	user, err := ScanSingleRow(row, ScanUser, "users")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError("apiKey: " + utils.MaskString(apiKey))
		}
		return nil, err
	}

	u.LogSlowQuery(start, "API key lookup", 50*time.Millisecond)

	return user, nil
}
