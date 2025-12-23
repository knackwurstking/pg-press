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
	sqlCreateCookiesTable string = `
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

	sqlListCookies string = `
		SELECT user_agent, value, user_id, last_login
		FROM cookies;
	`

	sqlListCookiesByUserID string = `
		SELECT user_agent, value, user_id, last_login
		FROM cookies
		WHERE user_id = :user_id
		ORDER BY last_login DESC;
	`

	sqlListCookiesByApiKey string = `
		SELECT user_agent, value, user_id, last_login
		FROM cookies
		WHERE user_id = :user_id
		ORDER BY last_login DESC;
	`

	sqlGetCookie string = `
		SELECT user_agent, value, user_id, last_login
		FROM cookies
		WHERE value = :value;
	`

	sqlAddCookie string = `
		INSERT INTO cookies (user_agent, value, user_id, last_login)
		VALUES (:user_agent, :value, :user_id, :last_login);
	`

	sqlUpdateCookie string = `
		UPDATE cookies
		SET
			user_agent = :user_agent,
			value = :value
			user_id = :user_id
			last_login = :last_login
		WHERE value = :value;
	`

	sqlDeleteCookie string = `
		DELETE FROM cookies
		WHERE value = :value;
	`

	sqlDeleteCookiesByUserID string = `
		DELETE FROM cookies
		WHERE user_id = :user_id;
	`
)

// -----------------------------------------------------------------------------
// Table Helpers: "cookies"
// -----------------------------------------------------------------------------

func ListCookies() (cookies []*shared.Cookie, merr *errors.MasterError) {
	rows, err := dbUser.Query(sqlListCookies)
	if err != nil {
		return nil, errors.NewMasterError(err)
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

func ListCookiesByUserID(userID shared.TelegramID) (cookies []*shared.Cookie, merr *errors.MasterError) {
	rows, err := dbUser.Query(sqlListCookiesByUserID, sql.Named("user_id", userID))
	if err != nil {
		return nil, errors.NewMasterError(err)
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

func ListCookiesByApiKey(apiKey string) (cookies []*shared.Cookie, merr *errors.MasterError) {
	// Get user id from the users table
	user, merr := GetUserByApiKey(apiKey)
	if merr != nil {
		return nil, merr
	}

	rows, err := dbUser.Query(sqlListCookiesByApiKey, sql.Named("user_id", user.ID))
	if err != nil {
		return nil, errors.NewMasterError(err)
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

func GetCookie(value string) (*shared.Cookie, *errors.MasterError) {
	return ScanCookie(dbUser.QueryRow(sqlGetCookie, sql.Named("value", value)))
}

func AddCookie(cookie *shared.Cookie) *errors.MasterError {
	if verr := cookie.Validate(); verr != nil {
		return verr.MasterError().Wrap("invalid cookie data")
	}

	_, err := dbUser.Exec(sqlAddCookie,
		sql.Named("user_agent", cookie.UserAgent),
		sql.Named("value", cookie.Value),
		sql.Named("user_id", cookie.UserID),
		sql.Named("last_login", cookie.LastLogin),
	)
	if err != nil {
		return errors.NewMasterError(err)
	}
	return nil
}

// UpdateCookie updates the given cookie in the database, it just replaces all fields including the value.
func UpdateCookie(value string, cookie *shared.Cookie) *errors.MasterError {
	if verr := cookie.Validate(); verr != nil {
		return verr.MasterError().Wrap("invalid cookie data")
	}

	_, err := dbUser.Exec(sqlUpdateCookie,
		sql.Named("user_agent", cookie.UserAgent),
		sql.Named("user_id", cookie.UserID),
		sql.Named("last_login", cookie.LastLogin),
		sql.Named("value", value),
	)
	if err != nil {
		return errors.NewMasterError(err)
	}
	return nil
}

func DeleteCookie(value string) *errors.MasterError {
	_, err := dbUser.Exec(sqlDeleteCookie, sql.Named("value", value))
	if err != nil {
		return errors.NewMasterError(err)
	}
	return nil
}

func DeleteCookiesByUserID(userID shared.TelegramID) *errors.MasterError {
	_, err := dbUser.Exec(sqlDeleteCookiesByUserID, sql.Named("user_id", userID))
	if err != nil {
		return errors.NewMasterError(err)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Scanners
// -----------------------------------------------------------------------------

func ScanCookie(row Scannable) (*shared.Cookie, *errors.MasterError) {
	var c shared.Cookie
	err := row.Scan(
		&c.UserAgent,
		&c.Value,
		&c.UserID,
		&c.LastLogin,
	)
	if err != nil {
		return nil, errors.NewMasterError(err)
	}
	return &c, nil
}
