// TODO: I really need to refactor this, Find a better way to handle feeds if possible
package feed

import (
	"fmt"
	"time"

	"github.com/knackwurstking/pgpress/internal/database/dberror"
	"github.com/knackwurstking/pgpress/internal/database/models/user"
)

const (
	TypeUserAdd                  = "user_add"
	TypeUserRemove               = "user_remove"
	TypeUserNameChange           = "user_name_change"
	TypeTroubleReportAdd         = "trouble_report_add"
	TypeTroubleReportUpdate      = "trouble_report_update"
	TypeTroubleReportRemove      = "trouble_report_remove"
	TypeToolAdd                  = "tool_add"
	TypeToolUpdate               = "tool_update"
	TypeToolDelete               = "tool_delete"
	TypeMetalSheetAdd            = "metal_sheet_add"
	TypeMetalSheetUpdate         = "metal_sheet_update"
	TypeMetalSheetDelete         = "metal_sheet_delete"
	TypeMetalSheetStatusChange   = "metal_sheet_status_change"
	TypeMetalSheetToolAssignment = "metal_sheet_tool_assignment"
	TypePressCycleAdd            = "press_cycle_add"
	TypePressCycleUpdate         = "press_cycle_update"
	TypePressCycleDelete         = "press_cycle_delete"

	TypeRegenerationAdd    = "regeneration_add"
	TypeRegenerationUpdate = "regeneration_update"
	TypeRegenerationDelete = "regeneration_delete"
)

// Feed represents a feed entry in the system that tracks activity events.
type Feed struct {
	ID       int64  `json:"id"`
	Time     int64  `json:"time"`
	DataType string `json:"data_type"`
	Data     any    `json:"data"`
}

