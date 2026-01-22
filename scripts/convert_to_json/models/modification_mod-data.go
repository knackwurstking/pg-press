package models

import "time"

// TroubleReportModData represents the data structure for trouble report modifications
type TroubleReportModData struct {
	Title             string         `json:"title"`
	Content           string         `json:"content"`
	LinkedAttachments []AttachmentID `json:"linked_attachments"`
	UseMarkdown       bool           `json:"use_markdown"`
}

func NewTroubleReportModData(tr *TroubleReport) *TroubleReportModData {
	return &TroubleReportModData{
		Title:             tr.Title,
		Content:           tr.Content,
		LinkedAttachments: tr.LinkedAttachments,
		UseMarkdown:       tr.UseMarkdown,
	}
}

func (trd *TroubleReportModData) CopyTo(tr *TroubleReport) *TroubleReport {
	tr.Title = trd.Title
	tr.Content = trd.Content
	tr.LinkedAttachments = trd.LinkedAttachments
	tr.UseMarkdown = trd.UseMarkdown

	return tr
}

// MetalSheetModData represents the data structure for metal sheet modifications
type MetalSheetModData struct {
	TileHeight  float64 `json:"tile_height"`
	Value       float64 `json:"value"`
	MarkeHeight int     `json:"marke_height"`
	STF         float64 `json:"stf"`
	STFMax      float64 `json:"stf_max"`
	ToolID      *int64  `json:"tool_id"`
}

// ToolModData represents the data structure for tool modifications
type ToolModData struct {
	Position     Position `json:"position"`
	Format       Format   `json:"format"`
	Type         string   `json:"type"`
	Code         string   `json:"code"`
	Regenerating bool     `json:"regenerating"`
	Press        *int     `json:"press"`
}

// PressCycleModData represents the data structure for press cycle modifications
type PressCycleModData struct {
	CycleNumber     int        `json:"cycle_number"`
	PressID         int        `json:"press_id"`
	StartTime       time.Time  `json:"start_time"`
	EndTime         *time.Time `json:"end_time,omitempty"`
	Status          string     `json:"status"`
	TotalPieces     int        `json:"total_pieces"`
	DefectivePieces int        `json:"defective_pieces"`
	Notes           string     `json:"notes,omitempty"`
}

// UserModData represents the data structure for user modifications
type UserModData struct {
	Name     string `json:"name"`
	Email    string `json:"email,omitempty"`
	IsActive bool   `json:"is_active"`
}

// NoteModData represents the data structure for note modifications
type NoteModData struct {
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	Priority    string     `json:"priority"`
	Category    string     `json:"category,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
}

// AttachmentModData represents the data structure for attachment modifications
type AttachmentModData struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	Description string `json:"description,omitempty"`
	IsPublic    bool   `json:"is_public"`
}
