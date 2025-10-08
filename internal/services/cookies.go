// Package database provides cookie and session management.
package services

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

// Cookies provides database operations for managing authentication cookies and sessions.
type Cookies struct {
	*BaseService
}

// NewCookies creates a new Service instance and initializes the database table.
func NewCookies(db *sql.DB) *Cookies {
	base := NewBaseService(db, "Cookies")

	query := `
		CREATE TABLE IF NOT EXISTS cookies (
			user_agent TEXT NOT NULL,
			value TEXT NOT NULL,
			api_key TEXT NOT NULL,
			last_login INTEGER NOT NULL,
			PRIMARY KEY("value")
		);
	`

	if err := base.CreateTable(query, "cookies"); err != nil {
		panic(err)
	}

	return &Cookies{
		BaseService: base,
	}
}

// List retrieves all cookies ordered by last login time (most recent first).
func (c *Cookies) List() ([]*models.Cookie, error) {
	c.LogOperation("Listing cookies")

	query := `SELECT * FROM cookies ORDER BY last_login DESC`
	rows, err := c.db.Query(query)
	if err != nil {
		return nil, c.HandleSelectError(err, "cookies")
	}
	defer rows.Close()

	cookies, err := ScanCookiesFromRows(rows)
	if err != nil {
		return nil, err
	}

	c.LogOperation("Listed cookies", fmt.Sprintf("count: %d", len(cookies)))
	return cookies, nil
}

// ListApiKey retrieves all cookies associated with a specific API key.
func (c *Cookies) ListApiKey(apiKey string) ([]*models.Cookie, error) {
	if err := ValidateAPIKey(apiKey); err != nil {
		return nil, err
	}

	c.LogOperation("Listing cookies by API key")

	query := `SELECT * FROM cookies WHERE api_key = ? ORDER BY last_login DESC`
	rows, err := c.db.Query(query, apiKey)
	if err != nil {
		return nil, c.HandleSelectError(err, "cookies")
	}
	defer rows.Close()

	cookies, err := ScanCookiesFromRows(rows)
	if err != nil {
		return nil, err
	}

	c.LogOperation("Found cookies for API key", fmt.Sprintf("count: %d", len(cookies)))
	return cookies, nil
}

// Get retrieves a specific cookie by its value.
func (c *Cookies) Get(value string) (*models.Cookie, error) {
	if err := ValidateNotEmpty(value, "value"); err != nil {
		return nil, err
	}

	c.LogOperation("Getting cookie by value")

	query := `SELECT * FROM cookies WHERE value = ?`
	row := c.db.QueryRow(query, value)

	cookie, err := ScanSingleRow(row, ScanCookie, "cookies")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(value)
		}
		return nil, err
	}

	return cookie, nil
}

// Add creates a new cookie session in the database.
func (c *Cookies) Add(cookie *models.Cookie) error {
	if err := ValidateCookie(cookie); err != nil {
		return err
	}

	c.LogOperation("Adding cookie")

	// Check if cookie already exists
	exists, err := c.CheckExistence(`SELECT COUNT(*) FROM cookies WHERE value = ?`, cookie.Value)
	if err != nil {
		return c.HandleSelectError(err, "cookies")
	}

	if exists {
		return utils.NewAlreadyExistsError("cookie already exists")
	}

	query := `INSERT INTO cookies (user_agent, value, api_key, last_login) VALUES (?, ?, ?, ?)`
	_, err = c.db.Exec(query, cookie.UserAgent, cookie.Value, cookie.ApiKey, cookie.LastLogin)
	if err != nil {
		return c.HandleInsertError(err, "cookies")
	}

	c.LogOperation("Successfully added cookie")
	return nil
}

// Update modifies an existing cookie session.
func (c *Cookies) Update(value string, cookie *models.Cookie) error {
	if err := ValidateNotEmpty(value, "current_value"); err != nil {
		return err
	}

	if err := ValidateCookie(cookie); err != nil {
		return err
	}

	c.LogOperation("Updating cookie")

	query := `UPDATE cookies SET user_agent = ?, value = ?, api_key = ?, last_login = ? WHERE value = ?`
	result, err := c.db.Exec(query, cookie.UserAgent, cookie.Value, cookie.ApiKey, cookie.LastLogin, value)
	if err != nil {
		return c.HandleUpdateError(err, "cookies")
	}

	if err := c.CheckRowsAffected(result, "cookie", value); err != nil {
		return err
	}

	c.LogOperation("Successfully updated cookie")
	return nil
}

// Remove deletes a cookie session by its value.
func (c *Cookies) Remove(value string) error {
	if err := ValidateNotEmpty(value, "value"); err != nil {
		return err
	}

	c.LogOperation("Removing cookie")

	query := `DELETE FROM cookies WHERE value = ?`
	result, err := c.db.Exec(query, value)
	if err != nil {
		return c.HandleDeleteError(err, "cookies")
	}

	if err := c.CheckRowsAffected(result, "cookie", value); err != nil {
		return err
	}

	c.LogOperation("Successfully removed cookie")
	return nil
}

// RemoveApiKey removes all cookie sessions associated with a specific API key.
func (c *Cookies) RemoveApiKey(apiKey string) error {
	if err := ValidateAPIKey(apiKey); err != nil {
		return err
	}

	c.log.Info("Removing cookies by API key")

	query := `DELETE FROM cookies WHERE api_key = ?`
	result, err := c.db.Exec(query, apiKey)
	if err != nil {
		return c.HandleDeleteError(err, "cookies")
	}

	rowsAffected, err := c.GetRowsAffected(result, "remove cookies by API key")
	if err != nil {
		return err
	}

	c.LogOperation("Successfully removed cookies for API key", fmt.Sprintf("count: %d", rowsAffected))
	return nil
}

// RemoveExpired removes all cookie sessions that are older than the specified timestamp.
func (c *Cookies) RemoveExpired(beforeTimestamp int64) (int64, error) {
	if err := ValidateTimestamp(beforeTimestamp, "timestamp"); err != nil {
		return 0, err
	}

	c.log.Info("Removing expired cookies, before_timestamp: %d", beforeTimestamp)

	query := `DELETE FROM cookies WHERE last_login < ?`
	result, err := c.db.Exec(query, beforeTimestamp)
	if err != nil {
		return 0, c.HandleDeleteError(err, "cookies")
	}

	rowsAffected, err := c.GetRowsAffected(result, "remove expired cookies")
	if err != nil {
		return 0, err
	}

	c.LogOperation("Removed expired cookies", fmt.Sprintf("count: %d", rowsAffected))
	return rowsAffected, nil
}
