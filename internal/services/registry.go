package services

import (
	"database/sql"

	"github.com/knackwurstking/pgpress/internal/services/entities/attachments"
	"github.com/knackwurstking/pgpress/internal/services/entities/cookies"
	"github.com/knackwurstking/pgpress/internal/services/entities/feeds"
	"github.com/knackwurstking/pgpress/internal/services/entities/metalsheets"
	"github.com/knackwurstking/pgpress/internal/services/entities/notes"
	"github.com/knackwurstking/pgpress/internal/services/entities/users"
)

type Registry struct {
	db *sql.DB

	Attachments *attachments.Service
	Cookies     *cookies.Service
	Feeds       *feeds.Service
	Users       *users.Service
	MetalSheets *metalsheets.Service
	Notes       *notes.Service
}

// New creates a new DB instance with all necessary table handlers initialized.
// Feeds must be created before Users and TroubleReports as they generate feed entries.
func NewRegistry(db *sql.DB) *Registry {
	notesService := notes.NewService(db)

	return &Registry{
		db:          db,
		Attachments: attachments.NewService(db),
		Cookies:     cookies.NewService(db),
		Feeds:       feeds.NewService(db),
		Users:       users.NewService(db),
		MetalSheets: metalsheets.NewService(db, notesService),
	}
}

// GetSQL returns the underlying sql.DB connection
func (db *Registry) GetSQL() *sql.DB {
	return db.db
}
