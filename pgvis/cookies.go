// Package pgvis cookie and session management.
//
// This file provides database operations for managing authentication cookies
// and user sessions. It handles session creation, validation, cleanup, and
// provides secure database operations using parameterized queries to prevent
// SQL injection attacks.
package pgvis

import (
	"database/sql"
	"slices"
)

// SQL queries for cookie operations
const (
	// createCookiesTableQuery creates the cookies table if it doesn't exist
	createCookiesTableQuery = `
		CREATE TABLE IF NOT EXISTS cookies (
			user_agent TEXT NOT NULL,
			value TEXT NOT NULL,
			api_key TEXT NOT NULL,
			last_login INTEGER NOT NULL,
			PRIMARY KEY("value")
		);
	`

	// selectAllCookiesQuery retrieves all cookies ordered by last login descending
	selectAllCookiesQuery = `SELECT * FROM cookies ORDER BY last_login DESC`

	// selectCookiesByAPIKeyQuery retrieves cookies for a specific API key
	selectCookiesByAPIKeyQuery = `SELECT * FROM cookies WHERE api_key = ? ORDER BY last_login DESC`

	// selectCookieByValueQuery retrieves a specific cookie by its value
	selectCookieByValueQuery = `SELECT * FROM cookies WHERE value = ?`

	// selectCookieExistsQuery checks if a cookie exists by value
	selectCookieExistsQuery = `SELECT COUNT(*) FROM cookies WHERE value = ?`

	// insertCookieQuery creates a new cookie session
	insertCookieQuery = `INSERT INTO cookies (user_agent, value, api_key, last_login) VALUES (?, ?, ?, ?)`

	// updateCookieQuery updates an existing cookie session
	updateCookieQuery = `UPDATE cookies SET user_agent = ?, value = ?, api_key = ?, last_login = ? WHERE value = ?`

	// deleteCookieByValueQuery removes a cookie by its value
	deleteCookieByValueQuery = `DELETE FROM cookies WHERE value = ?`

	// deleteCookiesByAPIKeyQuery removes all cookies for a specific API key
	deleteCookiesByAPIKeyQuery = `DELETE FROM cookies WHERE api_key = ?`

	// deleteCookiesBeforeTimeQuery removes cookies older than specified timestamp
	deleteCookiesBeforeTimeQuery = `DELETE FROM cookies WHERE last_login < ?`
)

// Cookies provides database operations for managing authentication cookies and sessions.
// It handles session lifecycle management and provides secure database operations
// for cookie storage and retrieval.
type Cookies struct {
	db *sql.DB // Database connection
}

// NewCookies creates a new Cookies instance and initializes the database table.
// It creates the cookies table if it doesn't exist and sets up the necessary
// database schema for session management.
//
// Parameters:
//   - db: Active database connection
//
// Returns:
//   - *Cookies: Initialized cookies handler
//
// Panics if the database table cannot be created.
func NewCookies(db *sql.DB) *Cookies {
	if _, err := db.Exec(createCookiesTableQuery); err != nil {
		panic(NewDatabaseError("create_table", "cookies",
			"failed to create cookies table", err))
	}

	return &Cookies{
		db: db,
	}
}

// List retrieves all cookies ordered by last login time (most recent first).
//
// Returns:
//   - []*Cookie: Slice of all cookies
//   - error: Database or scanning error
func (c *Cookies) List() ([]*Cookie, error) {
	rows, err := c.db.Query(selectAllCookiesQuery)
	if err != nil {
		return nil, NewDatabaseError("select", "cookies",
			"failed to query cookies", err)
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
		return nil, NewDatabaseError("select", "cookies",
			"error iterating over rows", err)
	}

	return cookies, nil
}

// ListApiKey retrieves all cookies associated with a specific API key.
// This is useful for finding all active sessions for a particular user.
//
// Parameters:
//   - apiKey: The API key to filter by
//
// Returns:
//   - []*Cookie: Slice of cookies for the specified API key
//   - error: Database or scanning error
func (c *Cookies) ListApiKey(apiKey string) ([]*Cookie, error) {
	if apiKey == "" {
		return nil, NewValidationError("api_key", "API key cannot be empty", apiKey)
	}

	rows, err := c.db.Query(selectCookiesByAPIKeyQuery, apiKey)
	if err != nil {
		return nil, NewDatabaseError("select", "cookies",
			"failed to query cookies by API key", err)
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
		return nil, NewDatabaseError("select", "cookies",
			"error iterating over rows", err)
	}

	return cookies, nil
}

// Get retrieves a specific cookie by its value.
// This is used for session validation during authentication.
//
// Parameters:
//   - value: The cookie value to look up
//
// Returns:
//   - *Cookie: The requested cookie
//   - error: ErrNotFound if not found, or database error
func (c *Cookies) Get(value string) (*Cookie, error) {
	if value == "" {
		return nil, NewValidationError("value", "cookie value cannot be empty", value)
	}

	row := c.db.QueryRow(selectCookieByValueQuery, value)

	cookie, err := c.scanCookieRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, NewDatabaseError("select", "cookies",
			"failed to get cookie by value", err)
	}

	return cookie, nil
}

