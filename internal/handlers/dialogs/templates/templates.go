package templates

import "github.com/a-h/templ"

// Dialog IDs used for dialog triggers `dialog.Trigger`
const (
	DialogIDNewCassette  = "cassette-dialog"
	DialogIDEditCassette = "cassette-edit-dialog"
)

// Standard attributes for HTMX dialog forms.
var (
	stdHxDialogProps = templ.Attributes{
		"enctype":   "multipart/form-data",
		"hx-target": "body",
		"hx-swap":   "beforeend",
	}
)

// DialogError represents an error that occurred during dialog processing, such as form validation errors.
type DialogError struct {
	InputID string
	Message string
}
