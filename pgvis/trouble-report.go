package pgvis

import "encoding/json"

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

func (tr *TroubleReport) LinkedAttachmentsToJSON() []byte {
	data, err := json.Marshal(tr)
	if err != nil {
		panic(err)
	}
	return data
}

func (tr *TroubleReport) JSONToLinkedAttachments(b []byte) {
	err := json.Unmarshal(b, tr.LinkedAttachments)
	if err != nil {
		panic(err)
	}
}
