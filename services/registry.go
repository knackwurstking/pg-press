package services

import "database/sql"

type Registry struct {
	DB *sql.DB

	Cookies *Cookies
	Users   *Users
	Feeds   *Feeds
	//Attachments       *attachments.Service
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
	registry.Feeds = NewFeeds(registry)

	return registry
}
