// Package database provides feed models for tracking system events and user actions.
package database

import (
	"fmt"
	"html/template"
	"time"
)

const (
	FeedTypeUserAdd                  = "user_add"
	FeedTypeUserRemove               = "user_remove"
	FeedTypeUserNameChange           = "user_name_change"
	FeedTypeTroubleReportAdd         = "trouble_report_add"
	FeedTypeTroubleReportUpdate      = "trouble_report_update"
	FeedTypeTroubleReportRemove      = "trouble_report_remove"
	FeedTypeToolAdd                  = "tool_add"
	FeedTypeToolUpdate               = "tool_update"
	FeedTypeToolDelete               = "tool_delete"
	FeedTypeMetalSheetAdd            = "metal_sheet_add"
	FeedTypeMetalSheetUpdate         = "metal_sheet_update"
	FeedTypeMetalSheetDelete         = "metal_sheet_delete"
	FeedTypeMetalSheetStatusChange   = "metal_sheet_status_change"
	FeedTypeMetalSheetToolAssignment = "metal_sheet_tool_assignment"

	AddUserRenderTemplate = `
<div class="feed-item">
	<div
		class="feed-item-content;"
		style="padding: var(--ui-spacing);"
	>
		Benutzer <strong>%s</strong> wurde hinzugefügt.
	</div>
</div>
`

	RemoveUserRenderTemplate = `
<div class="feed-item">
	<div
		class="feed-item-content;"
		style="padding: var(--ui-spacing);"
	>
		Benutzer <strong>%s</strong> wurde entfernt.
	</div>
</div>
`

	ChangeUserNameRenderTemplate = `
<div class="feed-item">
	<div
		class="feed-item-content "
		style="padding: var(--ui-spacing);"
	>
		Benutzer <strong>%s</strong> hat den Namen zu <strong>%s</strong> geändert.
	</div>
</div>
`

	AddTroubleReportRenderTemplate = `
<div class="feed-item">
	<div
		class="feed-item-content"
		style="padding: var(--ui-spacing);"
	>
		Benutzer <strong>%s</strong> hat einen neuen Problembericht mit dem Titel
		<a href="./trouble-reports#trouble-report-%d" class="info">%s</a> hinzugefügt.
	</div>
</div>
`

	UpdateTroubleReportRenderTemplate = `
<div class="feed-item">
	<div
		class="feed-item-content"
		style="padding: var(--ui-spacing);"
	>
		Benutzer <strong>%s</strong> hat den Problembericht mit dem Titel
		<a href="./trouble-reports#trouble-report-%d" class="info">%s</a> aktualisiert.
	</div>
</div>
`

	RemoveTroubleReportRenderTemplate = `
<div class="feed-item">
	<div
		class="feed-item-content"
		style="padding: var(--ui-spacing);"
	>
		User <strong>%s<strong> hat den Problembericht mit dem Titel
		<strong style="color: var(--ui-secondary);">%s</strong> entfernt.
	</div>
</div>
`

	AddToolRenderTemplate = `
<div class="feed-item">
	<div
		class="feed-item-content"
		style="padding: var(--ui-spacing);"
	>
		Benutzer <strong>%s</strong> hat ein neues Werkzeug
		<a href="./tools/all/%d" class="info">%s</a> zur <a href="./tools/#tool-%d" class="info">Werkzeugliste</a> hinzugefügt.
	</div>
</div>
`

	UpdateToolRenderTemplate = `
<div class="feed-item">
	<div
		class="feed-item-content"
		style="padding: var(--ui-spacing);"
	>
		Benutzer <strong>%s</strong> hat das Werkzeug
		<a href="./tools/all/%d" class="info">%s</a> aktualisiert.
	</div>
</div>
`

	DeleteToolRenderTemplate = `
<div class="feed-item">
	<div
		class="feed-item-content"
		style="padding: var(--ui-spacing);"
	>
		Benutzer <strong>%s</strong> hat das Werkzeug
		<strong style="color: var(--ui-secondary);">%s</strong> entfernt.
	</div>
</div>
`

	AddMetalSheetRenderTemplate = `
<div class="feed-item">
	<div
		class="feed-item-content"
		style="padding: var(--ui-spacing);"
	>
		Benutzer <strong>%s</strong> hat ein neues Blech
		<a href="./metal-sheets/%d" class="info">%s</a> hinzugefügt.
	</div>
</div>
`

	UpdateMetalSheetRenderTemplate = `
<div class="feed-item">
	<div
		class="feed-item-content"
		style="padding: var(--ui-spacing);"
	>
		Benutzer <strong>%s</strong> hat das Blech
		<a href="./metal-sheets/%d" class="info">%s</a> aktualisiert.
	</div>
</div>
`

	DeleteMetalSheetRenderTemplate = `
<div class="feed-item">
	<div
		class="feed-item-content"
		style="padding: var(--ui-spacing);"
	>
		Benutzer <strong>%s</strong> hat das Blech mit ID
		<strong style="color: var(--ui-secondary);">%d</strong> entfernt.
	</div>
</div>
`

	MetalSheetStatusChangeRenderTemplate = `
<div class="feed-item">
	<div
		class="feed-item-content"
		style="padding: var(--ui-spacing);"
	>
		Benutzer <strong>%s</strong> hat den Status von Blech
		<a href="./metal-sheets/%d" class="info">#%d</a> zu <strong>%s</strong> geändert.
	</div>
</div>
`

	MetalSheetToolAssignmentRenderTemplate = `
<div class="feed-item">
	<div
		class="feed-item-content"
		style="padding: var(--ui-spacing);"
	>
		Benutzer <strong>%s</strong> hat Blech
		<a href="./metal-sheets/%d" class="info">#%d</a> %s.
	</div>
</div>
`
)

