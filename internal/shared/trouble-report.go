package shared

type TroubleReport struct {
	ID                EntityID `json:"id"`
	Title             string   `json:"title"`
	Content           string   `json:"content"`
	LinkedAttachments []*Image `json:"linked_attachments"`
	UseMarkdown       bool     `json:"use_markdown"`
}
