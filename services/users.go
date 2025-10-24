package services

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/errors"
	"github.com/knackwurstking/pgpress/logger"
	"github.com/knackwurstking/pgpress/models"
	"github.com/knackwurstking/pgpress/utils"
)

const TableNameUsers = "users"

type Users struct {
	*Base
}

func NewUsers(r *Registry) *Users {
	base := NewBase(r, logger.NewComponentLogger("Service: Users"))

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
	u.Log.Debug("Listing users")

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

func (u *Users) Get(telegramID int64) (*models.User, error) {
	u.Log.Debug("Getting user: Telegram ID: %d", telegramID)

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

func (u *Users) Add(user *models.User) (int64, error) {
	u.Log.Debug("Adding user %s (Telegram ID: %d)", user.Name, user.TelegramID)

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

func (u *Users) Delete(telegramID int64) error {
	u.Log.Debug("Removing user %d", telegramID)

	if _, err := u.Get(telegramID); err != nil {
		if errors.IsNotFoundError(err) {
			return err
		}
		u.Log.Error("Failed to get user before deletion (ID: %d): %v", telegramID, err)
	}

	query := fmt.Sprintf(`DELETE FROM %s WHERE telegram_id = ?`, TableNameUsers)
	_, err := u.DB.Exec(query, telegramID)
	if err != nil {
		return u.GetDeleteError(err)
	}

	return nil
}

func (u *Users) Update(user *models.User) error {
	u.Log.Debug("Updating user %d: user=%#v", user.TelegramID, user)

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
	u.Log.Debug("Getting user by API key: %s", utils.MaskString(apiKey))

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
