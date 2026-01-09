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
	sqlCreateUsersTable string = `
		CREATE TABLE IF NOT EXISTS users (
			id 			INTEGER NOT NULL,
			name 		TEXT NOT NULL,
			api_key 	TEXT NOT NULL UNIQUE,

			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`

	sqlGetUser string = `
		SELECT id, name, api_key
		FROM users
		WHERE id = :id;
	`

	sqlGetUserByApiKey string = `
		SELECT id, name, api_key
		FROM users
		WHERE api_key = :api_key;
	`

	sqlAddUser string = `
		INSERT INTO users (id, name, api_key)
		VALUES (:id, :name, :api_key);
	`

	sqlUpdateUser string = `
		UPDATE users
		SET name = :name,
			api_key = :api_key
		WHERE id = :id;
	`

	sqlListUsers string = `
		SELECT id, name, api_key
		FROM users
		ORDER BY id ASC;
	`

	sqlDeleteUser string = `
		DELETE FROM users
		WHERE id = :id;
	`
)

// -----------------------------------------------------------------------------
// User Functions
// -----------------------------------------------------------------------------

// GetUser retrieves a user by its ID
func GetUser(id shared.TelegramID) (*shared.User, *errors.HTTPError) {
	return ScanUser(dbUser.QueryRow(sqlGetUser, sql.Named("id", id)))
}

// GetUserByApiKey retrieves a user by its API key
func GetUserByApiKey(apiKey string) (user *shared.User, merr *errors.HTTPError) {
	return ScanUser(dbUser.QueryRow(sqlGetUserByApiKey, sql.Named("api_key", apiKey)))
}

// AddUser adds a new user to the database
func AddUser(user *shared.User) *errors.HTTPError {
	if verr := user.Validate(); verr != nil {
		return verr.HTTPError().Wrap("invalid user data")
	}

	_, err := dbUser.Exec(sqlAddUser,
		sql.Named("id", user.ID),
		sql.Named("name", user.Name),
		sql.Named("api_key", user.ApiKey),
	)
	if err != nil {
		return errors.NewHTTPError(err)
	}
	return nil
}

// UpdateUser updates an existing user in the database
func UpdateUser(user *shared.User) *errors.HTTPError {
	if verr := user.Validate(); verr != nil {
		return verr.HTTPError().Wrap("invalid user data")
	}

	_, err := dbUser.Exec(sqlUpdateUser,
		sql.Named("id", user.ID),
		sql.Named("name", user.Name),
		sql.Named("api_key", user.ApiKey),
	)
	if err != nil {
		return errors.NewHTTPError(err)
	}
	return nil
}

// ListUsers retrieves all users from the database
func ListUsers() (users []*shared.User, merr *errors.HTTPError) {
	rows, err := dbUser.Query(sqlListUsers)
	if err != nil {
		return nil, errors.NewHTTPError(err)
	}
	defer rows.Close()

	for rows.Next() {
		user, merr := ScanUser(rows)
		if merr != nil {
			return nil, merr
		}
		users = append(users, user)
	}

	return users, nil
}

// DeleteUser removes a user from the database
func DeleteUser(id shared.TelegramID) *errors.HTTPError {
	_, err := dbUser.Exec(sqlDeleteUser, sql.Named("id", id))
	if err != nil {
		return errors.NewHTTPError(err)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Scan Helpers
// -----------------------------------------------------------------------------

// ScanUser scans a database row into a User struct
func ScanUser(row Scannable) (*shared.User, *errors.HTTPError) {
	var u shared.User
	err := row.Scan(
		&u.ID,
		&u.Name,
		&u.ApiKey,
	)
	if err != nil {
		return nil, errors.NewHTTPError(err)
	}
	return &u, nil
}
