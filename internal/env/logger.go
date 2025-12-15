package env

import (
	"github.com/knackwurstking/ui/ui-templ"
)

func NewLogger(group string) *ui.Logger {
	if Verbose {
		return ui.NewLogger(group)
	}

	return ui.NewLoggerWithVerbose(group)
}
