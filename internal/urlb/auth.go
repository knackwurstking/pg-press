package urlb

import (
	"fmt"

	"github.com/a-h/templ"
)

func Login(apiKey string, invalid *bool) templ.SafeURL {
	params := map[string]string{}
	if apiKey != "" {
		params["api-key"] = apiKey
	}
	if invalid != nil {
		params["invalid"] = fmt.Sprintf("%t", *invalid)
	}
	return BuildURLWithParams("/login", params)
}
