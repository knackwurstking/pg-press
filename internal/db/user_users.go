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
)

// -----------------------------------------------------------------------------
// Table Helpers: "users"
// -----------------------------------------------------------------------------

const SQLGetUser string = `
	SELECT id, name, api_key
	FROM users
	WHERE id = :id;
`

func GetUser(id shared.TelegramID) (*shared.User, *errors.MasterError) {
	return ScanUser(DBUser.QueryRow(SQLGetUser, sql.Named("id", id)))
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
