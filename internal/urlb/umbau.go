package urlb

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/internal/shared"
)

// Umbau constructs umbau URL
func Umbau(press shared.PressNumber) templ.SafeURL {
	return BuildURL(fmt.Sprintf("/umbau/%d", press))
}
