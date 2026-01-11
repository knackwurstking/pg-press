package urlb

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/internal/shared"
)

// UmbauPage constructs umbau URL
func UmbauPage(press shared.PressNumber) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/umbau/%d", press))
}
