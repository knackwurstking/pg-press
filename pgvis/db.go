package pgvis

import (
	"database/sql"
)

// DB contains all database tables
type DB struct {
	Users          *DBUsers
	Cookies        *DBCookies
	TroubleReports *DBTroubleReports

	db *sql.DB
}

func NewDB(db *sql.DB) *DB {
	return &DB{
		Users:          NewDBUsers(db),
		Cookies:        NewDBCookies(db),
		TroubleReports: NewDBTroubleReports(db),
		db:             db,
	}
}
