package types

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
