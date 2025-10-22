package services

import "database/sql"

type Registry struct {
	DB *sql.DB

	Cookies           *Cookies
	Users             *Users
	Feeds             *Feeds
	Attachments       *Attachments
	Modifications     *Modifications
	MetalSheets       *MetalSheets
	Notes             *Notes
	PressCycles       *PressCycles
	Tools             *Tools
	ToolRegenerations *ToolRegenerations
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
	registry.MetalSheets = NewMetalSheets(registry)
	registry.Notes = NewNotes(registry)
	registry.PressCycles = NewPressCycles(registry)
	registry.Tools = NewTools(registry)
	registry.ToolRegenerations = NewToolRegenerations(registry)

	return registry
}
