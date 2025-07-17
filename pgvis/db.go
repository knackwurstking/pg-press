package pgvis

import (
	"database/sql"
)

// DB represents the main database connection and provides access to all data access objects.
// It serves as the central point for database operations and maintains references to
// all table-specific handlers.
type DB struct {
	// Users provides access to user management operations
	Users *Users
	// Cookies handles session and cookie management
	Cookies *Cookies
	// TroubleReports manages trouble report operations
	TroubleReports *TroubleReports
	// Feeds handles activity feed operations
	Feeds *Feeds

	// db is the underlying database connection
	db *sql.DB
}

// New creates a new DB instance with all necessary table handlers initialized.
// It establishes the database schema and creates all required tables if they don't exist.
//
// The initialization order is important as some tables depend on others:
//   - Feeds must be created before Users and TroubleReports as they generate feed entries
//   - Users and TroubleReports can be created in any order after Feeds
//
// Parameters:
//   - db: An active SQL database connection
//
// Returns:
//   - *DB: A fully initialized database handler
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
