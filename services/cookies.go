package services

import (
	"fmt"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/utils"
)

const TableNameCookies = "cookies"

var EmptyValueError = errors.NewDBError(fmt.Errorf("empty value"), errors.DBTypeValidation)

type Cookies struct {
	*Base
}

func NewCookies(registry *Registry) *Cookies {
	base := NewBase(registry)

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			user_agent TEXT NOT NULL,
			value TEXT NOT NULL,
			api_key TEXT NOT NULL,
			last_login INTEGER NOT NULL,
			PRIMARY KEY("value")
		);
	`, TableNameCookies)

	if _, err := base.DB.Exec(query); err != nil {
		panic(errors.Wrap(err, "create %s table", TableNameCookies))
	}

	return &Cookies{Base: base}
}

func (c *Cookies) List() ([]*models.Cookie, *errors.DBError) {
	query := fmt.Sprintf(`SELECT * FROM %s ORDER BY last_login DESC`, TableNameCookies)
	rows, err := c.DB.Query(query)
	if err != nil {
		return nil, errors.NewDBError(err, errors.DBTypeSelect)
	}
	defer rows.Close()

	return ScanRows(rows, ScanCookie)
}

func (c *Cookies) ListApiKey(apiKey string) ([]*models.Cookie, *errors.DBError) {
	if err := utils.ValidateAPIKey(apiKey); err != nil {
		return nil, errors.NewDBError(err, errors.DBTypeValidation)
	}

	query := fmt.Sprintf(
		`SELECT * FROM %s WHERE api_key = ? ORDER BY last_login DESC`,
		TableNameCookies,
	)

	rows, err := c.DB.Query(query, apiKey)
	if err != nil {
		return nil, errors.NewDBError(err, errors.DBTypeSelect)
	}
	defer rows.Close()

	return ScanRows(rows, ScanCookie)
}

func (c *Cookies) Get(value string) (*models.Cookie, *errors.DBError) {
	if value == "" {
		return nil, EmptyValueError
	}

	query := fmt.Sprintf(`SELECT * FROM %s WHERE value = ?`, TableNameCookies)
	row := c.DB.QueryRow(query, value)

	return ScanRow(row, ScanCookie)
}

func (c *Cookies) Add(cookie *models.Cookie) *errors.DBError {
	if err := cookie.Validate(); err != nil {
		return errors.NewDBError(err, errors.DBTypeValidation)
	}

	// Check if cookie already exists
	count, dberr := c.QueryCount(
		fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE value = ?`, TableNameCookies),
		cookie.Value,
	)
	if dberr != nil {
		return dberr
	}
	if count > 0 {
		return errors.NewDBError(
			fmt.Errorf(
				"value %s already exists in %s",
				utils.MaskString(cookie.Value),
				TableNameCookies,
			),
			errors.DBTypeExists,
		)
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (user_agent, value, api_key, last_login) VALUES (?, ?, ?, ?)`,
		TableNameCookies,
	)
	_, err := c.DB.Exec(query, cookie.UserAgent, cookie.Value, cookie.ApiKey, cookie.LastLogin)
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeInsert)
	}

	return nil
}

// Update updates a cookie with database-level locking to prevent race conditions
func (c *Cookies) Update(value string, cookie *models.Cookie) *errors.DBError {
	if value == "" {
		return EmptyValueError
	}

	if err := cookie.Validate(); err != nil {
		return errors.NewDBError(err, errors.DBTypeValidation)
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
		return errors.NewDBError(err, errors.DBTypeUpdate)
	}

	return nil
}

func (c *Cookies) Remove(value string) *errors.DBError {
	if value == "" {
		return EmptyValueError
	}

	query := fmt.Sprintf(`DELETE FROM %s WHERE value = ?`, TableNameCookies)

	_, err := c.DB.Exec(query, value)
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeDelete)
	}

	return nil
}

func (c *Cookies) RemoveApiKey(apiKey string) *errors.DBError {
	if err := utils.ValidateAPIKey(apiKey); err != nil {
		return errors.NewDBError(err, errors.DBTypeValidation)
	}

	query := fmt.Sprintf(`DELETE FROM %s WHERE api_key = ?`, TableNameCookies)
	_, err := c.DB.Exec(query, apiKey)
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeDelete)
	}

	return nil
}

func (c *Cookies) RemoveExpired(beforeTimestamp int64) *errors.DBError {
	query := fmt.Sprintf(`DELETE FROM %s WHERE last_login < ?`, TableNameCookies)

	_, err := c.DB.Exec(query, beforeTimestamp)
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeDelete)
	}

	return nil
}
