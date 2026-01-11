package logger

import (
	"github.com/knackwurstking/ui/ui-templ"
)

func New(group string) *ui.Logger {
	return ui.NewLogger(group)
}
