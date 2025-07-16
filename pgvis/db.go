package pgvis

import (
	"database/sql"
)

// DB contains all database tables
type DB struct {
	Users          *Users
	Cookies        *Cookies
	TroubleReports *TroubleReports

	db *sql.DB
}

func New(db *sql.DB) *DB {
	return &DB{
		Users:          NewUsers(db),
		Cookies:        NewCookies(db),
		TroubleReports: NewTroubleReports(db),
		db:             db,
	}
}
