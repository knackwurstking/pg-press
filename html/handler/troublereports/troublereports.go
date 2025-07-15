package troublereports

import "github.com/knackwurstking/pg-vis/pgvis"

type PageData struct {
	ID                int
	Submitted         bool // Submitted set to true will close the dialog
	Title             string
	Content           string
	LinkedAttachments []*pgvis.Attachment
	InvalidTitle      bool
	InvalidContent    bool
}
