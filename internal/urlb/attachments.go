package urlb

import (
	"fmt"

	"github.com/a-h/templ"
)

func Attachment(name string) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/images/%s", name))
}
