package urlb

import (
	"strings"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/shared"
)

func Editor(_type shared.EditorType, id string, returnURL templ.SafeURL) templ.SafeURL {
	a, _ := strings.CutPrefix(string(returnURL), env.ServerPathPrefix)
	return BuildURLWithParams("/editor", map[string]string{
		"type":       string(_type),
		"id":         id,
		"return_url": string(a),
	})
}

// UrlEditorSave constructs editor save URL
func EditorSave() templ.SafeURL {
	return BuildURL("/editor/save")
}
