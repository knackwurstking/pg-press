package services

import "database/sql"

type Registry struct {
	DB *sql.DB

	Cookies *Cookies
	Users   *Users
	//Attachments       *attachments.Service
	//Feeds             *feeds.Service
	//Modifications     *modifications.Service
	//MetalSheets       *metalsheets.Service
	//Notes             *notes.Service
	//PressCycles       *presscycles.Service
	//Tools             *tools.Service
	//ToolRegenerations *toolregenerations.Service
	//TroubleReports    *troublereports.Service
}

func NewRegistry(db *sql.DB) *Registry {
	registry := &Registry{
		DB: db,
	}

	registry.Cookies = NewCookies(registry)
	registry.Users = NewUsers(registry)

	return registry
}