// Add creates a new cookie session in the database.
//
// Parameters:
//   - cookie: The cookie session to create
//
// Returns:
//   - error: Validation, database, or duplicate error
func (c *Cookies) Add(cookie *Cookie) error {
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

	// Check if cookie already exists
	var count int
	err := c.db.QueryRow(selectCookieExistsQuery, cookie.Value).Scan(&count)
	if err != nil {
		return NewDatabaseError("select", "cookies",
			"failed to check cookie existence", err)
	}

	if count > 0 {
		return ErrAlreadyExists
	}

	// Insert the new cookie
	_, err = c.db.Exec(insertCookieQuery,
		cookie.UserAgent, cookie.Value, cookie.ApiKey, cookie.LastLogin)
	if err != nil {
		return NewDatabaseError("insert", "cookies",
			"failed to insert cookie", err)
	}

	return nil
}

// Update modifies an existing cookie session.
//
// Parameters:
//   - value: The current cookie value to update
//   - cookie: The updated cookie data
//
// Returns:
//   - error: Validation, database, or not found error
func (c *Cookies) Update(value string, cookie *Cookie) error {
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

	result, err := c.db.Exec(updateCookieQuery,
		cookie.UserAgent, cookie.Value, cookie.ApiKey, cookie.LastLogin, value)
	if err != nil {
		return NewDatabaseError("update", "cookies",
			"failed to update cookie", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return NewDatabaseError("update", "cookies",
			"failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// Remove deletes a cookie session by its value.
//
// Parameters:
//   - value: The cookie value to remove
//
// Returns:
//   - error: Database error (returns success even if cookie not found)
func (c *Cookies) Remove(value string) error {
	if value == "" {
		return NewValidationError("value", "cookie value cannot be empty", value)
	}

	_, err := c.db.Exec(deleteCookieByValueQuery, value)
	if err != nil {
		return NewDatabaseError("delete", "cookies",
			"failed to delete cookie", err)
	}

	return nil
}

// RemoveApiKey removes all cookie sessions associated with a specific API key.
// This is useful for logging out a user from all devices.
//
// Parameters:
//   - apiKey: The API key whose sessions should be removed
//
// Returns:
//   - error: Database error
func (c *Cookies) RemoveApiKey(apiKey string) error {
	if apiKey == "" {
		return NewValidationError("api_key", "API key cannot be empty", apiKey)
	}

	_, err := c.db.Exec(deleteCookiesByAPIKeyQuery, apiKey)
	if err != nil {
		return NewDatabaseError("delete", "cookies",
			"failed to delete cookies by API key", err)
	}

	return nil
}

// RemoveExpired removes all cookie sessions that are older than the specified timestamp.
// This is useful for cleaning up expired sessions.
//
// Parameters:
//   - beforeTimestamp: Unix timestamp; cookies older than this will be removed
//
// Returns:
//   - int64: Number of cookies removed
//   - error: Database error
func (c *Cookies) RemoveExpired(beforeTimestamp int64) (int64, error) {
	if beforeTimestamp <= 0 {
		return 0, NewValidationError("timestamp", "timestamp must be positive", beforeTimestamp)
	}

	result, err := c.db.Exec(deleteCookiesBeforeTimeQuery, beforeTimestamp)
	if err != nil {
		return 0, NewDatabaseError("delete", "cookies",
			"failed to delete expired cookies", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, NewDatabaseError("delete", "cookies",
			"failed to get rows affected", err)
	}

	return rowsAffected, nil
}

// scanCookie scans a cookie from database rows.
// This is a helper method for handling multiple rows from SELECT queries.
func (c *Cookies) scanCookie(rows *sql.Rows) (*Cookie, error) {
	cookie := &Cookie{}

	err := rows.Scan(&cookie.UserAgent, &cookie.Value, &cookie.ApiKey, &cookie.LastLogin)
	if err != nil {
		return nil, NewDatabaseError("scan", "cookies",
			"failed to scan row", err)
	}

	return cookie, nil
}

// scanCookieRow scans a cookie from a single database row.
// This is a helper method for handling single row results from QueryRow.
func (c *Cookies) scanCookieRow(row *sql.Row) (*Cookie, error) {
	cookie := &Cookie{}

	err := row.Scan(&cookie.UserAgent, &cookie.Value, &cookie.ApiKey, &cookie.LastLogin)
	if err != nil {
		return nil, err // Return raw error for ErrNoRows detection
	}

	return cookie, nil
}

// SortCookies sorts a slice of cookies by last login time in descending order.
// This function provides client-side sorting for cases where database ordering
// is not sufficient or when working with pre-loaded cookie slices.
//
// Parameters:
//   - cookies: Slice of cookies to sort
//
// Returns:
//   - []*Cookie: New slice with cookies sorted by last login (most recent first)
func SortCookies(cookies []*Cookie) []*Cookie {
	if len(cookies) <= 1 {
		return cookies
	}

	cookiesSorted := make([]*Cookie, 0, len(cookies))

outer:
	for _, cookie := range cookies {
		// Insert cookie in the correct position to maintain descending order
		for i, sortedCookie := range cookiesSorted {
			if cookie.LastLogin > sortedCookie.LastLogin {
				cookiesSorted = slices.Insert(cookiesSorted, i, cookie)
				continue outer
			}
		}
		// If we get here, this cookie has the lowest LastLogin so far
		cookiesSorted = append(cookiesSorted, cookie)
	}

	return cookiesSorted
}
