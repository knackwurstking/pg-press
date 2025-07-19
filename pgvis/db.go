package pgvis

import (
	"database/sql"
)

// DB represents the main database connection and provides access to all data access objects.
type DB struct {
	Users          *Users
	Cookies        *Cookies
	TroubleReports *TroubleReports
	Feeds          *Feeds
	db             *sql.DB
}

// New creates a new DB instance with all necessary table handlers initialized.
// Feeds must be created before Users and TroubleReports as they generate feed entries.
func New(db *sql.DB) *DB {
	feeds := NewFeeds(db)

	return &DB{
		Users:          NewUsers(db, feeds),
		Cookies:        NewCookies(db),
		TroubleReports: NewTroubleReports(db, feeds),
		Feeds:          feeds,
		db:             db,
	}
}
