// Package pgvis provides feed models for tracking system events and user actions.
package pgvis

import (
	"fmt"
	"html/template"
	"time"
)

const (
	FeedTypeUserAdd             = "user_add"
	FeedTypeUserRemove          = "user_remove"
	FeedTypeUserNameChange      = "user_name_change"
	FeedTypeTroubleReportAdd    = "trouble_report_add"
	FeedTypeTroubleReportUpdate = "trouble_report_update"
	FeedTypeTroubleReportRemove = "trouble_report_remove"

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
		<a href="./trouble-reports#trouble-report-%d">%s</a> hinzugefügt.
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
		<a href="./trouble-reports#trouble-report-%d">%s</a> aktualisiert.
	</div>
</div>
`

	RemoveTroubleReportRenderTemplate = `
<div class="feed-item">
	<div
		class="feed-item-content"
		style="padding: var(--ui-spacing);"
	>
		Benutzer <strong>%s</strong> hat den Problembericht mit dem Titel
		<strong>%s</strong> entfernt.
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
	return template.HTML(fmt.Sprintf(AddTroubleReportRenderTemplate, f.ModifiedBy.UserName, f.ID, f.Title))
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
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	ModifiedBy *User  `json:"modified_by"`
}

func NewFeedTroubleReportRemove(data map[string]any) *FeedTroubleReportRemove {
	return &FeedTroubleReportRemove{
		ID:    int64(data["id"].(float64)),
		Title: data["title"].(string),
		ModifiedBy: NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

func (f *FeedTroubleReportRemove) Render() template.HTML {
	return template.HTML(fmt.Sprintf(
		RemoveTroubleReportRenderTemplate,
		f.ModifiedBy.UserName, f.Title,
	))
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

// NewFeedWithTime creates a new feed entry with a specific timestamp.
func NewFeedWithTime(dataType string, data any, timestamp int64) *Feed {
	return &Feed{
		Time:     timestamp,
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
