package urlb

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/internal/shared"
)

// PressRegenerationPage constructs press regeneration page URL
func PressRegenerationPage(press shared.PressNumber, pressRegenerationID shared.EntityID) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/press-regeneration/%d", press))
}

// PressRegenerationDelete constructs press regeneration delete URL
func PressRegenerationDelete(press shared.PressNumber, pressRegenerationID shared.EntityID) templ.SafeURL {
	params := map[string]string{
		"id": fmt.Sprintf("%d", pressRegenerationID),
	}
	return BuildURLWithParams(fmt.Sprintf("/press-regeneration/%d/delete", press), params)
}
