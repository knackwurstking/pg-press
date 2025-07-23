// Package pgvis provides cookie and session management.
package pgvis

import (
	"database/sql"
	"slices"

	"github.com/knackwurstking/pg-vis/pgvis/logger"
)

const (
	createCookiesTableQuery = `
		CREATE TABLE IF NOT EXISTS cookies (
			user_agent TEXT NOT NULL,
			value TEXT NOT NULL,
			api_key TEXT NOT NULL,
			last_login INTEGER NOT NULL,
			PRIMARY KEY("value")
		);
	`

	selectAllCookiesQuery        = `SELECT * FROM cookies ORDER BY last_login DESC`
	selectCookiesByAPIKeyQuery   = `SELECT * FROM cookies WHERE api_key = ? ORDER BY last_login DESC`
	selectCookieByValueQuery     = `SELECT * FROM cookies WHERE value = ?`
	selectCookieExistsQuery      = `SELECT COUNT(*) FROM cookies WHERE value = ?`
	insertCookieQuery            = `INSERT INTO cookies (user_agent, value, api_key, last_login) VALUES (?, ?, ?, ?)`
	updateCookieQuery            = `UPDATE cookies SET user_agent = ?, value = ?, api_key = ?, last_login = ? WHERE value = ?`
	deleteCookieByValueQuery     = `DELETE FROM cookies WHERE value = ?`
	deleteCookiesByAPIKeyQuery   = `DELETE FROM cookies WHERE api_key = ?`
	deleteCookiesBeforeTimeQuery = `DELETE FROM cookies WHERE last_login < ?`
)

// Cookies provides database operations for managing authentication cookies and sessions.
type Cookies struct {
	db *sql.DB
}

// NewCookies creates a new Cookies instance and initializes the database table.
func NewCookies(db *sql.DB) *Cookies {
	if _, err := db.Exec(createCookiesTableQuery); err != nil {
		panic(NewDatabaseError("create_table", "cookies",
			"failed to create cookies table", err))
	}

	return &Cookies{db: db}
}

// List retrieves all cookies ordered by last login time (most recent first).
func (c *Cookies) List() ([]*Cookie, error) {
	logger.Cookie().Info("Listing all cookies")

	rows, err := c.db.Query(selectAllCookiesQuery)
	if err != nil {
		return nil, NewDatabaseError("select", "cookies", "failed to query cookies", err)
	}
	defer rows.Close()

	var cookies []*Cookie
	for rows.Next() {
		cookie, err := c.scanCookie(rows)
		if err != nil {
			return nil, WrapError(err, "failed to scan cookie")
		}
		cookies = append(cookies, cookie)
	}

	if err := rows.Err(); err != nil {
		return nil, NewDatabaseError("select", "cookies", "error iterating over rows", err)
	}

	return cookies, nil
}

// ListApiKey retrieves all cookies associated with a specific API key.
func (c *Cookies) ListApiKey(apiKey string) ([]*Cookie, error) {
	logger.Cookie().Info("Listing cookies by API key")

	if apiKey == "" {
		return nil, NewValidationError("api_key", "API key cannot be empty", apiKey)
	}

	rows, err := c.db.Query(selectCookiesByAPIKeyQuery, apiKey)
	if err != nil {
		return nil, NewDatabaseError("select", "cookies", "failed to query cookies by API key", err)
	}
	defer rows.Close()

	var cookies []*Cookie
	for rows.Next() {
		cookie, err := c.scanCookie(rows)
		if err != nil {
			return nil, WrapError(err, "failed to scan cookie")
		}
		cookies = append(cookies, cookie)
	}

	if err := rows.Err(); err != nil {
		return nil, NewDatabaseError("select", "cookies", "error iterating over rows", err)
	}

	return cookies, nil
}

// Get retrieves a specific cookie by its value.
func (c *Cookies) Get(value string) (*Cookie, error) {
	logger.Cookie().Debug("Getting cookie by value")

	if value == "" {
		return nil, NewValidationError("value", "cookie value cannot be empty", value)
	}

	row := c.db.QueryRow(selectCookieByValueQuery, value)
	cookie, err := c.scanCookieRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, NewDatabaseError("select", "cookies", "failed to get cookie by value", err)
	}

	return cookie, nil
}