// FeedUserAdd represents a user addition event.
type FeedUserAdd struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func NewFeedUserAdd(data map[string]any) *FeedUserAdd {
	return &FeedUserAdd{
		ID:   int64(data["id"].(float64)),
		Name: data["name"].(string),
	}
}

func (f *FeedUserAdd) Render() template.HTML {
	return template.HTML(fmt.Sprintf(AddUserRenderTemplate, f.Name))
}

// FeedUserRemove represents a user removal event.
type FeedUserRemove struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func NewFeedUserRemove(data map[string]any) *FeedUserRemove {
	return &FeedUserRemove{
		ID:   int64(data["id"].(float64)),
		Name: data["name"].(string),
	}
}

func (f *FeedUserRemove) Render() template.HTML {
	return template.HTML(fmt.Sprintf(RemoveUserRenderTemplate, f.Name))
}

// FeedUserNameChange represents a user name change event.
type FeedUserNameChange struct {
	ID  int64  `json:"id"`
	Old string `json:"old"`
	New string `json:"new"`
}

func NewFeedUserNameChange(data map[string]any) *FeedUserNameChange {
	return &FeedUserNameChange{
		ID:  int64(data["id"].(float64)),
		Old: data["old"].(string),
		New: data["new"].(string),
	}
}

func (f *FeedUserNameChange) Render() template.HTML {
	return template.HTML(fmt.Sprintf(ChangeUserNameRenderTemplate, f.Old, f.New))
}

// FeedTroubleReportAdd represents a trouble report creation event.
type FeedTroubleReportAdd struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	ModifiedBy *User  `json:"modified_by"`
}

