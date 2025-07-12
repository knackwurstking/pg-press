package pgvis

type Attachment struct {
	Name         string `json:"name"`
	Link         string `json:"link"`
	RelativePath string `json:"relative_path"`
}

func NewAttachment(name, link, relativePath string) *Attachment {
	return &Attachment{}
}
