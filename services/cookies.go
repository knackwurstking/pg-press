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

	if _, err := base.DB.Exec(query); err != nil {
		panic(errors.Wrap(err, "create %s table", TableNameCookies))
	}

	return &Cookies{Base: base}
}

func (c *Cookies) List() ([]*models.Cookie, error) {
	slog.Info("Listing cookies")

	query := fmt.Sprintf(`SELECT * FROM %s ORDER BY last_login DESC`, TableNameCookies)
	rows, err := c.DB.Query(query)
	if err != nil {
		return nil, c.GetSelectError(err)
	}
	defer rows.Close()

	return ScanRows(rows, scanCookie)
}

func (c *Cookies) ListApiKey(apiKey string) ([]*models.Cookie, error) {
	slog.Info("Listing cookies by API key")

	if err := utils.ValidateAPIKey(apiKey); err != nil {
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
	slog.Info("Getting cookie by value")

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
	slog.Info("Add new cookie")

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

// Update updates a cookie with database-level locking to prevent race conditions
func (c *Cookies) Update(value string, cookie *models.Cookie) error {
	slog.Info("Updating a cookie")

	if value == "" {
		return errors.NewValidationError("current_value cannot be empty")
	}

	if err := cookie.Validate(); err != nil {
		return err
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
	currentCookie, err := c.Get(value)
	if err != nil {
		return err
	}

	// Try to update with the current timestamp to ensure we're updating the same row
	result, err := c.DB.Exec(
		query,
		cookie.UserAgent,
		cookie.Value,
		cookie.ApiKey,
		cookie.LastLogin,
		value,
		currentCookie.LastLogin,
	)
	if err != nil {
		return c.GetUpdateError(err)
	}

	// Check if any rows were affected (if not, it means another goroutine updated it first)
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		// Another goroutine updated the cookie first, so we need to retry or handle appropriately
		// For now, we'll return an error to indicate the race condition occurred
		return errors.NewValidationError("cookie was updated by another process")
	}

	return nil
}

func (c *Cookies) Remove(value string) error {
	slog.Info("Removing cookie", "value", utils.MaskString(value))

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
	slog.Info("Removing cookies by API key", "api_key", utils.MaskString(apiKey))

	if err := utils.ValidateAPIKey(apiKey); err != nil {
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
	slog.Info("Removing expired cookies", "before_timestamp", beforeTimestamp)

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
