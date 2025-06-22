package pgvis

import "database/sql"

// DB contains all database tables
type DB struct {
	Users *DBUsers

	db *sql.DB
}

func NewDB(db *sql.DB) *DB {
	return &DB{
		Users: NewDBUsers(db),
	}
}

type DBUsers struct {
	db *sql.DB
}

func NewDBUsers(db *sql.DB) *DBUsers {
	return &DBUsers{}
}

func (db *DBUsers) List() ([]*User, error) {
	users := NewUsers()

	// TODO: ...

	return users, nil
}
