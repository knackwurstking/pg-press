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
	Users          *services.Users
	TroubleReports *services.TroubleReports
	Notes          *services.Notes
	Tools          *services.Tools
	MetalSheets    *services.MetalSheets

	Attachments       *services.Attachments
	Cookies           *services.Cookies
	ToolRegenerations *services.ToolRegenerations
	Feeds             *services.Feeds
	Modifications     *services.Modifications
}

// New creates a new DB instance with all necessary table handlers initialized.
// Feeds must be created before Users and TroubleReports as they generate feed entries.
func New(db *sql.DB) *DB {
	feeds := services.NewFeeds(db)
	modifications := services.NewModifications(db)
	attachments := services.NewAttachments(db)
	notes := services.NewNotes(db)

	tools := services.NewTools(db, notes)
	toolRegenerations := services.NewToolRegenerations(db, tools)

	dbInstance := &DB{
		Users:             services.NewUsers(db),
		Cookies:           services.NewCookies(db),
		Attachments:       attachments,
		TroubleReports:    services.NewTroubleReports(db, attachments, modifications),
		Notes:             notes,
		Tools:             tools,
		MetalSheets:       services.NewMetalSheets(db, notes),
		PressCycles:       services.NewPressCycles(db),
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
