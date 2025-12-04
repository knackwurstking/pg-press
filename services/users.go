package services

import (
	"fmt"

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

	if _, err := base.DB.Exec(query); err != nil {
		panic(errors.Wrap(err, "create %s table", TableNameUsers))
	}

	return &Users{Base: base}
}

func (u *Users) Get(telegramID models.TelegramID) (*models.User, *errors.DBError) {
	query := fmt.Sprintf(`SELECT * FROM %s WHERE telegram_id = ?`, TableNameUsers)
	row := u.DB.QueryRow(query, telegramID)

	return ScanRow(row, ScanUser)
}

func (u *Users) GetUserFromApiKey(apiKey string) (*models.User, *errors.DBError) {
	err := utils.ValidateAPIKey(apiKey)
	if err != nil {
		return nil, errors.NewDBError(err, errors.DBTypeValidation)
	}

	query := fmt.Sprintf(`SELECT * FROM %s WHERE api_key = ?`, TableNameUsers)
	row := u.DB.QueryRow(query, apiKey)

	return ScanRow(row, ScanUser)
}

func (u *Users) List() ([]*models.User, *errors.DBError) {
	query := fmt.Sprintf(`SELECT * FROM %s`, TableNameUsers)
	rows, err := u.DB.Query(query)
	if err != nil {
		return nil, errors.NewDBError(err, errors.DBTypeSelect)
	}
	defer rows.Close()

	return ScanRows(rows, ScanUser)
}

func (u *Users) Add(user *models.User) (models.TelegramID, *errors.DBError) {
	if err := user.Validate(); err != nil {
		return 0, errors.NewDBError(err, errors.DBTypeValidation)
	}

	// Check if user already exists
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE telegram_id = ?`, TableNameUsers)
	count, dberr := u.QueryCount(query, user.TelegramID)
	if dberr != nil {
		return 0, dberr
	}
	if count > 0 {
		return 0, errors.NewDBError(
			fmt.Errorf("User with Telegram ID %d already exists", user.TelegramID),
			errors.DBTypeExists,
		)
	}

	// Insert the new user
	query = fmt.Sprintf(`INSERT INTO %s (telegram_id, user_name, api_key, last_feed) VALUES (?, ?, ?, ?)`, TableNameUsers)
	_, err := u.DB.Exec(query, user.TelegramID, user.Name, user.ApiKey, user.LastFeed)
	if err != nil {
		return 0, errors.NewDBError(err, errors.DBTypeInsert)
	}

	return user.TelegramID, nil
}

func (u *Users) Delete(telegramID models.TelegramID) *errors.DBError {
	_, dberr := u.Get(telegramID)
	if dberr != nil {
		return dberr
	}

	query := fmt.Sprintf(`DELETE FROM %s WHERE telegram_id = ?`, TableNameUsers)
	_, err := u.DB.Exec(query, telegramID)
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeDelete)
	}

	return nil
}

func (u *Users) Update(user *models.User) *errors.DBError {
	err := user.Validate()
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeValidation)
	}

	query := fmt.Sprintf(`UPDATE %s SET user_name = ?, api_key = ?, last_feed = ? WHERE telegram_id = ?`, TableNameUsers)
	_, err = u.DB.Exec(query, user.Name, user.ApiKey, user.LastFeed, user.TelegramID)
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeUpdate)
	}

	return nil
}
