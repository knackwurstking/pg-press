package logger

import (
	"github.com/knackwurstking/pg-press/internal/env"

	"github.com/knackwurstking/ui/pkg/ui"
)

func New(group string) *ui.Logger {
	if env.Verbose {
		return ui.NewLoggerWithVerbose(group)
	}
	return ui.NewLogger(group)
}
