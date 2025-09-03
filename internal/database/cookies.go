// Package database provides cookie and session management.
package database

import (
	"database/sql"
	"slices"

	"github.com/knackwurstking/pgpress/internal/dberror"
	"github.com/knackwurstking/pgpress/internal/logger"
)

// Cookies provides database operations for managing authentication cookies and sessions.
type Cookies struct {
	db *sql.DB
}

// NewCookies creates a new Cookies instance and initializes the database table.
func NewCookies(db *sql.DB) *Cookies {
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
		panic(dberror.NewDatabaseError("create_table", "cookies",
			"failed to create cookies table", err))
	}

	return &Cookies{db: db}
}

// List retrieves all cookies ordered by last login time (most recent first).
func (c *Cookies) List() ([]*Cookie, error) {
	logger.DBCookies().Info("Listing all cookies")

	query := `SELECT * FROM cookies ORDER BY last_login DESC`
	rows, err := c.db.Query(query)
	if err != nil {
		return nil, dberror.NewDatabaseError("select", "cookies", "failed to query cookies", err)
	}
	defer rows.Close()

	var cookies []*Cookie
	for rows.Next() {
		cookie, err := c.scanCookie(rows)
		if err != nil {
			return nil, dberror.WrapError(err, "failed to scan cookie")
		}
		cookies = append(cookies, cookie)
	}

	if err := rows.Err(); err != nil {
		return nil, dberror.NewDatabaseError("select", "cookies", "error iterating over rows", err)
	}

	logger.DBCookies().Debug("Listed %d cookies", len(cookies))
	return cookies, nil
}

// ListApiKey retrieves all cookies associated with a specific API key.
func (c *Cookies) ListApiKey(apiKey string) ([]*Cookie, error) {
	logger.DBCookies().Info("Listing cookies by API key")

	if apiKey == "" {
		logger.DBCookies().Debug("Validation failed: empty API key")
		return nil, dberror.NewValidationError("api_key", "API key cannot be empty", apiKey)
	}

	query := `SELECT * FROM cookies
		WHERE api_key = ? ORDER BY last_login DESC`
	rows, err := c.db.Query(query, apiKey)
	if err != nil {
		return nil, dberror.NewDatabaseError("select", "cookies", "failed to query cookies by API key", err)
	}
	defer rows.Close()

	var cookies []*Cookie
	for rows.Next() {
		cookie, err := c.scanCookie(rows)
		if err != nil {
			return nil, dberror.WrapError(err, "failed to scan cookie")
		}
		cookies = append(cookies, cookie)
	}

	if err := rows.Err(); err != nil {
		return nil, dberror.NewDatabaseError("select", "cookies", "error iterating over rows", err)
	}

	logger.DBCookies().Debug("Found %d cookies for API key", len(cookies))
	return cookies, nil
}

// Get retrieves a specific cookie by its value.
func (c *Cookies) Get(value string) (*Cookie, error) {
	logger.DBCookies().Debug("Getting cookie by value")

	if value == "" {
		logger.DBCookies().Debug("Validation failed: empty cookie value")
		return nil, dberror.NewValidationError("value", "cookie value cannot be empty", value)
	}

	query := `SELECT * FROM cookies WHERE value = ?`
	row := c.db.QueryRow(query, value)
	cookie, err := c.scanCookie(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, dberror.ErrNotFound
		}
		return nil, dberror.NewDatabaseError("select", "cookies", "failed to get cookie by value", err)
	}

	return cookie, nil
}

// Add creates a new cookie session in the database.
func (c *Cookies) Add(cookie *Cookie) error {
	logger.DBCookies().Info("Adding cookie: %+v", cookie)

	if cookie == nil {
		logger.DBCookies().Debug("Validation failed: cookie is nil")
		return dberror.NewValidationError("cookie", "cookie cannot be nil", nil)
	}
	if cookie.Value == "" {
		logger.DBCookies().Debug("Validation failed: empty cookie value")
		return dberror.NewValidationError("value", "cookie value cannot be empty", cookie.Value)
	}
	if cookie.ApiKey == "" {
		logger.DBCookies().Debug("Validation failed: empty API key")
		return dberror.NewValidationError("api_key", "API key cannot be empty", cookie.ApiKey)
	}
	if cookie.LastLogin <= 0 {
		logger.DBCookies().Debug("Validation failed: invalid last login timestamp %d", cookie.LastLogin)
		return dberror.NewValidationError("last_login",
			"last login timestamp must be positive", cookie.LastLogin)
	}

	var count int
	query := `SELECT COUNT(*) FROM cookies WHERE value = ?`
	err := c.db.QueryRow(query, cookie.Value).Scan(&count)
	if err != nil {
		return dberror.NewDatabaseError("select", "cookies", "failed to check cookie existence", err)
	}
	if count > 0 {
		logger.DBCookies().Debug("Cookie already exists with value")
		return dberror.ErrAlreadyExists
	}

	query = `INSERT INTO cookies
		(user_agent, value, api_key, last_login) VALUES (?, ?, ?, ?)`
	_, err = c.db.Exec(query,
		cookie.UserAgent, cookie.Value, cookie.ApiKey, cookie.LastLogin)
	if err != nil {
		return dberror.NewDatabaseError("insert", "cookies", "failed to insert cookie", err)
	}

	logger.DBCookies().Debug("Successfully added cookie")
	return nil
}

