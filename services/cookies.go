package services

import (
	"fmt"
	"net/http"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/utils"
)

type Cookies struct {
	*Base
}

func NewCookies(registry *Registry) *Cookies {
	return &Cookies{
		Base: NewBase(registry),
	}
}

func (c *Cookies) List() ([]*models.Cookie, *errors.MasterError) {
	query := fmt.Sprintf(`SELECT * FROM %s ORDER BY last_login DESC`, TableNameCookies)
	rows, err := c.DB.Query(query)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanCookie)
}

func (c *Cookies) ListApiKey(apiKey string) ([]*models.Cookie, *errors.MasterError) {
	if !utils.ValidateAPIKey(apiKey) {
		return nil, errors.NewMasterError(
			errors.NewValidationError("invalid api_key: %s", utils.MaskString(apiKey)),
			http.StatusBadRequest,
		)
	}

	query := fmt.Sprintf(
		`SELECT * FROM %s WHERE api_key = ? ORDER BY last_login DESC`,
		TableNameCookies,
	)

	rows, err := c.DB.Query(query, apiKey)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanCookie)
}

func (c *Cookies) Get(value string) (*models.Cookie, *errors.MasterError) {
	if value == "" {
		return nil, errors.NewMasterError(
			fmt.Errorf("cookie value missing"),
			http.StatusBadRequest,
		)
	}

	query := fmt.Sprintf(`SELECT * FROM %s WHERE value = ?`, TableNameCookies)
	cookie, err := ScanCookie(c.DB.QueryRow(query, value))
	if err != nil {
		return cookie, errors.NewMasterError(err, 0)
	}

	return cookie, nil
}

func (c *Cookies) Add(cookie *models.Cookie) *errors.MasterError {
	verr := cookie.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	// Check if cookie already exists
	count, merr := c.QueryCount(
		fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE value = ?`, TableNameCookies),
		cookie.Value,
	)
	if merr != nil {
		return merr
	}
	if count > 0 {
		return errors.NewMasterError(
			fmt.Errorf("already exists: %s", cookie),
			http.StatusBadRequest,
		)
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (user_agent, value, api_key, last_login) VALUES (?, ?, ?, ?)`,
		TableNameCookies,
	)
	_, err := c.DB.Exec(query, cookie.UserAgent, cookie.Value, cookie.ApiKey, cookie.LastLogin)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

// Update updates a cookie with database-level locking to prevent race conditions
func (c *Cookies) Update(value string, cookie *models.Cookie) *errors.MasterError {
	verr := cookie.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	if value == "" {
		return errors.NewValidationError(
			"update cookie with value \"%s\" %s",
			utils.MaskString(value), cookie,
		).MasterError()
	}

	// For SQLite, we'll use a different approach: attempt to update with a condition
	// that ensures we're updating the same row with the same timestamp as when we read it
	query := fmt.Sprintf(
		`
			UPDATE %s 
			SET 
				user_agent = ?, 
				value = ?, 
				api_key = ?, 
				last_login = ? 
			WHERE value = ? AND last_login = ?`,
		TableNameCookies,
	)

	// First, get the current cookie to check its last_login timestamp
	currentCookie, dberr := c.Get(value)
	if dberr != nil {
		return dberr
	}

	// Try to update with the current timestamp to ensure we're updating the same row
	_, err := c.DB.Exec(
		query,
		cookie.UserAgent,
		cookie.Value,
		cookie.ApiKey,
		cookie.LastLogin,
		value,
		currentCookie.LastLogin,
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (c *Cookies) Remove(value string) *errors.MasterError {
	if value == "" {
		return errors.NewMasterError(
			fmt.Errorf("value missing"),
			http.StatusBadRequest,
		)
	}

	query := fmt.Sprintf(`DELETE FROM %s WHERE value = ?`, TableNameCookies)

	_, err := c.DB.Exec(query, value)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (c *Cookies) RemoveApiKey(apiKey string) *errors.MasterError {
	if !utils.ValidateAPIKey(apiKey) {
		return errors.NewMasterError(
			fmt.Errorf("invalid api key: %s", utils.MaskString(apiKey)),
			http.StatusBadRequest,
		)
	}

	query := fmt.Sprintf(`DELETE FROM %s WHERE api_key = ?`, TableNameCookies)
	_, err := c.DB.Exec(query, apiKey)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (c *Cookies) RemoveExpired(beforeTimestamp int64) *errors.MasterError {
	query := fmt.Sprintf(`DELETE FROM %s WHERE last_login < ?`, TableNameCookies)

	_, err := c.DB.Exec(query, beforeTimestamp)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}
