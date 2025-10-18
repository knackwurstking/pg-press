package models

import "time"

// TroubleReportModData represents the data structure for trouble report modifications
type TroubleReportModData struct {
	Title             string  `json:"title"`
	Content           string  `json:"content"`
	LinkedAttachments []int64 `json:"linked_attachments"`
	UseMarkdown       bool    `json:"use_markdown"`
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

// ModificationAction represents the type of action performed
type ModificationAction string

const (
	ActionCreate ModificationAction = "create"
	ActionUpdate ModificationAction = "update"
	ActionDelete ModificationAction = "delete"
	ActionAssign ModificationAction = "assign"
	ActionRemove ModificationAction = "remove"
)

// ModificationContext provides additional context about the modification
type ModificationContext struct {
	Action      ModificationAction `json:"action"`
	Description string             `json:"description,omitempty"`
	IPAddress   string             `json:"ip_address,omitempty"`
	UserAgent   string             `json:"user_agent,omitempty"`
	SessionID   string             `json:"session_id,omitempty"`
}

// ExtendedModificationData wraps the actual data with context information
type ExtendedModificationData[T any] struct {
	Data    T                   `json:"data"`
	Context ModificationContext `json:"context"`
}

// NewExtendedModificationData creates a new extended modification data with context
func NewExtendedModificationData[T any](data T, action ModificationAction, description string) *ExtendedModificationData[T] {
	return &ExtendedModificationData[T]{
		Data: data,
		Context: ModificationContext{
			Action:      action,
			Description: description,
		},
	}
}

// WithIPAddress adds IP address to the context
func (e *ExtendedModificationData[T]) WithIPAddress(ip string) *ExtendedModificationData[T] {
	e.Context.IPAddress = ip
	return e
}

// WithUserAgent adds user agent to the context
func (e *ExtendedModificationData[T]) WithUserAgent(userAgent string) *ExtendedModificationData[T] {
	e.Context.UserAgent = userAgent
	return e
}

// WithSessionID adds session ID to the context
func (e *ExtendedModificationData[T]) WithSessionID(sessionID string) *ExtendedModificationData[T] {
	e.Context.SessionID = sessionID
	return e
}

// GetAction returns the action type
func (e *ExtendedModificationData[T]) GetAction() ModificationAction {
	return e.Context.Action
}

// GetDescription returns the description
func (e *ExtendedModificationData[T]) GetDescription() string {
	return e.Context.Description
}

// IsAction checks if the modification has the specified action
func (e *ExtendedModificationData[T]) IsAction(action ModificationAction) bool {
	return e.Context.Action == action
}

// IsCreate checks if this is a create action
func (e *ExtendedModificationData[T]) IsCreate() bool {
	return e.IsAction(ActionCreate)
}

// IsUpdate checks if this is an update action
func (e *ExtendedModificationData[T]) IsUpdate() bool {
	return e.IsAction(ActionUpdate)
}

// IsDelete checks if this is a delete action
func (e *ExtendedModificationData[T]) IsDelete() bool {
	return e.IsAction(ActionDelete)
}
