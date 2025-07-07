package pgvis

import (
	"database/sql"
)

// DB contains all database tables
type DB struct {
	Users   *DBUsers
	Cookies *DBCookies

	db *sql.DB
}

func NewDB(db *sql.DB) *DB {
	return &DB{
		Users:   NewDBUsers(db),
		Cookies: NewDBCookies(db),
		db:      db,
	}
}
