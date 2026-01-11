package urlb

import "github.com/a-h/templ"

func HelpMarkdown() templ.SafeURL {
	return BuildURL("/help/markdown")
}
