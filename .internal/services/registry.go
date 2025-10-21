package services

import (
	"database/sql"

	"github.com/knackwurstking/pgpress/internal/services/entities/attachments"
	"github.com/knackwurstking/pgpress/internal/services/entities/cookies"
	"github.com/knackwurstking/pgpress/internal/services/entities/feeds"
	"github.com/knackwurstking/pgpress/internal/services/entities/metalsheets"
	"github.com/knackwurstking/pgpress/internal/services/entities/modifications"
	"github.com/knackwurstking/pgpress/internal/services/entities/notes"
	"github.com/knackwurstking/pgpress/internal/services/entities/presscycles"
	"github.com/knackwurstking/pgpress/internal/services/entities/toolregenerations"
	"github.com/knackwurstking/pgpress/internal/services/entities/tools"
	"github.com/knackwurstking/pgpress/internal/services/entities/troublereports"
	"github.com/knackwurstking/pgpress/internal/services/entities/users"
)

type Registry struct {
	db *sql.DB

	Attachments       *attachments.Service
	Cookies           *cookies.Service
	Feeds             *feeds.Service
	Users             *users.Service
	Modifications     *modifications.Service
	MetalSheets       *metalsheets.Service
	Notes             *notes.Service
	PressCycles       *presscycles.Service
	Tools             *tools.Service
	ToolRegenerations *toolregenerations.Service
	TroubleReports    *troublereports.Service
}

// New creates a new DB instance with all necessary table handlers initialized.
// Feeds must be created before Users and TroubleReports as they generate feed entries.
func NewRegistry(db *sql.DB) *Registry {
	notesService := notes.NewService(db)
	attachmentsService := attachments.NewService(db)
	modificationsService := modifications.NewService(db)
	toolsService := tools.NewService(db, notesService)
	pressCyclesService := presscycles.NewService(db)

	return &Registry{
		db:                db,
		Attachments:       attachmentsService,
		Cookies:           cookies.NewService(db),
		Feeds:             feeds.NewService(db),
		Users:             users.NewService(db),
		Modifications:     modificationsService,
		MetalSheets:       metalsheets.NewService(db, notesService),
		PressCycles:       pressCyclesService,
		Tools:             toolsService,
		ToolRegenerations: toolregenerations.NewService(db, toolsService, pressCyclesService),
		TroubleReports:    troublereports.NewService(db, attachmentsService, modificationsService),

		Notes: notesService,
	}
}

// GetSQL returns the underlying sql.DB connection
func (db *Registry) GetSQL() *sql.DB {
	return db.db
}
