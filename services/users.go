package services

import (
	"fmt"
	"net/http"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/utils"
)

type Users struct {
	*Base
}

func NewUsers(r *Registry) *Users {
	return &Users{
		Base: NewBase(r),
	}
}

func (u *Users) Get(telegramID models.TelegramID) (*models.User, *errors.MasterError) {
	query := `SELECT * FROM users WHERE telegram_id = ?`
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

	query := `SELECT * FROM users WHERE api_key = ?`
	row := u.DB.QueryRow(query, apiKey)
	user, err := ScanUser(row)
	if err != nil {
		return user, errors.NewMasterError(err, 0)
	}
	return user, nil
}

func (u *Users) List() ([]*models.User, *errors.MasterError) {
	query := `SELECT * FROM users`
	rows, err := u.DB.Query(query)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanUser)
}

func (u *Users) Add(user *models.User) (models.TelegramID, *errors.MasterError) {
	verr := user.Validate()
	if verr != nil {
		return 0, verr.MasterError()
	}

	// Check if user already exists
	query := `SELECT COUNT(*) FROM users WHERE telegram_id = ?`
	count, dberr := u.QueryCount(query, user.TelegramID)
	if dberr != nil {
		return 0, dberr
	}
	if count > 0 {
		return 0, errors.NewMasterError(fmt.Errorf("User with Telegram ID %d already exists", user.TelegramID), http.StatusBadRequest)
	}

	// Insert the new user
	query = `INSERT INTO users (telegram_id, user_name, api_key, last_feed) VALUES (?, ?, ?, ?)`
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

	query := `DELETE FROM users WHERE telegram_id = ?`
	_, err := u.DB.Exec(query, telegramID)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (u *Users) Update(user *models.User) *errors.MasterError {
	verr := user.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	query := `UPDATE users SET user_name = ?, api_key = ?, last_feed = ? WHERE telegram_id = ?`
	_, err := u.DB.Exec(query, user.Name, user.ApiKey, user.LastFeed, user.TelegramID)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}
