package services

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/utils"
)

const TableNameCookies = "cookies"

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

	if err := base.CreateTable(query, TableNameCookies); err != nil {
		panic(err)
	}

	return &Cookies{Base: base}
}

func (c *Cookies) List() ([]*models.Cookie, error) {
	slog.Debug("Listing cookies")

	query := fmt.Sprintf(`SELECT * FROM %s ORDER BY last_login DESC`, TableNameCookies)
	rows, err := c.DB.Query(query)
	if err != nil {
		return nil, c.GetSelectError(err)
	}
	defer rows.Close()

	return ScanRows(rows, scanCookie)
}

func (c *Cookies) ListApiKey(apiKey string) ([]*models.Cookie, error) {
	slog.Debug("Listing cookies by API key")

	if err := ValidateAPIKey(apiKey); err != nil {
		return nil, err
	}

	query := fmt.Sprintf(
		`SELECT * FROM %s WHERE api_key = ? ORDER BY last_login DESC`,
		TableNameCookies,
	)
	rows, err := c.DB.Query(query, apiKey)
	if err != nil {
		return nil, c.GetSelectError(err)
	}
	defer rows.Close()

	return ScanRows(rows, scanCookie)
}

func (c *Cookies) Get(value string) (*models.Cookie, error) {
	slog.Debug("Getting cookie by value")

	if value == "" {
		return nil, errors.NewValidationError("value cannot be empty")
	}

	query := fmt.Sprintf(`SELECT * FROM %s WHERE value = ?`, TableNameCookies)
	row := c.DB.QueryRow(query, value)

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
	slog.Debug("Add new cookie")

	if err := cookie.Validate(); err != nil {
		return err
	}

	// Check if cookie already exists
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE value = ?`, TableNameCookies)
	count, err := c.QueryCount(query, cookie.Value)
	if err != nil {
		return c.GetSelectError(err)
	}
	if count > 0 {
		return errors.NewAlreadyExistsError(TableNameCookies)
	}

	query = fmt.Sprintf(
		`INSERT INTO %s (user_agent, value, api_key, last_login) VALUES (?, ?, ?, ?)`,
		TableNameCookies,
	)
	_, err = c.DB.Exec(query, cookie.UserAgent, cookie.Value, cookie.ApiKey, cookie.LastLogin)
	if err != nil {
		return c.GetInsertError(err)
	}

	return nil
}

func (c *Cookies) Update(value string, cookie *models.Cookie) error {
	slog.Debug("Updating a cookie")

	if value == "" {
		return errors.NewValidationError("current_value cannot be empty")
	}

	if err := cookie.Validate(); err != nil {
		return err
	}

	query := fmt.Sprintf(
		`UPDATE %s SET user_agent = ?, value = ?, api_key = ?, last_login = ? WHERE value = ?`,
		TableNameCookies,
	)
	_, err := c.DB.Exec(query, cookie.UserAgent, cookie.Value, cookie.ApiKey, cookie.LastLogin, value)
	if err != nil {
		return c.GetUpdateError(err)
	}

	return nil
}

func (c *Cookies) Remove(value string) error {
	slog.Debug("Removing cookie", "value", utils.MaskString(value))

	if value == "" {
		return errors.NewValidationError("value cannot be empty")
	}

	query := fmt.Sprintf(`DELETE FROM %s WHERE value = ?`, TableNameCookies)
	_, err := c.DB.Exec(query, value)
	if err != nil {
		return c.GetDeleteError(err)
	}

	return nil
}

func (c *Cookies) RemoveApiKey(apiKey string) error {
	slog.Debug("Removing cookies by API key", "api_key", utils.MaskString(apiKey))

	if err := ValidateAPIKey(apiKey); err != nil {
		return err
	}

	query := fmt.Sprintf(`DELETE FROM %s WHERE api_key = ?`, TableNameCookies)
	_, err := c.DB.Exec(query, apiKey)
	if err != nil {
		return c.GetDeleteError(err)
	}

	return nil
}

func (c *Cookies) RemoveExpired(beforeTimestamp int64) error {
	slog.Debug("Removing expired cookies", "before_timestamp", beforeTimestamp)

	query := fmt.Sprintf(`DELETE FROM %s WHERE last_login < ?`, TableNameCookies)
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
		return nil, err
	}
	return cookie, nil
}
