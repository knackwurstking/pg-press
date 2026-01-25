package models

// CREATE TABLE IF NOT EXISTS trouble_reports (
// 		id INTEGER NOT NULL,
// 		title TEXT NOT NULL,
// 		content TEXT NOT NULL,
// 		linked_attachments TEXT NOT NULL,
// 		use_markdown BOOLEAN DEFAULT 0,
// 		PRIMARY KEY("id" AUTOINCREMENT)
// 	);

type TroubleReport struct {
	ID                   int64    `json:"id"`
	Title                string   `json:"title"`
	Content              string   `json:"content"`
	LinkedAttachments    []int64  `json:"linked_attachments"`
	NewLinkedAttachments []string `json:"new_linked_attachments"`
	UseMarkdown          bool     `json:"use_markdown"`
}
