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

	"github.com/knackwurstking/pgpress/internal/interfaces"
	"github.com/knackwurstking/pgpress/pkg/constants"
	"github.com/knackwurstking/pgpress/pkg/logger"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

type Users struct {
	db  *sql.DB
	log *logger.Logger
}

func NewUsers(db *sql.DB) *Users {
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
		panic(fmt.Errorf("failed to create users table: %v", err))
	}

	return &Users{
		db:  db,
		log: logger.GetComponentLogger("Service: Users"),
	}
}

// List retrieves all users from the database.
func (u *Users) List() ([]*models.User, error) {
	start := time.Now()

	query := `SELECT * FROM users`
	rows, err := u.db.Query(query)
	if err != nil {
		u.log.Error("Failed to execute user list query: %v", err)
		return nil, fmt.Errorf("select error: users: %v", err)
	}
	defer rows.Close()

	var users []*models.User
	userCount := 0
	for rows.Next() {
		user, err := u.scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("scan error: users: %v", err)
		}
		users = append(users, user)
		userCount++
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("select error: users: %v", err)
	}

	elapsed := time.Since(start)
	if elapsed > 100*time.Millisecond {
		u.log.Warn("Slow user list query took %v for %d users", elapsed, len(users))
	}

	return users, nil
}

// Get retrieves a specific user by Telegram ID.
func (u *Users) Get(telegramID int64) (*models.User, error) {
	query := `SELECT * FROM users WHERE telegram_id = ?`
	row := u.db.QueryRow(query, telegramID)

	user, err := u.scanUser(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("user with Telegram ID %d not found", telegramID))
		}

		return nil, fmt.Errorf("select error: users: %v", err)
	}

	return user, nil
}

// Add creates a new user and generates a corresponding activity feed entry.
func (u *Users) Add(user *models.User) (int64, error) {
	if user == nil {
		u.log.Error("Attempted to add nil user")
		return 0, utils.NewValidationError("user: user cannot be nil")
	}

	u.log.Info("Adding user %s (Telegram ID: %d)", user.Name, user.TelegramID)

	if err := user.Validate(); err != nil {
		return 0, err
	}

	// Check if user already exists
	var count int
	query := `SELECT COUNT(*) FROM users WHERE telegram_id = ?`
	err := u.db.QueryRow(query, user.TelegramID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("select error: users: %v", err)
	}

	if count > 0 {
		return 0, utils.NewAlreadyExistsError(fmt.Sprintf("User with Telegram ID %d already exists", user.TelegramID))
	}

	// Insert the new user
	query = `INSERT INTO users (telegram_id, user_name, api_key, last_feed) VALUES (?, ?, ?, ?)`
	_, err = u.db.Exec(query, user.TelegramID, user.Name, user.ApiKey, user.LastFeed)
	if err != nil {
		return 0, fmt.Errorf("insert error: users: %v", err)
	}

	return user.TelegramID, nil
}

// Delete deletes a user by Telegram ID and generates an activity feed entry.
func (u *Users) Delete(telegramID int64) error {
	u.log.Info("Removing user %d", telegramID)

	// Get the user before deleting for the feed entry and logging
	if _, err := u.Get(telegramID); utils.IsNotFoundError(err) {
		return err
	} else if err != nil {
		u.log.Error("Failed to get user before deletion (ID: %d): %v", telegramID, err)
	}

	query := `DELETE FROM users WHERE telegram_id = ?`
	result, err := u.db.Exec(query, telegramID)
	if err != nil {
		return fmt.Errorf("delete error: users: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete error: users: %v", err)
	}

	if rowsAffected == 0 {
		return utils.NewNotFoundError(
			fmt.Sprintf("user with Telegram ID %d not found", telegramID),
		)
	}

	return nil
}

// Update modifies an existing user and generates activity feed entries for changes.
func (u *Users) Update(user *models.User) error {
	if user == nil {
		return utils.NewValidationError("user: user cannot be nil")
	}

	telegramID := user.TelegramID
	u.log.Info("Updating user %d: new_name=%s", telegramID, user.Name)

	if user.Name == "" {
		return utils.NewValidationError("user_name: username cannot be empty")
	}

	if user.ApiKey == "" {
		return utils.NewValidationError("api_key: API key cannot be empty")
	}

	if len(user.ApiKey) < constants.MinAPIKeyLength {
		return utils.NewValidationError(
			fmt.Sprintf(
				"api_key: API key must be at least %d characters",
				constants.MinAPIKeyLength,
			),
		)
	}

	// Update the user
	query := `UPDATE users SET user_name = ?, api_key = ?, last_feed = ? WHERE telegram_id = ?`
	_, err := u.db.Exec(query, user.Name, user.ApiKey, user.LastFeed, telegramID)
	if err != nil {
		return fmt.Errorf("update error: users: %v", err)
	}

	return nil
}

// GetUserFromApiKey retrieves a user by their API key.
func (s *Users) GetUserFromApiKey(apiKey string) (*models.User, error) {
	start := time.Now()

	if apiKey == "" {
		return nil, utils.NewValidationError("api_key: API key cannot be empty")
	}

	query := `SELECT * FROM users WHERE api_key = ?`
	row := s.db.QueryRow(query, apiKey)

	user := &models.User{}
	err := row.Scan(&user.TelegramID, &user.Name, &user.ApiKey, &user.LastFeed)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError("apiKey: " + utils.MaskString(apiKey))
		}

		return nil, fmt.Errorf("select error: users: %v", err)
	}

	elapsed := time.Since(start)
	if elapsed > 50*time.Millisecond {
		s.log.Warn("Slow API key lookup took %v", elapsed)
	}

	return user, nil
}

func (u *Users) scanUser(scanner interfaces.Scannable) (*models.User, error) {
	user := &models.User{}
	err := scanner.Scan(&user.TelegramID, &user.Name, &user.ApiKey, &user.LastFeed)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to scan user: %v", err)
	}
	return user, nil
}
