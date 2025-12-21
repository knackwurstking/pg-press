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

const SQLAddUser string = `
	INSERT INTO users (id, name, api_key)
	VALUES (:id, :name, :api_key);
`

func AddUser(user *shared.User) *errors.MasterError {
	if verr := user.Validate(); verr != nil {
		return verr.MasterError().Wrap("invalid user data")
	}

	_, err := DBUser.Exec(SQLAddUser,
		sql.Named("id", user.ID),
		sql.Named("name", user.Name),
		sql.Named("api_key", user.ApiKey),
	)
	if err != nil {
		return errors.NewMasterError(err)
	}
	return nil
}

const SQLUpdateUser string = `
	UPDATE users
	SET name = :name,
		api_key = :api_key
	WHERE id = :id;
`

func UpdateUser(user *shared.User) *errors.MasterError {
	if verr := user.Validate(); verr != nil {
		return verr.MasterError().Wrap("invalid user data")
	}

	_, err := DBUser.Exec(SQLUpdateUser,
		sql.Named("id", user.ID),
		sql.Named("name", user.Name),
		sql.Named("api_key", user.ApiKey),
	)
	if err != nil {
		return errors.NewMasterError(err)
	}
	return nil
}

const SQLListUsers string = `
	SELECT id, name, api_key
	FROM users
	ORDER BY id ASC;
`

func ListUsers() (users []*shared.User, merr *errors.MasterError) {
	rows, err := DBUser.Query(SQLListUsers)
	if err != nil {
		return nil, errors.NewMasterError(err)
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

const SQLDeleteUser string = `
	DELETE FROM users
	WHERE id = :id;
`

func DeleteUser(id shared.TelegramID) *errors.MasterError {
	_, err := DBUser.Exec(SQLDeleteUser, sql.Named("id", id))
	if err != nil {
		return errors.NewMasterError(err)
	}
	return nil
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
		return nil, errors.NewMasterError(err)
	}
	return &u, nil
}
