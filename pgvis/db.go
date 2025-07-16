package pgvis

import (
	"database/sql"
)

// DB contains all database tables
type DB struct {
	Users          *DBUsers
	Cookies        *Cookies
	TroubleReports *DBTroubleReports

	db *sql.DB
}

func New(db *sql.DB) *DB {
	return &DB{
		Users:          NewDBUsers(db),
		Cookies:        NewCookies(db),
		TroubleReports: NewDBTroubleReports(db),
		db:             db,
	}
}
