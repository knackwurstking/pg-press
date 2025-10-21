package users

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/services/base"
	"github.com/knackwurstking/pgpress/internal/services/shared/scanner"
	"github.com/knackwurstking/pgpress/internal/services/shared/validation"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

type Service struct {
	*base.BaseService
}

func NewService(db *sql.DB) *Service {
	base := base.NewBaseService(db, "Users")

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

	return &Service{
		BaseService: base,
	}
}

func (u *Service) List() ([]*models.User, error) {
	u.Log.Debug("Listing users")

	query := `SELECT * FROM users`
	rows, err := u.DB.Query(query)
	if err != nil {
		return nil, u.HandleSelectError(err, "users")
	}
	defer rows.Close()

	users, err := scanUsersFromRows(rows)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (u *Service) Get(telegramID int64) (*models.User, error) {
	u.Log.Debug("Getting user: Telegram ID: %d", telegramID)

	query := `SELECT * FROM users WHERE telegram_id = ?`
	row := u.DB.QueryRow(query, telegramID)

	user, err := scanner.ScanSingleRow(row, scanUser, "users")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("user with Telegram ID %d not found", telegramID))
		}
		return nil, err
	}

	return user, nil
}

func (u *Service) Add(user *models.User) (int64, error) {
	if err := validateUser(user); err != nil {
		return 0, err
	}

	u.Log.Debug("Adding user %s (Telegram ID: %d)", user.Name, user.TelegramID)

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
	_, err = u.DB.Exec(query, user.TelegramID, user.Name, user.ApiKey, user.LastFeed)
	if err != nil {
		return 0, u.HandleInsertError(err, "users")
	}

	return user.TelegramID, nil
}

func (u *Service) Delete(telegramID int64) error {
	u.Log.Debug("Removing user %d", telegramID)

	// Get the user before deleting for validation
	if _, err := u.Get(telegramID); utils.IsNotFoundError(err) {
		return err
	} else if err != nil {
		u.Log.Error("Failed to get user before deletion (ID: %d): %v", telegramID, err)
	}

	query := `DELETE FROM users WHERE telegram_id = ?`
	result, err := u.DB.Exec(query, telegramID)
	if err != nil {
		return u.HandleDeleteError(err, "users")
	}

	return u.CheckRowsAffected(result, "user", telegramID)
}

func (u *Service) Update(user *models.User) error {
	if err := validateUser(user); err != nil {
		return err
	}

	telegramID := user.TelegramID
	u.Log.Debug("Updating user %d: new_name=%s", telegramID, user.Name)

	// Update the user
	query := `UPDATE users SET user_name = ?, api_key = ?, last_feed = ? WHERE telegram_id = ?`
	result, err := u.DB.Exec(query, user.Name, user.ApiKey, user.LastFeed, telegramID)
	if err != nil {
		return u.HandleUpdateError(err, "users")
	}

	return u.CheckRowsAffected(result, "user", telegramID)
}

func (u *Service) GetUserFromApiKey(apiKey string) (*models.User, error) {
	if err := validation.ValidateAPIKey(apiKey); err != nil {
		return nil, err
	}

	u.Log.Debug("Getting user by API key")

	query := `SELECT * FROM users WHERE api_key = ?`
	row := u.DB.QueryRow(query, apiKey)

	user, err := scanner.ScanSingleRow(row, scanUser, "users")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError("apiKey: " + utils.MaskString(apiKey))
		}
		return nil, err
	}

	return user, nil
}