func NewFeedTroubleReportAdd(data map[string]any) *FeedTroubleReportAdd {
	return &FeedTroubleReportAdd{
		ID:    int64(data["id"].(float64)),
		Title: data["title"].(string),
		ModifiedBy: NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

func (f *FeedTroubleReportAdd) Render() template.HTML {
	return template.HTML(fmt.Sprintf(AddTroubleReportRenderTemplate,
		f.ModifiedBy.UserName, f.ID, f.Title))
}

// FeedTroubleReportUpdate represents a trouble report update event.
type FeedTroubleReportUpdate struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	ModifiedBy *User  `json:"modified_by"`
}

func NewFeedTroubleReportUpdate(data map[string]any) *FeedTroubleReportUpdate {
	return &FeedTroubleReportUpdate{
		ID:    int64(data["id"].(float64)),
		Title: data["title"].(string),
		ModifiedBy: NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

func (f *FeedTroubleReportUpdate) Render() template.HTML {
	return template.HTML(fmt.Sprintf(
		UpdateTroubleReportRenderTemplate,
		f.ModifiedBy.UserName, f.ID, f.Title,
	))
}

// FeedTroubleReportRemove represents a trouble report removal event.
type FeedTroubleReportRemove struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	RemovedBy *User  `json:"removed_by"`
}

func NewFeedTroubleReportRemove(data map[string]any) *FeedTroubleReportRemove {
	return &FeedTroubleReportRemove{
		ID:    int64(data["id"].(float64)),
		Title: data["title"].(string),
		RemovedBy: NewUserFromInterfaceMap(
			data["removed_by"].(map[string]any),
		),
	}
}

func (f *FeedTroubleReportRemove) Render() template.HTML {
	return template.HTML(
		fmt.Sprintf(RemoveTroubleReportRenderTemplate, f.RemovedBy.UserName, f.Title),
	)
}

// FeedToolAdd represents a tool addition event.
type FeedToolAdd struct {
	ID         int64  `json:"id"`
	Tool       string `json:"tool"`
	ModifiedBy *User  `json:"modified_by"`
}

func NewFeedToolAdd(data map[string]any) *FeedToolAdd {
	return &FeedToolAdd{
		ID:   int64(data["id"].(float64)),
		Tool: data["tool"].(string),
		ModifiedBy: NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

func (f *FeedToolAdd) Render() template.HTML {
	return template.HTML(fmt.Sprintf(AddToolRenderTemplate,
		f.ModifiedBy.UserName, f.ID, f.Tool, f.ID))
}

// FeedToolUpdate represents a tool update event.
type FeedToolUpdate struct {
	ID         int64  `json:"id"`
	Tool       string `json:"tool"`
	ModifiedBy *User  `json:"modified_by"`
}

func NewFeedToolUpdate(data map[string]any) *FeedToolUpdate {
	return &FeedToolUpdate{
		ID:   int64(data["id"].(float64)),
		Tool: data["tool"].(string),
		ModifiedBy: NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

func (f *FeedToolUpdate) Render() template.HTML {
	return template.HTML(fmt.Sprintf(UpdateToolRenderTemplate,
		f.ModifiedBy.UserName, f.ID, f.Tool))
}

// FeedToolDelete represents a tool deletion event.
type FeedToolDelete struct {
	ID         int64  `json:"id"`
	Tool       string `json:"tool"`
	ModifiedBy *User  `json:"modified_by"`
}

func NewFeedToolDelete(data map[string]any) *FeedToolDelete {
	return &FeedToolDelete{
		ID:   int64(data["id"].(float64)),
		Tool: data["tool"].(string),
		ModifiedBy: NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

func (f *FeedToolDelete) Render() template.HTML {
	return template.HTML(fmt.Sprintf(DeleteToolRenderTemplate,
		f.ModifiedBy.UserName, f.Tool))
}

// FeedMetalSheetAdd represents a metal sheet addition event.
type FeedMetalSheetAdd struct {
	ID         int64  `json:"id"`
	MetalSheet string `json:"metal_sheet"`
	ModifiedBy *User  `json:"modified_by"`
}

func NewFeedMetalSheetAdd(data map[string]any) *FeedMetalSheetAdd {
	return &FeedMetalSheetAdd{
		ID:         int64(data["id"].(float64)),
		MetalSheet: data["metal_sheet"].(string),
		ModifiedBy: NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

func (f *FeedMetalSheetAdd) Render() template.HTML {
	return template.HTML(fmt.Sprintf(AddMetalSheetRenderTemplate,
		f.ModifiedBy.UserName, f.ID, f.MetalSheet))
}

// FeedMetalSheetUpdate represents a metal sheet update event.
type FeedMetalSheetUpdate struct {
	ID         int64  `json:"id"`
	MetalSheet string `json:"metal_sheet"`
	ModifiedBy *User  `json:"modified_by"`
}

func NewFeedMetalSheetUpdate(data map[string]any) *FeedMetalSheetUpdate {
	return &FeedMetalSheetUpdate{
		ID:         int64(data["id"].(float64)),
		MetalSheet: data["metal_sheet"].(string),
		ModifiedBy: NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

func (f *FeedMetalSheetUpdate) Render() template.HTML {
	return template.HTML(fmt.Sprintf(UpdateMetalSheetRenderTemplate,
		f.ModifiedBy.UserName, f.ID, f.MetalSheet))
}

// FeedMetalSheetDelete represents a metal sheet deletion event.
type FeedMetalSheetDelete struct {
	ID         int64 `json:"id"`
	ModifiedBy *User `json:"modified_by"`
}

func NewFeedMetalSheetDelete(data map[string]any) *FeedMetalSheetDelete {
	return &FeedMetalSheetDelete{
		ID: int64(data["id"].(float64)),
		ModifiedBy: NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

func (f *FeedMetalSheetDelete) Render() template.HTML {
	return template.HTML(fmt.Sprintf(DeleteMetalSheetRenderTemplate,
		f.ModifiedBy.UserName, f.ID))
}

// FeedMetalSheetStatusChange represents a metal sheet status change event.
type FeedMetalSheetStatusChange struct {
	ID         int64  `json:"id"`
	NewStatus  string `json:"new_status"`
	ModifiedBy *User  `json:"modified_by"`
}

func NewFeedMetalSheetStatusChange(data map[string]any) *FeedMetalSheetStatusChange {
	return &FeedMetalSheetStatusChange{
		ID:        int64(data["id"].(float64)),
		NewStatus: data["new_status"].(string),
		ModifiedBy: NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

func (f *FeedMetalSheetStatusChange) Render() template.HTML {
	statusTranslation := map[string]string{
		"available":   "Verfügbar",
		"in_use":      "In Verwendung",
		"maintenance": "Wartung",
		"reserved":    "Reserviert",
		"damaged":     "Beschädigt",
	}
	status := statusTranslation[f.NewStatus]
	if status == "" {
		status = f.NewStatus
	}
	return template.HTML(fmt.Sprintf(MetalSheetStatusChangeRenderTemplate,
		f.ModifiedBy.UserName, f.ID, f.ID, status))
}

// FeedMetalSheetToolAssignment represents a metal sheet tool assignment event.
type FeedMetalSheetToolAssignment struct {
	SheetID    int64  `json:"sheet_id"`
	ToolID     *int64 `json:"tool_id"`
	ModifiedBy *User  `json:"modified_by"`
}

func NewFeedMetalSheetToolAssignment(data map[string]any) *FeedMetalSheetToolAssignment {
	assignment := &FeedMetalSheetToolAssignment{
		SheetID: int64(data["sheet_id"].(float64)),
		ModifiedBy: NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
	if toolID, ok := data["tool_id"]; ok && toolID != nil {
		id := int64(toolID.(float64))
		assignment.ToolID = &id
	}
	return assignment
}

func (f *FeedMetalSheetToolAssignment) Render() template.HTML {
	action := "vom Werkzeug getrennt"
	if f.ToolID != nil {
		action = fmt.Sprintf("dem Werkzeug <a href=\"./tools/all/%d\" class=\"info\">#%d</a> zugewiesen", *f.ToolID, *f.ToolID)
	}
	return template.HTML(fmt.Sprintf(MetalSheetToolAssignmentRenderTemplate,
		f.ModifiedBy.UserName, f.SheetID, f.SheetID, action))
}

// Feed represents a feed entry in the system that tracks activity events.
type Feed struct {
	ID       int64  `json:"id"`
	Time     int64  `json:"time"`
	DataType string `json:"data_type"`
	Data     any    `json:"data"`
}

// NewFeed creates a new feed entry with the current timestamp.
func NewFeed(dataType string, data any) *Feed {
	return &Feed{
		Time:     time.Now().UnixMilli(),
		DataType: dataType,
		Data:     data,
	}
}

// Render generates HTML for the feed entry.
func (f *Feed) Render() template.HTML {
	timeStr := f.GetTime().Format("2006-01-02 15:04:05")
	var feedContent template.HTML

	data, _ := f.Data.(map[string]any)

	switch f.DataType {

	// User Types

	case FeedTypeUserAdd:
		feedContent = NewFeedUserAdd(data).Render()
	case FeedTypeUserNameChange:
		feedContent = NewFeedUserNameChange(data).Render()
	case FeedTypeUserRemove:
		feedContent = NewFeedUserRemove(data).Render()

	// Trouble Report Types

	case FeedTypeTroubleReportAdd:
		feedContent = NewFeedTroubleReportAdd(data).Render()
	case FeedTypeTroubleReportUpdate:
		feedContent = NewFeedTroubleReportUpdate(data).Render()
	case FeedTypeTroubleReportRemove:
		feedContent = NewFeedTroubleReportRemove(data).Render()

	// Tool Types

	case FeedTypeToolAdd:
		feedContent = NewFeedToolAdd(data).Render()
	case FeedTypeToolUpdate:
		feedContent = NewFeedToolUpdate(data).Render()
	case FeedTypeToolDelete:
		feedContent = NewFeedToolDelete(data).Render()

	// Metal Sheet Types

	case FeedTypeMetalSheetAdd:
		feedContent = NewFeedMetalSheetAdd(data).Render()
	case FeedTypeMetalSheetUpdate:
		feedContent = NewFeedMetalSheetUpdate(data).Render()
	case FeedTypeMetalSheetDelete:
		feedContent = NewFeedMetalSheetDelete(data).Render()
	case FeedTypeMetalSheetStatusChange:
		feedContent = NewFeedMetalSheetStatusChange(data).Render()
	case FeedTypeMetalSheetToolAssignment:
		feedContent = NewFeedMetalSheetToolAssignment(data).Render()

	// Fallback

	default:
		return template.HTML(fmt.Sprintf(
			`
			<article id="feed-%d" class="card" data-id="%d" data-time="%d">
				<div style="margin-bottom: var(--ui-spacing);" class="card-body"><pre>%#v</pre></div>
				<div class="card-footer">
					<small style="float: right;">%s</small>
				</div>
			</article>`,
			f.ID, f.ID, f.Time, f.Data, timeStr,
		))
	}

	return template.HTML(fmt.Sprintf(`
		<article id="feed-%d" class="card" data-id="%d" data-time="%d">
			<div style="margin-bottom: var(--ui-spacing);" class="card-body">%s</div>
			<div class="card-footer">
				<small style="float: right;">%s</small>
			</div>
		</article>`,
		f.ID, f.ID, f.Time, feedContent, timeStr,
	))
}

// Validate checks if the feed has valid data.
func (f *Feed) Validate() error {
	if f.Data == nil {
		return NewValidationError("cache", "cannot be nil", f.Data)
	}
	if f.DataType == "" {
		return NewValidationError("data type", "cannot be empty", f.DataType)
	}
	if f.Time <= 0 {
		return NewValidationError("time", "must be positive", f.Time)
	}
	return nil
}

// GetTime returns the feed time as a Go time.Time.
func (f *Feed) GetTime() time.Time {
	return time.UnixMilli(f.Time)
}

// Age returns the duration since the feed was created.
func (f *Feed) Age() time.Duration {
	return time.Since(f.GetTime())
}

// IsOlderThan checks if the feed is older than the specified duration.
func (f *Feed) IsOlderThan(duration time.Duration) bool {
	return f.Age() > duration
}

// String returns a string representation of the feed.
func (f *Feed) String() string {
	return fmt.Sprintf("Feed{ID: %d, Time: %s, Cache: %#v}",
		f.ID, f.GetTime().Format("2006-01-02 15:04:05"), f.Data)
}

// Clone creates a copy of the feed.
func (f *Feed) Clone() *Feed {
	return &Feed{
		ID:   f.ID,
		Time: f.Time,
		Data: f.Data,
	}
}
