package utils

import (
	"fmt"
	"strings"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pgpress/env"
)

func HXGetNotesEditDialog(noteID *int64, linkToTables ...string) templ.SafeURL {
	var params []string

	if noteID != nil {
		params = append(params, fmt.Sprintf("id=%d", *noteID))
	}

	if len(linkToTables) > 0 {
		params = append(params, fmt.Sprintf(
			"link_to_tables=%s", strings.Join(linkToTables, ","),
		))
	}

	url := fmt.Sprintf("%s/htmx/notes/edit", env.ServerPathPrefix)
	if len(params) > 0 {
		url += "?" + strings.Join(params, "&")
	}

	return templ.SafeURL(url)
}

func HXPostNotesEditDialog(linkToTables ...string) templ.SafeURL {
	if len(linkToTables) == 0 {
		return templ.SafeURL(fmt.Sprintf("%s/htmx/notes/edit", env.ServerPathPrefix))
	}

	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/notes/edit?link_to_tables=%s",
		env.ServerPathPrefix, strings.Join(linkToTables, ","),
	))
}

func HXPutNotesEditDialog(noteID int64, linkToTables ...string) templ.SafeURL {
	if len(linkToTables) == 0 {
		return templ.SafeURL(fmt.Sprintf("%s/htmx/notes/edit?id=%d", env.ServerPathPrefix, noteID))
	}

	return templ.SafeURL(fmt.Sprintf(
		"%s/htmx/notes/edit?id=%d&link_to_tables=%s",
		env.ServerPathPrefix, noteID, strings.Join(linkToTables, ","),
	))
}

func HXDeleteNote(noteID int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf("%s/htmx/notes/delete?id=%d", env.ServerPathPrefix, noteID))
}
