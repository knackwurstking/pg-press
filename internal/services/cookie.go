// Package database provides cookie and session management.
package services

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/interfaces"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

// Service provides database operations for managing authentication cookies and sessions.
type Cookie struct {
	db *sql.DB
}

// NewCookie creates a new Service instance and initializes the database table.
func NewCookie(db *sql.DB) *Cookie {
	query := `
		CREATE TABLE IF NOT EXISTS cookies (
			user_agent TEXT NOT NULL,
			value TEXT NOT NULL,
			api_key TEXT NOT NULL,
			last_login INTEGER NOT NULL,
			PRIMARY KEY("value")
		);
	`
	if _, err := db.Exec(query); err != nil {
		panic(fmt.Errorf("failed to create cookies table: %v", err))
	}

	return &Cookie{db: db}
}

// List retrieves all cookies ordered by last login time (most recent first).
func (c *Cookie) List() ([]*models.Cookie, error) {
	logger.DBCookies().Info("Listing all cookies")

	query := `SELECT * FROM cookies ORDER BY last_login DESC`
	rows, err := c.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("select error: cookies: %v", err)
	}
	defer rows.Close()

	var cookies []*models.Cookie
	for rows.Next() {
		cookie, err := c.scanCookie(rows)
		if err != nil {
			return nil, fmt.Errorf("scan error: cookies: %v", err)
		}
		cookies = append(cookies, cookie)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("select error: cookies: %v", err)
	}

	logger.DBCookies().Debug("Listed %d cookies", len(cookies))
	return cookies, nil
}

// ListApiKey retrieves all cookies associated with a specific API key.
func (c *Cookie) ListApiKey(apiKey string) ([]*models.Cookie, error) {
	logger.DBCookies().Info("Listing cookies by API key")

	if apiKey == "" {
		logger.DBCookies().Debug("Validation failed: empty API key")
		return nil, utils.NewValidationError("api_key: API key cannot be empty")
	}

	query := `SELECT * FROM cookies
		WHERE api_key = ? ORDER BY last_login DESC`
	rows, err := c.db.Query(query, apiKey)
	if err != nil {
		return nil, fmt.Errorf("select error: cookies: %v", err)
	}
	defer rows.Close()

	var cookies []*models.Cookie
	for rows.Next() {
		cookie, err := c.scanCookie(rows)
		if err != nil {
			return nil, fmt.Errorf("scan error: cookies: %v", err)
		}
		cookies = append(cookies, cookie)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("select error: cookies: %v", err)
	}

	logger.DBCookies().Debug("Found %d cookies for API key", len(cookies))
	return cookies, nil
}

// Get retrieves a specific cookie by its value.
func (c *Cookie) Get(value string) (*models.Cookie, error) {
	logger.DBCookies().Debug("Getting cookie by value")

	if value == "" {
		logger.DBCookies().Debug("Validation failed: empty cookie value")
		return nil, utils.NewValidationError("value: cookie value cannot be empty")
	}

	query := `SELECT * FROM cookies WHERE value = ?`
	row := c.db.QueryRow(query, value)
	cookie, err := c.scanCookie(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(value)
		}

		return nil, fmt.Errorf("select error: cookies: %v", err)
	}

	return cookie, nil
}

// Add creates a new cookie session in the database.
func (c *Cookie) Add(cookie *models.Cookie) error {
	logger.DBCookies().Info("Adding cookie: %+v", cookie)

	if cookie == nil {
		logger.DBCookies().Debug("Validation failed: cookie is nil")
		return utils.NewValidationError("cookie: cookie cannot be nil")
	}

	if cookie.Value == "" {
		logger.DBCookies().Debug("Validation failed: empty cookie value")
		return utils.NewValidationError("value: cookie value cannot be empty")
	}

	if cookie.ApiKey == "" {
		logger.DBCookies().Debug("Validation failed: empty API key")
		return utils.NewValidationError("api_key: API key cannot be empty")
	}

	if cookie.LastLogin <= 0 {
		logger.DBCookies().Debug("Validation failed: invalid last login timestamp %d", cookie.LastLogin)
		return utils.NewValidationError("last_login: last login timestamp must be positive")
	}

	var count int
	query := `SELECT COUNT(*) FROM cookies WHERE value = ?`
	err := c.db.QueryRow(query, cookie.Value).Scan(&count)
	if err != nil {
		return fmt.Errorf("select error: cookies: %v", err)
	}
	if count > 0 {
		logger.DBCookies().Debug("Cookie already exists with value")
		return utils.NewAlreadyExistsError("cookie already exists")
	}

	query = `INSERT INTO cookies
		(user_agent, value, api_key, last_login) VALUES (?, ?, ?, ?)`
	_, err = c.db.Exec(query,
		cookie.UserAgent, cookie.Value, cookie.ApiKey, cookie.LastLogin)
	if err != nil {
		return fmt.Errorf("insert error: cookies: %v", err)
	}

	logger.DBCookies().Debug("Successfully added cookie")
	return nil
}

