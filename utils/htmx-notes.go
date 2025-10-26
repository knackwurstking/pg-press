package utils

import (
	"fmt"
	"strings"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pgpress/models"
)

func HXGetNotesGrid() templ.SafeURL {
	return buildURL("/htmx/notes/grid", nil)
}

func HXGetNotesEditDialog(noteID *models.NoteID, linkToTables ...string) templ.SafeURL {
	params := make(map[string]string)

	if noteID != nil {
		params["id"] = fmt.Sprintf("%d", *noteID)
	}

	if len(linkToTables) > 0 {
		params["link_to_tables"] = strings.Join(linkToTables, ",")
	}

	return buildURL("/htmx/notes/edit", params)
}

func HXPostNotesEditDialog(linkToTables ...string) templ.SafeURL {
	params := make(map[string]string)

	if len(linkToTables) > 0 {
		params["link_to_tables"] = strings.Join(linkToTables, ",")
	}

	return buildURL("/htmx/notes/edit", params)
}

func HXPutNotesEditDialog(noteID models.NoteID, linkToTables ...string) templ.SafeURL {
	params := map[string]string{
		"id": fmt.Sprintf("%d", noteID),
	}

	if len(linkToTables) > 0 {
		params["link_to_tables"] = strings.Join(linkToTables, ",")
	}

	return buildURL("/htmx/notes/edit", params)
}

func HXDeleteNote(noteID models.NoteID) templ.SafeURL {
	params := map[string]string{
		"id": fmt.Sprintf("%d", noteID),
	}

	return buildURL("/htmx/notes/delete", params)
}
