package services

import (
	"database/sql"

	"github.com/knackwurstking/pgpress/internal/services/entities/attachments"
	"github.com/knackwurstking/pgpress/internal/services/entities/cookies"
	"github.com/knackwurstking/pgpress/internal/services/entities/feeds"
)

type Registry struct {
	db *sql.DB

	Attachments *attachments.Service
	Cookies     *cookies.Service
	Feeds       *feeds.Service
}

// New creates a new DB instance with all necessary table handlers initialized.
// Feeds must be created before Users and TroubleReports as they generate feed entries.
func NewRegistry(db *sql.DB) *Registry {
	return &Registry{
		db:          db,
		Attachments: attachments.NewService(db),
		Cookies:     cookies.NewService(db),
		Feeds:       feeds.NewService(db),
	}
}

// GetDB returns the underlying sql.DB connection
func (db *Registry) GetSQL() *sql.DB {
	return db.db
}
