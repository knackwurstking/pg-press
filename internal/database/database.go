package database

import (
	"database/sql"

	"github.com/knackwurstking/pgpress/internal/services"
)

// DB represents the main database connection and provides access to all data access objects.
type DB struct {
	db *sql.DB

	// Kind of DataOperations
	PressCycles    *services.PressCycles
	Users          *services.User
	TroubleReports *services.TroubleReports
	Notes          *services.Notes
	Tools          *services.Tools
	MetalSheets    *services.MetalSheets

	Attachments       *services.Attachment
	Cookies           *services.Cookie
	ToolRegenerations *services.ToolRegenerations
	Feeds             *services.Feed
	Modifications     *services.Modifications
}

// New creates a new DB instance with all necessary table handlers initialized.
// Feeds must be created before Users and TroubleReports as they generate feed entries.
func New(db *sql.DB) *DB {
	modifications := services.NewModifications(db)

	feeds := services.NewFeed(db)

	attachments := services.NewAttachment(db)
	troubleReports := services.NewTroubleReports(db, attachments, modifications)

	pressCycles := services.NewPressCycles(db)

	notes := services.NewNotes(db)
	tools := services.NewTools(db, notes)

	metalSheets := services.NewMetalSheets(db, notes)
	toolRegenerations := services.NewToolRegenerations(db, tools)

	dbInstance := &DB{
		Users:             services.NewUser(db, feeds),
		Cookies:           services.NewCookie(db),
		Attachments:       attachments,
		TroubleReports:    troubleReports,
		Notes:             notes,
		Tools:             tools,
		MetalSheets:       metalSheets,
		PressCycles:       pressCycles,
		ToolRegenerations: toolRegenerations,
		Modifications:     modifications,
		Feeds:             feeds,
		db:                db,
	}

	return dbInstance
}

// GetDB returns the underlying sql.DB connection
func (db *DB) GetDB() *sql.DB {
	return db.db
}
