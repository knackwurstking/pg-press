// TODO: I really need to refactor this, Find a better way to handle feeds if possible
package feed

import (
	"fmt"
	"time"

	"github.com/knackwurstking/pgpress/internal/database/dberror"
	"github.com/knackwurstking/pgpress/internal/database/models/user"
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
	FeedTypePressCycleAdd            = "press_cycle_add"
	FeedTypePressCycleUpdate         = "press_cycle_update"
	FeedTypePressCycleDelete         = "press_cycle_delete"
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

// FeedTroubleReportAdd represents a trouble report creation event.
type FeedTroubleReportAdd struct {
	ID         int64      `json:"id"`
	Title      string     `json:"title"`
	ModifiedBy *user.User `json:"modified_by"`
}

func NewFeedTroubleReportAdd(data map[string]any) *FeedTroubleReportAdd {
	return &FeedTroubleReportAdd{
		ID:    int64(data["id"].(float64)),
		Title: data["title"].(string),
		ModifiedBy: user.NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

// FeedTroubleReportUpdate represents a trouble report update event.
type FeedTroubleReportUpdate struct {
	ID         int64      `json:"id"`
	Title      string     `json:"title"`
	ModifiedBy *user.User `json:"modified_by"`
}

func NewFeedTroubleReportUpdate(data map[string]any) *FeedTroubleReportUpdate {
	return &FeedTroubleReportUpdate{
		ID:    int64(data["id"].(float64)),
		Title: data["title"].(string),
		ModifiedBy: user.NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

// FeedTroubleReportRemove represents a trouble report removal event.
type FeedTroubleReportRemove struct {
	ID        int64      `json:"id"`
	Title     string     `json:"title"`
	RemovedBy *user.User `json:"removed_by"`
}

func NewFeedTroubleReportRemove(data map[string]any) *FeedTroubleReportRemove {
	return &FeedTroubleReportRemove{
		ID:    int64(data["id"].(float64)),
		Title: data["title"].(string),
		RemovedBy: user.NewUserFromInterfaceMap(
			data["removed_by"].(map[string]any),
		),
	}
}

// FeedToolAdd represents a tool addition event.
type FeedToolAdd struct {
	ID         int64      `json:"id"`
	Tool       string     `json:"tool"`
	ModifiedBy *user.User `json:"modified_by"`
}

func NewFeedToolAdd(data map[string]any) *FeedToolAdd {
	return &FeedToolAdd{
		ID:   int64(data["id"].(float64)),
		Tool: data["tool"].(string),
		ModifiedBy: user.NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

// FeedToolUpdate represents a tool update event.
type FeedToolUpdate struct {
	ID         int64      `json:"id"`
	Tool       string     `json:"tool"`
	ModifiedBy *user.User `json:"modified_by"`
}

func NewFeedToolUpdate(data map[string]any) *FeedToolUpdate {
	return &FeedToolUpdate{
		ID:   int64(data["id"].(float64)),
		Tool: data["tool"].(string),
		ModifiedBy: user.NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

// FeedToolDelete represents a tool deletion event.
type FeedToolDelete struct {
	ID         int64      `json:"id"`
	Tool       string     `json:"tool"`
	ModifiedBy *user.User `json:"modified_by"`
}

func NewFeedToolDelete(data map[string]any) *FeedToolDelete {
	return &FeedToolDelete{
		ID:   int64(data["id"].(float64)),
		Tool: data["tool"].(string),
		ModifiedBy: user.NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

// FeedMetalSheetAdd represents a metal sheet addition event.
type FeedMetalSheetAdd struct {
	ID         int64      `json:"id"`
	MetalSheet string     `json:"metal_sheet"`
	ModifiedBy *user.User `json:"modified_by"`
}

func NewFeedMetalSheetAdd(data map[string]any) *FeedMetalSheetAdd {
	return &FeedMetalSheetAdd{
		ID:         int64(data["id"].(float64)),
		MetalSheet: data["metal_sheet"].(string),
		ModifiedBy: user.NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

// FeedMetalSheetUpdate represents a metal sheet update event.
type FeedMetalSheetUpdate struct {
	ID         int64      `json:"id"`
	MetalSheet string     `json:"metal_sheet"`
	ModifiedBy *user.User `json:"modified_by"`
}

func NewFeedMetalSheetUpdate(data map[string]any) *FeedMetalSheetUpdate {
	return &FeedMetalSheetUpdate{
		ID:         int64(data["id"].(float64)),
		MetalSheet: data["metal_sheet"].(string),
		ModifiedBy: user.NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

// FeedMetalSheetDelete represents a metal sheet deletion event.
type FeedMetalSheetDelete struct {
	ID         int64      `json:"id"`
	ModifiedBy *user.User `json:"modified_by"`
}

func NewFeedMetalSheetDelete(data map[string]any) *FeedMetalSheetDelete {
	return &FeedMetalSheetDelete{
		ID: int64(data["id"].(float64)),
		ModifiedBy: user.NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

// FeedMetalSheetStatusChange represents a metal sheet status change event.
type FeedMetalSheetStatusChange struct {
	ID         int64      `json:"id"`
	NewStatus  string     `json:"new_status"`
	ModifiedBy *user.User `json:"modified_by"`
}

func NewFeedMetalSheetStatusChange(data map[string]any) *FeedMetalSheetStatusChange {
	return &FeedMetalSheetStatusChange{
		ID:        int64(data["id"].(float64)),
		NewStatus: data["new_status"].(string),
		ModifiedBy: user.NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

// FeedMetalSheetToolAssignment represents a metal sheet tool assignment event.
type FeedMetalSheetToolAssignment struct {
	SheetID    int64      `json:"sheet_id"`
	ToolID     *int64     `json:"tool_id"`
	ModifiedBy *user.User `json:"modified_by"`
}

func NewFeedMetalSheetToolAssignment(data map[string]any) *FeedMetalSheetToolAssignment {
	assignment := &FeedMetalSheetToolAssignment{
		SheetID: int64(data["sheet_id"].(float64)),
		ModifiedBy: user.NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
	if toolID, ok := data["tool_id"]; ok && toolID != nil {
		id := int64(toolID.(float64))
		assignment.ToolID = &id
	}
	return assignment
}

// FeedPressCycleAdd represents a press cycle creation event.
type FeedPressCycleAdd struct {
	SlotTop         int64      `json:"slot_top"`
	SlotTopCassette int64      `json:"slot_top_cassette"`
	SlotBottom      int64      `json:"slot_bottom"`
	TotalCycles     int64      `json:"total_cycles"`
	ModifiedBy      *user.User `json:"modified_by"`
}

func NewFeedPressCycleAdd(data map[string]any) *FeedPressCycleAdd {
	return &FeedPressCycleAdd{
		SlotTop:         int64(data["slot_top"].(float64)),
		SlotTopCassette: int64(data["slot_top_cassette"].(float64)),
		SlotBottom:      int64(data["slot_bottom"].(float64)),
		TotalCycles:     int64(data["total_cycles"].(float64)),
		ModifiedBy: user.NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

// FeedPressCycleUpdate represents a press cycle update event.
type FeedPressCycleUpdate struct {
	SlotTop         int64      `json:"slot_top"`
	SlotTopCassette int64      `json:"slot_top_cassette"`
	SlotBottom      int64      `json:"slot_bottom"`
	TotalCycles     int64      `json:"total_cycles"`
	ModifiedBy      *user.User `json:"modified_by"`
}

func NewFeedPressCycleUpdate(data map[string]any) *FeedPressCycleUpdate {
	return &FeedPressCycleUpdate{
		SlotTop:         int64(data["slot_top"].(float64)),
		SlotTopCassette: int64(data["slot_top_cassette"].(float64)),
		SlotBottom:      int64(data["slot_bottom"].(float64)),
		TotalCycles:     int64(data["total_cycles"].(float64)),
		ModifiedBy: user.NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
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

// Validate checks if the feed has valid data.
func (f *Feed) Validate() error {
	if f.Data == nil {
		return dberror.NewValidationError("cache", "cannot be nil", f.Data)
	}
	if f.DataType == "" {
		return dberror.NewValidationError("data type", "cannot be empty", f.DataType)
	}
	if f.Time <= 0 {
		return dberror.NewValidationError("time", "must be positive", f.Time)
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
