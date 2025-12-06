package services

import (
	"fmt"
	"net/http"

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

func (u *Users) Get(telegramID models.TelegramID) (*models.User, *errors.MasterError) {
	query := fmt.Sprintf(`SELECT * FROM %s WHERE telegram_id = ?`, TableNameUsers)
	row := u.DB.QueryRow(query, telegramID)
	user, err := ScanUser(row)
	if err != nil {
		return user, errors.NewMasterError(err, 0)
	}
	return user, nil
}

func (u *Users) GetUserFromApiKey(apiKey string) (*models.User, *errors.MasterError) {
	if !utils.ValidateAPIKey(apiKey) {
		return nil, errors.NewMasterError(fmt.Errorf("invalid api key: %s", utils.MaskString(apiKey)), http.StatusBadRequest)
	}

	query := fmt.Sprintf(`SELECT * FROM %s WHERE api_key = ?`, TableNameUsers)
	row := u.DB.QueryRow(query, apiKey)
	user, err := ScanUser(row)
	if err != nil {
		return user, errors.NewMasterError(err, 0)
	}
	return user, nil
}

func (u *Users) List() ([]*models.User, *errors.MasterError) {
	query := fmt.Sprintf(`SELECT * FROM %s`, TableNameUsers)
	rows, err := u.DB.Query(query)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanUser)
}

func (u *Users) Add(user *models.User) (models.TelegramID, *errors.MasterError) {
	if !user.Validate() {
		return 0, errors.NewMasterError(fmt.Errorf("invalid user: %s", user), http.StatusBadRequest)
	}

	// Check if user already exists
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE telegram_id = ?`, TableNameUsers)
	count, dberr := u.QueryCount(query, user.TelegramID)
	if dberr != nil {
		return 0, dberr
	}
	if count > 0 {
		return 0, errors.NewMasterError(fmt.Errorf("User with Telegram ID %d already exists", user.TelegramID), http.StatusBadRequest)
	}

	// Insert the new user
	query = fmt.Sprintf(`INSERT INTO %s (telegram_id, user_name, api_key, last_feed) VALUES (?, ?, ?, ?)`, TableNameUsers)
	_, err := u.DB.Exec(query, user.TelegramID, user.Name, user.ApiKey, user.LastFeed)
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
	}

	return user.TelegramID, nil
}

func (u *Users) Delete(telegramID models.TelegramID) *errors.MasterError {
	_, dberr := u.Get(telegramID)
	if dberr != nil {
		return dberr
	}

	query := fmt.Sprintf(`DELETE FROM %s WHERE telegram_id = ?`, TableNameUsers)
	_, err := u.DB.Exec(query, telegramID)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (u *Users) Update(user *models.User) *errors.MasterError {
	if !user.Validate() {
		return errors.NewMasterError(fmt.Errorf("invalid user: %s", user), http.StatusBadRequest)
	}

	query := fmt.Sprintf(`UPDATE %s SET user_name = ?, api_key = ?, last_feed = ? WHERE telegram_id = ?`, TableNameUsers)
	_, err := u.DB.Exec(query, user.Name, user.ApiKey, user.LastFeed, user.TelegramID)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}
