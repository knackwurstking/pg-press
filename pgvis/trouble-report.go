package pgvis

type TroubleReport struct {
	ID                int                       `json:"id"`
	Title             string                    `json:"title"`
	Content           string                    `json:"content"`
	LinkedAttachments []*Attachment             `json:"linked_attachments"`
	Modified          *Modified[*TroubleReport] `json:"modified"`
}

func NewTroubleReport(m *Modified[*TroubleReport], title, content string) *TroubleReport {
	return &TroubleReport{
		Title:             title,
		Content:           content,
		LinkedAttachments: make([]*Attachment, 0),
		Modified:          m,
	}
}
