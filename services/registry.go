package services

import "database/sql"

type Registry struct {
	DB *sql.DB

	//Attachments       *attachments.Service
	//Cookies           *cookies.Service
	//Feeds             *feeds.Service
	//Users             *users.Service
	//Modifications     *modifications.Service
	//MetalSheets       *metalsheets.Service
	//Notes             *notes.Service
	//PressCycles       *presscycles.Service
	//Tools             *tools.Service
	//ToolRegenerations *toolregenerations.Service
	//TroubleReports    *troublereports.Service
}

// New creates a new DB instance with all necessary table handlers initialized.
// Feeds must be created before Users and TroubleReports as they generate feed entries.
func NewRegistry(db *sql.DB) *Registry {
	//notesService := notes.NewService(db)
	//attachmentsService := attachments.NewService(db)
	//modificationsService := modifications.NewService(db)
	//toolsService := tools.NewService(db, notesService)
	//pressCyclesService := presscycles.NewService(db)

	return &Registry{
		DB: db,
		//Attachments:       attachmentsService,
		//Cookies:           cookies.NewService(db),
		//Feeds:             feeds.NewService(db),
		//Users:             users.NewService(db),
		//Modifications:     modificationsService,
		//MetalSheets:       metalsheets.NewService(db, notesService),
		//PressCycles:       pressCyclesService,
		//Tools:             toolsService,
		//ToolRegenerations: toolregenerations.NewService(db, toolsService, pressCyclesService),
		//TroubleReports:    troublereports.NewService(db, attachmentsService, modificationsService),

		//Notes: notesService,
	}
}
