package services

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/utils"
)

const TableNameUsers = "users"

type Users struct {
	*Base
}

func NewUsers(r *Registry) *Users {
	base := NewBase(r)

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			telegram_id INTEGER NOT NULL,
			user_name TEXT NOT NULL,
			api_key TEXT NOT NULL UNIQUE,
			last_feed TEXT NOT NULL,
			PRIMARY KEY("telegram_id")
		);
	`, TableNameUsers)

	if err := base.CreateTable(query, TableNameUsers); err != nil {
		panic(err)
	}

	return &Users{Base: base}
}

func (u *Users) List() ([]*models.User, error) {
	slog.Debug("Listing users")

	query := fmt.Sprintf(`SELECT * FROM %s`, TableNameUsers)
	rows, err := u.DB.Query(query)
	if err != nil {
		return nil, u.GetSelectError(err)
	}
	defer rows.Close()

	users, err := ScanRows(rows, scanUser)
	if err != nil {
		return nil, fmt.Errorf("failed to scan users: %v", err)
	}

	return users, nil
}

func (u *Users) Get(telegramID models.TelegramID) (*models.User, error) {
	slog.Debug("Getting user", "telegram_id", telegramID)

	query := fmt.Sprintf(`SELECT * FROM %s WHERE telegram_id = ?`, TableNameUsers)
	row := u.DB.QueryRow(query, telegramID)

	user, err := ScanSingleRow(row, scanUser)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError(
				fmt.Sprintf("user with Telegram ID %d not found", telegramID),
			)
		}
		return nil, u.GetSelectError(err)
	}

	return user, nil
}

func (u *Users) Add(user *models.User) (models.TelegramID, error) {
	slog.Debug("Adding user", "user_name", user.Name, "telegram_id", user.TelegramID)

	if err := user.Validate(); err != nil {
		return 0, err
	}

	// Check if user already exists
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE telegram_id = ?`, TableNameUsers)
	count, err := u.QueryCount(query, user.TelegramID)
	if err != nil {
		return 0, u.GetSelectError(err)
	}
	if count > 0 {
		return 0, errors.NewAlreadyExistsError(
			fmt.Sprintf("User with Telegram ID %d already exists", user.TelegramID),
		)
	}

	// Insert the new user
	query = fmt.Sprintf(`INSERT INTO %s (telegram_id, user_name, api_key, last_feed) VALUES (?, ?, ?, ?)`, TableNameUsers)
	_, err = u.DB.Exec(query, user.TelegramID, user.Name, user.ApiKey, user.LastFeed)
	if err != nil {
		return 0, u.GetInsertError(err)
	}

	return user.TelegramID, nil
}

func (u *Users) Delete(telegramID models.TelegramID) error {
	slog.Debug("Removing user", "telegram_id", telegramID)

	if _, err := u.Get(telegramID); err != nil {
		if errors.IsNotFoundError(err) {
			return err
		}
		slog.Error("Failed to get user before deletion", "telegram_id", telegramID, "error", err)
	}

	query := fmt.Sprintf(`DELETE FROM %s WHERE telegram_id = ?`, TableNameUsers)
	_, err := u.DB.Exec(query, telegramID)
	if err != nil {
		return u.GetDeleteError(err)
	}

	return nil
}

func (u *Users) Update(user *models.User) error {
	slog.Debug("Updating user", "telegram_id", user.TelegramID, "user", user)

	if err := user.Validate(); err != nil {
		return err
	}

	query := fmt.Sprintf(`UPDATE %s SET user_name = ?, api_key = ?, last_feed = ? WHERE telegram_id = ?`, TableNameUsers)
	_, err := u.DB.Exec(query, user.Name, user.ApiKey, user.LastFeed, user.TelegramID)
	if err != nil {
		return u.GetUpdateError(err)
	}

	return nil
}

func (u *Users) GetUserFromApiKey(apiKey string) (*models.User, error) {
	slog.Debug("Getting user by API key", "api_key", utils.MaskString(apiKey))

	if err := ValidateAPIKey(apiKey); err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`SELECT * FROM %s WHERE api_key = ?`, TableNameUsers)
	row := u.DB.QueryRow(query, apiKey)

	user, err := ScanSingleRow(row, scanUser)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("apiKey: " + utils.MaskString(apiKey))
		}
		return nil, fmt.Errorf("failed to get user from API key: %v", err)
	}

	return user, nil
}

func scanUser(scanner Scannable) (*models.User, error) {
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
