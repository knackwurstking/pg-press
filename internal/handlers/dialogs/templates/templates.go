package templates

import "github.com/a-h/templ"

const (
	DialogIDNewCassette  = "cassette-dialog"
	DialogIDEditCassette = "cassette-edit-dialog"
)

var (
	stdHxDialogProps = templ.Attributes{
		"enctype":   "multipart/form-data",
		"hx-target": "body",
		"hx-swap":   "beforeend",
	}
)

type DialogError struct {
	InputID string
	Message string
}
