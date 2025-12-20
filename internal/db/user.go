package db

import (
	"database/sql"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

// -----------------------------------------------------------------------------
// Table Creation Statements
// -----------------------------------------------------------------------------

const (
	SQLCreateUsersTable string = `
		CREATE TABLE IF NOT EXISTS users (
			id 			INTEGER NOT NULL,
			name 		TEXT NOT NULL,
			api_key 	TEXT NOT NULL UNIQUE,

			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`

	SQLCreateCookiesTable string = `
		CREATE TABLE IF NOT EXISTS cookies (
			user_agent 	TEXT NOT NULL,
			value 		TEXT NOT NULL,
			user_id 	INTEGER NOT NULL,
			last_login 	INTEGER NOT NULL,

			PRIMARY KEY("value")
		);

		-- Index to quickly find cookies by user_id

		CREATE INDEX IF NOT EXISTS idx_cookies_user_id
		ON cookies(user_id);
	`
)

// -----------------------------------------------------------------------------
// Table Helpers: "users"
// -----------------------------------------------------------------------------

const SQLGetUserByID string = `
	SELECT id, name, api_key
	FROM users
	WHERE id = :id;
`

func GetUserByID(id shared.TelegramID) (*shared.User, *errors.MasterError) {
	return ScanUser(DBUser.QueryRow(SQLGetUserByID, sql.Named("id", id)))
}

const SQLGetUserByApiKey string = `
	SELECT id, name, api_key
	FROM users
	WHERE api_key = :api_key;
`

func GetUserByApiKey(apiKey string) (user *shared.User, merr *errors.MasterError) {
	return ScanUser(DBUser.QueryRow(SQLGetUserByApiKey, sql.Named("api_key", apiKey)))
}

// -----------------------------------------------------------------------------
// Table Helpers: "cookies"
// -----------------------------------------------------------------------------

const SQLGetCookie string = `
	SELECT user_agent, value, user_id, last_login
	FROM cookies
	WHERE value = :value;
`

func GetCookie(value string) (*shared.Cookie, *errors.MasterError) {
	return ScanCookie(DBUser.QueryRow(SQLGetCookie, sql.Named("value", value)))
}

const SQLAddCookie string = `
	INSERT INTO cookies (user_agent, value, user_id, last_login)
	VALUES (:user_agent, :value, :user_id, :last_login);
`

func AddCookie(cookie *shared.Cookie) *errors.MasterError {
	_, err := DBUser.Exec(SQLAddCookie,
		sql.Named("user_agent", cookie.UserAgent),
		sql.Named("value", cookie.Value),
		sql.Named("user_id", cookie.UserID),
		sql.Named("last_login", cookie.LastLogin),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}
	return nil
}

const SQLUpdateCookie string = `
	UPDATE cookies
	SET
		user_agent = :user_agent,
		value = :value
		user_id = :user_id
		last_login = :last_login
	WHERE value = :value;
`

// UpdateCookie updates the given cookie in the database, it just replaces all fields including the value.
func UpdateCookie(value string, cookie *shared.Cookie) *errors.MasterError {
	_, err := DBUser.Exec(SQLUpdateCookie,
		sql.Named("user_agent", cookie.UserAgent),
		sql.Named("user_id", cookie.UserID),
		sql.Named("last_login", cookie.LastLogin),
		sql.Named("value", value),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}
	return nil
}

const SQLListCookiesByApiKey string = `
	SELECT user_agent, value, user_id, last_login
	FROM cookies
	WHERE user_id = :user_id
	ORDER BY last_login DESC;
`

func ListCookiesByApiKey(apiKey string) (cookies []*shared.Cookie, merr *errors.MasterError) {
	// Get user id from the users table
	user, merr := GetUserByApiKey(apiKey)
	if merr != nil {
		return nil, merr
	}

	rows, err := DBUser.Query(SQLListCookiesByApiKey, sql.Named("user_id", user.ID))
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	for rows.Next() {
		cookie, merr := ScanCookie(rows)
		if merr != nil {
			return nil, merr
		}
		cookies = append(cookies, cookie)
	}

	return cookies, nil
}

// TODO: DeleteCookie
// TODO: DeleteCookiesByUserID

// -----------------------------------------------------------------------------
// Scanners
// -----------------------------------------------------------------------------

func ScanUser(row Scannable) (*shared.User, *errors.MasterError) {
	var u shared.User
	err := row.Scan(
		&u.ID,
		&u.Name,
		&u.ApiKey,
	)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	return &u, nil
}

func ScanCookie(row Scannable) (*shared.Cookie, *errors.MasterError) {
	var c shared.Cookie
	err := row.Scan(
		&c.UserAgent,
		&c.Value,
		&c.UserID,
		&c.LastLogin,
	)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	return &c, nil
}