// New creates a new feed entry with the current timestamp.
func New(dataType string, data any) *Feed {
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

/*************
 * Data Types
 *************/

// UserAdd represents a user addition event.
type UserAdd struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func NewUserAdd(data map[string]any) *UserAdd {
	return &UserAdd{
		ID:   int64(data["id"].(float64)),
		Name: data["name"].(string),
	}
}

// UserRemove represents a user removal event.
type UserRemove struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func NewUserRemove(data map[string]any) *UserRemove {
	return &UserRemove{
		ID:   int64(data["id"].(float64)),
		Name: data["name"].(string),
	}
}

// UserNameChange represents a user name change event.
type UserNameChange struct {
	ID  int64  `json:"id"`
	Old string `json:"old"`
	New string `json:"new"`
}

func NewUserNameChange(data map[string]any) *UserNameChange {
	return &UserNameChange{
		ID:  int64(data["id"].(float64)),
		Old: data["old"].(string),
		New: data["new"].(string),
	}
}

// TroubleReportAdd represents a trouble report creation event.
type TroubleReportAdd struct {
	ID         int64      `json:"id"`
	Title      string     `json:"title"`
	ModifiedBy *user.User `json:"modified_by"`
}

func NewTroubleReportAdd(data map[string]any) *TroubleReportAdd {
	return &TroubleReportAdd{
		ID:    int64(data["id"].(float64)),
		Title: data["title"].(string),
		ModifiedBy: user.NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

// TroubleReportUpdate represents a trouble report update event.
type TroubleReportUpdate struct {
	ID         int64      `json:"id"`
	Title      string     `json:"title"`
	ModifiedBy *user.User `json:"modified_by"`
}

func NewTroubleReportUpdate(data map[string]any) *TroubleReportUpdate {
	return &TroubleReportUpdate{
		ID:    int64(data["id"].(float64)),
		Title: data["title"].(string),
		ModifiedBy: user.NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

// TroubleReportRemove represents a trouble report removal event.
type TroubleReportRemove struct {
	ID        int64      `json:"id"`
	Title     string     `json:"title"`
	RemovedBy *user.User `json:"removed_by"`
}

func NewTroubleReportRemove(data map[string]any) *TroubleReportRemove {
	return &TroubleReportRemove{
		ID:    int64(data["id"].(float64)),
		Title: data["title"].(string),
		RemovedBy: user.NewUserFromInterfaceMap(
			data["removed_by"].(map[string]any),
		),
	}
}

// ToolAdd represents a tool addition event.
type ToolAdd struct {
	ID         int64      `json:"id"`
	Tool       string     `json:"tool"`
	ModifiedBy *user.User `json:"modified_by"`
}

func NewToolAdd(data map[string]any) *ToolAdd {
	return &ToolAdd{
		ID:   int64(data["id"].(float64)),
		Tool: data["tool"].(string),
		ModifiedBy: user.NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

// ToolUpdate represents a tool update event.
type ToolUpdate struct {
	ID         int64      `json:"id"`
	Tool       string     `json:"tool"`
	ModifiedBy *user.User `json:"modified_by"`
}

func NewToolUpdate(data map[string]any) *ToolUpdate {
	return &ToolUpdate{
		ID:   int64(data["id"].(float64)),
		Tool: data["tool"].(string),
		ModifiedBy: user.NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

// ToolDelete represents a tool deletion event.
type ToolDelete struct {
	ID         int64      `json:"id"`
	Tool       string     `json:"tool"`
	ModifiedBy *user.User `json:"modified_by"`
}

func NewToolDelete(data map[string]any) *ToolDelete {
	return &ToolDelete{
		ID:   int64(data["id"].(float64)),
		Tool: data["tool"].(string),
		ModifiedBy: user.NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

// MetalSheetAdd represents a metal sheet addition event.
type MetalSheetAdd struct {
	ID         int64      `json:"id"`
	MetalSheet string     `json:"metal_sheet"`
	ModifiedBy *user.User `json:"modified_by"`
}

func NewMetalSheetAdd(data map[string]any) *MetalSheetAdd {
	return &MetalSheetAdd{
		ID:         int64(data["id"].(float64)),
		MetalSheet: data["metal_sheet"].(string),
		ModifiedBy: user.NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

// MetalSheetUpdate represents a metal sheet update event.
type MetalSheetUpdate struct {
	ID         int64      `json:"id"`
	MetalSheet string     `json:"metal_sheet"`
	ModifiedBy *user.User `json:"modified_by"`
}

func NewMetalSheetUpdate(data map[string]any) *MetalSheetUpdate {
	return &MetalSheetUpdate{
		ID:         int64(data["id"].(float64)),
		MetalSheet: data["metal_sheet"].(string),
		ModifiedBy: user.NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

// MetalSheetDelete represents a metal sheet deletion event.
type MetalSheetDelete struct {
	ID         int64      `json:"id"`
	ModifiedBy *user.User `json:"modified_by"`
}

func NewMetalSheetDelete(data map[string]any) *MetalSheetDelete {
	return &MetalSheetDelete{
		ID: int64(data["id"].(float64)),
		ModifiedBy: user.NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

// MetalSheetStatusChange represents a metal sheet status change event.
type MetalSheetStatusChange struct {
	ID         int64      `json:"id"`
	NewStatus  string     `json:"new_status"`
	ModifiedBy *user.User `json:"modified_by"`
}

func NewMetalSheetStatusChange(data map[string]any) *MetalSheetStatusChange {
	return &MetalSheetStatusChange{
		ID:        int64(data["id"].(float64)),
		NewStatus: data["new_status"].(string),
		ModifiedBy: user.NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

// MetalSheetToolAssignment represents a metal sheet tool assignment event.
type MetalSheetToolAssignment struct {
	SheetID    int64      `json:"sheet_id"`
	ToolID     *int64     `json:"tool_id"`
	ModifiedBy *user.User `json:"modified_by"`
}

func NewMetalSheetToolAssignment(data map[string]any) *MetalSheetToolAssignment {
	assignment := &MetalSheetToolAssignment{
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

// PressCycleAdd represents a press cycle creation event.
type PressCycleAdd struct {
	SlotTop         int64      `json:"slot_top"`
	SlotTopCassette int64      `json:"slot_top_cassette"`
	SlotBottom      int64      `json:"slot_bottom"`
	TotalCycles     int64      `json:"total_cycles"`
	ModifiedBy      *user.User `json:"modified_by"`
}

func NewPressCycleAdd(data map[string]any) *PressCycleAdd {
	return &PressCycleAdd{
		SlotTop:         int64(data["slot_top"].(float64)),
		SlotTopCassette: int64(data["slot_top_cassette"].(float64)),
		SlotBottom:      int64(data["slot_bottom"].(float64)),
		TotalCycles:     int64(data["total_cycles"].(float64)),
		ModifiedBy: user.NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}

// PressCycleUpdate represents a press cycle update event.
type PressCycleUpdate struct {
	SlotTop         int64      `json:"slot_top"`
	SlotTopCassette int64      `json:"slot_top_cassette"`
	SlotBottom      int64      `json:"slot_bottom"`
	TotalCycles     int64      `json:"total_cycles"`
	ModifiedBy      *user.User `json:"modified_by"`
}

func NewPressCycleUpdate(data map[string]any) *PressCycleUpdate {
	return &PressCycleUpdate{
		SlotTop:         int64(data["slot_top"].(float64)),
		SlotTopCassette: int64(data["slot_top_cassette"].(float64)),
		SlotBottom:      int64(data["slot_bottom"].(float64)),
		TotalCycles:     int64(data["total_cycles"].(float64)),
		ModifiedBy: user.NewUserFromInterfaceMap(
			data["modified_by"].(map[string]any),
		),
	}
}