// Update modifies an existing cookie session.
func (c *Cookies) Update(value string, cookie *Cookie) error {
	logger.DBCookies().Info("Updating cookie: %+v, value: %s", cookie, value)

	if value == "" {
		logger.DBCookies().Debug("Validation failed: empty current cookie value")
		return dberror.NewValidationError("value", "current cookie value cannot be empty", value)
	}
	if cookie == nil {
		logger.DBCookies().Debug("Validation failed: cookie is nil")
		return dberror.NewValidationError("cookie", "cookie cannot be nil", nil)
	}
	if cookie.Value == "" {
		logger.DBCookies().Debug("Validation failed: empty new cookie value")
		return dberror.NewValidationError("value", "new cookie value cannot be empty", cookie.Value)
	}
	if cookie.ApiKey == "" {
		logger.DBCookies().Debug("Validation failed: empty API key")
		return dberror.NewValidationError("api_key", "API key cannot be empty", cookie.ApiKey)
	}
	if cookie.LastLogin <= 0 {
		logger.DBCookies().Debug("Validation failed: invalid last login timestamp %d", cookie.LastLogin)
		return dberror.NewValidationError("last_login",
			"last login timestamp must be positive", cookie.LastLogin)
	}

	query := `UPDATE cookies
		SET user_agent = ?, value = ?, api_key = ?, last_login = ? WHERE value = ?`
	result, err := c.db.Exec(query,
		cookie.UserAgent, cookie.Value, cookie.ApiKey, cookie.LastLogin, value)
	if err != nil {
		return dberror.NewDatabaseError("update", "cookies", "failed to update cookie", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return dberror.NewDatabaseError("update", "cookies", "failed to get rows affected", err)
	}
	if rowsAffected == 0 {
		logger.DBCookies().Debug("No cookie found to update")
		return dberror.ErrNotFound
	}

	logger.DBCookies().Debug("Successfully updated cookie")
	return nil
}

// Remove deletes a cookie session by its value.
func (c *Cookies) Remove(value string) error {
	logger.DBCookies().Info("Removing cookie, value: %s", value)

	if value == "" {
		logger.DBCookies().Debug("Validation failed: empty cookie value")
		return dberror.NewValidationError("value", "cookie value cannot be empty", value)
	}

	query := `DELETE FROM cookies WHERE value = ?`
	_, err := c.db.Exec(query, value)
	if err != nil {
		return dberror.NewDatabaseError("delete", "cookies", "failed to delete cookie", err)
	}

	logger.DBCookies().Debug("Successfully removed cookie")
	return nil
}

// RemoveApiKey removes all cookie sessions associated with a specific API key.
func (c *Cookies) RemoveApiKey(apiKey string) error {
	logger.DBCookies().Info("Removing cookies by API key")

	if apiKey == "" {
		logger.DBCookies().Debug("Validation failed: empty API key")
		return dberror.NewValidationError("api_key", "API key cannot be empty", apiKey)
	}

	query := `DELETE FROM cookies WHERE api_key = ?`
	_, err := c.db.Exec(query, apiKey)
	if err != nil {
		return dberror.NewDatabaseError("delete", "cookies", "failed to delete cookies by API key", err)
	}

	logger.DBCookies().Debug("Successfully removed cookies for API key")
	return nil
}

// RemoveExpired removes all cookie sessions that are older than the specified timestamp.
func (c *Cookies) RemoveExpired(beforeTimestamp int64) (int64, error) {
	logger.DBCookies().Info("Removing expired cookies, before_timestamp: %d", beforeTimestamp)

	if beforeTimestamp <= 0 {
		logger.DBCookies().Debug("Validation failed: invalid timestamp %d", beforeTimestamp)
		return 0, dberror.NewValidationError("timestamp", "timestamp must be positive", beforeTimestamp)
	}

	query := `DELETE FROM cookies WHERE last_login < ?`
	result, err := c.db.Exec(query, beforeTimestamp)
	if err != nil {
		return 0, dberror.NewDatabaseError("delete", "cookies", "failed to delete expired cookies", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, dberror.NewDatabaseError("delete", "cookies", "failed to get rows affected", err)
	}

	logger.DBCookies().Debug("Removed %d expired cookies", rowsAffected)
	return rowsAffected, nil
}

func (c *Cookies) scanCookie(scanner scannable) (*Cookie, error) {
	cookie := &Cookie{}
	err := scanner.Scan(&cookie.UserAgent, &cookie.Value, &cookie.ApiKey, &cookie.LastLogin)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, dberror.NewDatabaseError("scan", "cookies", "failed to scan row", err)
	}
	return cookie, nil
}

// SortCookies sorts a slice of cookies by last login time in descending order.
func SortCookies(cookies []*Cookie) []*Cookie {
	if len(cookies) <= 1 {
		return cookies
	}

	cookiesSorted := make([]*Cookie, 0, len(cookies))

outer:
	for _, cookie := range cookies {
		for i, sortedCookie := range cookiesSorted {
			if cookie.LastLogin > sortedCookie.LastLogin {
				cookiesSorted = slices.Insert(cookiesSorted, i, cookie)
				continue outer
			}
		}
		cookiesSorted = append(cookiesSorted, cookie)
	}

	return cookiesSorted
}
