package services

import "database/sql"

type Registry struct {
	DB *sql.DB

	Cookies       *Cookies
	Users         *Users
	Feeds         *Feeds
	Attachments   *Attachments
	Modifications *Modifications
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
	registry.Attachments = NewAttachments(registry)
	registry.Modifications = NewModifications(registry)

	return registry
}
