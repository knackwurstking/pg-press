package database

import (
	"database/sql"
)

// DB represents the main database connection and provides access to all data access objects.
type DB struct {
	Users                *Users
	Cookies              *Cookies
	Attachments          *Attachments
	TroubleReports       *TroubleReports
	TroubleReportsHelper *TroubleReportsHelper
	Tools                *Tools
	Feeds                *Feeds
	db                   *sql.DB
}

// New creates a new DB instance with all necessary table handlers initialized.
// Feeds must be created before Users and TroubleReports as they generate feed entries.
func New(db *sql.DB) *DB {
	feeds := NewFeeds(db)
	troubleReports := NewTroubleReports(db, feeds)

	attachments := NewAttachments(db)
	troubleReportsHelper := NewTroubleReportsHelper(troubleReports, attachments)

	return &DB{
		Users:                NewUsers(db, feeds),
		Cookies:              NewCookies(db),
		Attachments:          attachments,
		TroubleReports:       troubleReports,
		TroubleReportsHelper: troubleReportsHelper,
		Tools:                NewTools(db, feeds),
		Feeds:                feeds,
		db:                   db,
	}
}
