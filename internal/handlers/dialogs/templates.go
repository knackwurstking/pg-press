package dialogs

import (
	"github.com/a-h/templ"
)

// Dialog IDs used for dialog triggers `dialog.Trigger`
const (
	DialogIDNewCassette  = "cassette-dialog"
	DialogIDEditCassette = "cassette-edit-dialog"
)

// Standard attributes for HTMX dialog forms.
var (
	stdHxDialogFormProps = templ.Attributes{
		"enctype":   "multipart/form-data",
		"hx-target": "body",
		"hx-swap":   "beforeend",
	}
)
