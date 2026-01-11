package urlb

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/internal/shared"
)

func Notes() templ.SafeURL {
	return BuildURL("/notes")
}

func NotesDelete(noteID shared.EntityID) templ.SafeURL {
	return BuildURLWithParams("/notes/delete", map[string]string{
		"id": fmt.Sprintf("%d", noteID),
	})
}

func NotesGrid() templ.SafeURL {
	return BuildURL("/notes/grid")
}
