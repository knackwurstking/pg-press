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
	Feeds                *Feeds
	Migration            *Migration
	db                   *sql.DB
}

// New creates a new DB instance with all necessary table handlers initialized.
// Feeds must be created before Users and TroubleReports as they generate feed entries.
func New(db *sql.DB) *DB {
	// Create migration instance and run all necessary migrations first
	migration := NewMigration(db)
	if err := migration.RunAllMigrations(); err != nil {
		panic(NewDatabaseError("migration", "all_tables",
			"failed to run database migrations", err))
	}

	feeds := NewFeeds(db)
	attachments := NewAttachments(db)
	troubleReports := NewTroubleReports(db, feeds)
	troubleReportsHelper := NewTroubleReportsHelper(troubleReports, attachments)

	return &DB{
		Users:                NewUsers(db, feeds),
		Cookies:              NewCookies(db),
		Attachments:          attachments,
		TroubleReports:       troubleReports,
		TroubleReportsHelper: troubleReportsHelper,
		Feeds:                feeds,
		Migration:            migration,
		db:                   db,
	}
}
