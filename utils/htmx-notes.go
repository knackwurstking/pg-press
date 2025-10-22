package utils

import (
	"fmt"
	"strings"

	"github.com/a-h/templ"
)

func HXGetNotesEditDialog(noteID *int64, linkToTables ...string) templ.SafeURL {
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

func HXPutNotesEditDialog(noteID int64, linkToTables ...string) templ.SafeURL {
	params := map[string]string{
		"id": fmt.Sprintf("%d", noteID),
	}

	if len(linkToTables) > 0 {
		params["link_to_tables"] = strings.Join(linkToTables, ",")
	}

	return buildURL("/htmx/notes/edit", params)
}

func HXDeleteNote(noteID int64) templ.SafeURL {
	params := map[string]string{
		"id": fmt.Sprintf("%d", noteID),
	}

	return buildURL("/htmx/notes/delete", params)
}