// Update modifies an existing cookie session.
func (c *Cookie) Update(value string, cookie *models.Cookie) error {
	logger.DBCookies().Info("Updating cookie: %+v, value: %s", cookie, value)

	if value == "" {
		logger.DBCookies().Debug("Validation failed: empty current cookie value")
		return utils.NewValidationError("value: current cookie value cannot be empty")
	}
	if cookie == nil {
		logger.DBCookies().Debug("Validation failed: cookie is nil")
		return utils.NewValidationError("cookie: cookie cannot be nil")
	}
	if cookie.Value == "" {
		logger.DBCookies().Debug("Validation failed: empty new cookie value")
		return utils.NewValidationError("value: new cookie value cannot be empty")
	}
	if cookie.ApiKey == "" {
		logger.DBCookies().Debug("Validation failed: empty API key")
		return utils.NewValidationError("api_key: API key cannot be empty")
	}
	if cookie.LastLogin <= 0 {
		logger.DBCookies().Debug("Validation failed: invalid last login timestamp %d", cookie.LastLogin)
		return utils.NewValidationError("last_login: last login timestamp must be positive")
	}

	query := `UPDATE cookies
		SET user_agent = ?, value = ?, api_key = ?, last_login = ? WHERE value = ?`
	result, err := c.db.Exec(query,
		cookie.UserAgent, cookie.Value, cookie.ApiKey, cookie.LastLogin, value)
	if err != nil {
		return fmt.Errorf("update error: cookies: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update error: cookies: %v", err)
	}
	if rowsAffected == 0 {
		logger.DBCookies().Debug("No cookie found to update")
		return utils.NewNotFoundError(value)
	}

	logger.DBCookies().Debug("Successfully updated cookie")
	return nil
}

// Remove deletes a cookie session by its value.
func (c *Cookie) Remove(value string) error {
	logger.DBCookies().Info("Removing cookie, value: %s", value)

	if value == "" {
		logger.DBCookies().Debug("Validation failed: empty cookie value")
		return utils.NewValidationError("value: cookie value cannot be empty")
	}

	query := `DELETE FROM cookies WHERE value = ?`
	_, err := c.db.Exec(query, value)
	if err != nil {
		return fmt.Errorf("delete error: cookies: %v", err)
	}

	logger.DBCookies().Debug("Successfully removed cookie")
	return nil
}

// RemoveApiKey removes all cookie sessions associated with a specific API key.
func (c *Cookie) RemoveApiKey(apiKey string) error {
	logger.DBCookies().Info("Removing cookies by API key")

	if apiKey == "" {
		logger.DBCookies().Debug("Validation failed: empty API key")
		return utils.NewValidationError("api_key: API key cannot be empty")
	}

	query := `DELETE FROM cookies WHERE api_key = ?`
	_, err := c.db.Exec(query, apiKey)
	if err != nil {
		return fmt.Errorf("delete error: cookies: %v", err)
	}

	logger.DBCookies().Debug("Successfully removed cookies for API key")
	return nil
}

// RemoveExpired removes all cookie sessions that are older than the specified timestamp.
func (c *Cookie) RemoveExpired(beforeTimestamp int64) (int64, error) {
	logger.DBCookies().Info("Removing expired cookies, before_timestamp: %d", beforeTimestamp)

	if beforeTimestamp <= 0 {
		logger.DBCookies().Debug("Validation failed: invalid timestamp %d", beforeTimestamp)
		return 0, utils.NewValidationError("timestamp: timestamp must be positive")
	}

	query := `DELETE FROM cookies WHERE last_login < ?`
	result, err := c.db.Exec(query, beforeTimestamp)
	if err != nil {
		return 0, fmt.Errorf("delete error: cookies: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("delete error: cookies: %v", err)
	}

	logger.DBCookies().Debug("Removed %d expired cookies", rowsAffected)
	return rowsAffected, nil
}

func (c *Cookie) scanCookie(scanner interfaces.Scannable) (*models.Cookie, error) {
	cookie := &models.Cookie{}
	err := scanner.Scan(&cookie.UserAgent, &cookie.Value, &cookie.ApiKey, &cookie.LastLogin)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return cookie, nil
}