// Add creates a new cookie session in the database.
func (c *Cookies) Add(cookie *Cookie) error {
	logger.Cookie().Info("Adding cookie: %+v", cookie)

	if cookie == nil {
		return NewValidationError("cookie", "cookie cannot be nil", nil)
	}
	if cookie.Value == "" {
		return NewValidationError("value", "cookie value cannot be empty", cookie.Value)
	}
	if cookie.ApiKey == "" {
		return NewValidationError("api_key", "API key cannot be empty", cookie.ApiKey)
	}
	if cookie.LastLogin <= 0 {
		return NewValidationError("last_login", "last login timestamp must be positive", cookie.LastLogin)
	}

	var count int
	err := c.db.QueryRow(selectCookieExistsQuery, cookie.Value).Scan(&count)
	if err != nil {
		return NewDatabaseError("select", "cookies", "failed to check cookie existence", err)
	}
	if count > 0 {
		return ErrAlreadyExists
	}

	_, err = c.db.Exec(insertCookieQuery, cookie.UserAgent, cookie.Value, cookie.ApiKey, cookie.LastLogin)
	if err != nil {
		return NewDatabaseError("insert", "cookies", "failed to insert cookie", err)
	}

	return nil
}

// Update modifies an existing cookie session.
func (c *Cookies) Update(value string, cookie *Cookie) error {
	logger.Cookie().Info("Updating cookie: %+v, value: %s", cookie, value)

	if value == "" {
		return NewValidationError("value", "current cookie value cannot be empty", value)
	}
	if cookie == nil {
		return NewValidationError("cookie", "cookie cannot be nil", nil)
	}
	if cookie.Value == "" {
		return NewValidationError("value", "new cookie value cannot be empty", cookie.Value)
	}
	if cookie.ApiKey == "" {
		return NewValidationError("api_key", "API key cannot be empty", cookie.ApiKey)
	}
	if cookie.LastLogin <= 0 {
		return NewValidationError("last_login", "last login timestamp must be positive", cookie.LastLogin)
	}

	result, err := c.db.Exec(updateCookieQuery, cookie.UserAgent, cookie.Value, cookie.ApiKey, cookie.LastLogin, value)
	if err != nil {
		return NewDatabaseError("update", "cookies", "failed to update cookie", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return NewDatabaseError("update", "cookies", "failed to get rows affected", err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// Remove deletes a cookie session by its value.
func (c *Cookies) Remove(value string) error {
	logger.Cookie().Info("Removing cookie, value: %s", value)

	if value == "" {
		return NewValidationError("value", "cookie value cannot be empty", value)
	}

	_, err := c.db.Exec(deleteCookieByValueQuery, value)
	if err != nil {
		return NewDatabaseError("delete", "cookies", "failed to delete cookie", err)
	}

	return nil
}

// RemoveApiKey removes all cookie sessions associated with a specific API key.
func (c *Cookies) RemoveApiKey(apiKey string) error {
	logger.Cookie().Info("Removing cookies by API key")

	if apiKey == "" {
		return NewValidationError("api_key", "API key cannot be empty", apiKey)
	}

	_, err := c.db.Exec(deleteCookiesByAPIKeyQuery, apiKey)
	if err != nil {
		return NewDatabaseError("delete", "cookies", "failed to delete cookies by API key", err)
	}

	return nil
}

// RemoveExpired removes all cookie sessions that are older than the specified timestamp.
func (c *Cookies) RemoveExpired(beforeTimestamp int64) (int64, error) {
	logger.Cookie().Info("Removing expired cookies, before_timestamp: %d", beforeTimestamp)

	if beforeTimestamp <= 0 {
		return 0, NewValidationError("timestamp", "timestamp must be positive", beforeTimestamp)
	}

	result, err := c.db.Exec(deleteCookiesBeforeTimeQuery, beforeTimestamp)
	if err != nil {
		return 0, NewDatabaseError("delete", "cookies", "failed to delete expired cookies", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, NewDatabaseError("delete", "cookies", "failed to get rows affected", err)
	}

	return rowsAffected, nil
}

func (c *Cookies) scanCookie(rows *sql.Rows) (*Cookie, error) {
	cookie := &Cookie{}
	err := rows.Scan(&cookie.UserAgent, &cookie.Value, &cookie.ApiKey, &cookie.LastLogin)
	if err != nil {
		return nil, NewDatabaseError("scan", "cookies", "failed to scan row", err)
	}
	return cookie, nil
}

func (c *Cookies) scanCookieRow(row *sql.Row) (*Cookie, error) {
	cookie := &Cookie{}
	err := row.Scan(&cookie.UserAgent, &cookie.Value, &cookie.ApiKey, &cookie.LastLogin)
	if err != nil {
		return nil, err
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
