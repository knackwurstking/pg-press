package services

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/errors"
	"github.com/knackwurstking/pgpress/logger"
	"github.com/knackwurstking/pgpress/models"
	"github.com/knackwurstking/pgpress/utils"
)

const TableNameCookies = "cookies"

type Cookies struct {
	*Base
}

func NewCookies(registry *Registry) *Cookies {
	base := NewBase(registry, logger.NewComponentLogger("Service: Cookies"))

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			user_agent TEXT NOT NULL,
			value TEXT NOT NULL,
			api_key TEXT NOT NULL,
			last_login INTEGER NOT NULL,
			PRIMARY KEY("value")
		);
	`, TableNameCookies)

	if err := base.CreateTable(query, TableNameCookies); err != nil {
		panic(err)
	}

	return &Cookies{
		Base: base,
	}
}

func (c *Cookies) List() ([]*models.Cookie, error) {
	c.Log.Debug("Listing cookies")

	query := `SELECT * FROM cookies ORDER BY last_login DESC`
	rows, err := c.DB.Query(query)
	if err != nil {
		return nil, c.GetSelectError(err)
	}
	defer rows.Close()

	cookies, err := ScanRows(rows, scanCookie)
	if err != nil {
		return nil, fmt.Errorf("failed to scan cookies: %v", err)
	}

	return cookies, nil
}

func (c *Cookies) ListApiKey(apiKey string) ([]*models.Cookie, error) {
	c.Log.Debug("Listing cookies by API key")

	if err := ValidateAPIKey(apiKey); err != nil {
		return nil, err
	}

	query := `SELECT * FROM cookies WHERE api_key = ? ORDER BY last_login DESC`
	rows, err := c.DB.Query(query, apiKey)
	if err != nil {
		return nil, c.GetSelectError(err)
	}
	defer rows.Close()

	cookies, err := ScanRows(rows, scanCookie)
	if err != nil {
		return nil, fmt.Errorf("failed to scan cookies: %v", err)
	}

	return cookies, nil
}

func (c *Cookies) Get(value string) (*models.Cookie, error) {
	c.Log.Debug("Getting cookie by value")

	if value == "" {
		return nil, errors.NewValidationError("value cannot be empty")
	}

	row := c.DB.QueryRow(`SELECT * FROM cookies WHERE value = ?`, value)
	cookie, err := ScanSingleRow(row, scanCookie)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError(value)
		}
		return nil, c.GetSelectError(err)
	}

	return cookie, nil
}

func (c *Cookies) Add(cookie *models.Cookie) error {
	c.Log.Debug("Add new cookie")

	if err := cookie.Validate(); err != nil {
		return err
	}

	// Check if cookie already exists
	count, err := c.QueryCount(`SELECT COUNT(*) FROM cookies WHERE value = ?`, cookie.Value)
	if err != nil {
		return c.GetSelectError(err)
	}
	if count > 0 {
		return errors.NewAlreadyExistsError(TableNameCookies)
	}

	query := `INSERT INTO cookies (user_agent, value, api_key, last_login) VALUES (?, ?, ?, ?)`
	_, err = c.DB.Exec(query, cookie.UserAgent, cookie.Value, cookie.ApiKey, cookie.LastLogin)
	if err != nil {
		return c.GetInsertError(err)
	}

	return nil
}

func (c *Cookies) Update(value string, cookie *models.Cookie) error {
	c.Log.Debug("Updating a cookie")

	if value == "" {
		return errors.NewValidationError("current_value cannot be empty")
	}

	if err := cookie.Validate(); err != nil {
		return err
	}

	query := `UPDATE cookies SET user_agent = ?, value = ?, api_key = ?, last_login = ? WHERE value = ?`
	_, err := c.DB.Exec(query, cookie.UserAgent, cookie.Value, cookie.ApiKey, cookie.LastLogin, value)
	if err != nil {
		return c.GetUpdateError(err)
	}

	return nil
}

func (c *Cookies) Remove(value string) error {
	c.Log.Debug("Removing cookie with value: %s", utils.MaskString(value))

	if value == "" {
		return errors.NewValidationError("value cannot be empty")
	}

	query := `DELETE FROM cookies WHERE value = ?`
	_, err := c.DB.Exec(query, value)
	if err != nil {
		return c.GetDeleteError(err)
	}

	return nil
}

func (c *Cookies) RemoveApiKey(apiKey string) error {
	c.Log.Debug("Removing cookies by API key: %s", utils.MaskString(apiKey))

	if err := ValidateAPIKey(apiKey); err != nil {
		return err
	}

	query := `DELETE FROM cookies WHERE api_key = ?`
	_, err := c.DB.Exec(query, apiKey)
	if err != nil {
		return c.GetDeleteError(err)
	}

	return nil
}

func (c *Cookies) RemoveExpired(beforeTimestamp int64) error {
	c.Log.Debug("Removing expired cookies, before_timestamp: %d", beforeTimestamp)

	query := `DELETE FROM cookies WHERE last_login < ?`
	_, err := c.DB.Exec(query, beforeTimestamp)
	if err != nil {
		return c.GetDeleteError(err)
	}

	return nil
}

func scanCookie(scanner Scannable) (*models.Cookie, error) {
	cookie := &models.Cookie{}
	err := scanner.Scan(&cookie.UserAgent, &cookie.Value, &cookie.ApiKey, &cookie.LastLogin)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to scan cookie: %v", err)
	}
	return cookie, nil
}
