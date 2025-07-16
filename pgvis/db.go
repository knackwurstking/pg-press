package pgvis

import (
	"database/sql"
)

// DB contains all database tables
type DB struct {
	Users          *Users
	Cookies        *Cookies
	TroubleReports *TroubleReports
	Feeds          *Feeds

	db *sql.DB
}

func New(db *sql.DB) *DB {
	feeds := NewFeeds(db)

	return &DB{
		Users:          NewUsers(db, feeds),
		Cookies:        NewCookies(db),
		TroubleReports: NewTroubleReports(db),
		Feeds:          feeds,
		db:             db,
	}
}
