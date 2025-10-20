package helpers

import (
	"fmt"
	"strings"

	"github.com/knackwurstking/pgpress/internal/env"
)

// Category: Cycles

func URLGetCycleEdit(toolID int64, cycleID *int64, toolChangeMode bool) string {
	props := make([]string, 0)

	props = append(props, fmt.Sprintf("tool_id=%d", toolID))

	if cycleID != nil {
		props = append(props, fmt.Sprintf("id=%d", *cycleID))
	}

	props = append(props, fmt.Sprintf("tool_change_mode=%t", toolChangeMode))

	return fmt.Sprintf(
		"%s/htmx/tools/cycle/edit?%s",
		env.ServerPathPrefix, strings.Join(props, "&"),
	)
}

func URLDeleteCycle(cycleID, toolID int64) string {
	return fmt.Sprintf(
		"%s/htmx/tools/cycle/delete?id=%d&tool_id=%d",
		env.ServerPathPrefix, cycleID, toolID,
	)
}

// Category: Feeds

func URLGetFeedList() string {
	return fmt.Sprintf(
		"%s/htmx/feed/list",
		env.ServerPathPrefix,
	)
}

// Category: Metal Sheets

func URLGetMetalSheetEdit(metalSheetID *int64) string {
	if metalSheetID == nil {
		return fmt.Sprintf("%s/htmx/metal-sheets/edit", env.ServerPathPrefix)
	}

	return fmt.Sprintf(
		"%s/htmx/metal-sheets/edit?id=%d",
		env.ServerPathPrefix, metalSheetID,
	)
}

func URLPostMetalSheetEdit(toolID int64) string {
	return fmt.Sprintf(
		"%s/htmx/metal-sheets/edit?tool_id=%d",
		env.ServerPathPrefix, toolID,
	)
}

func URLPutMetalSheetEdit(metalSheetID int64) string {
	return fmt.Sprintf(
		"%s/htmx/metal-sheets/edit?id=%d",
		env.ServerPathPrefix, metalSheetID,
	)
}

func URLDeleteMetalSheet(metalSheetID int64) string {
	return fmt.Sprintf(
		"%s/htmx/metal-sheets/delete?id=%d",
		env.ServerPathPrefix, metalSheetID,
	)
}

// Category: Notes

func URLGetNotesEdit(noteID *int64, linkToTables ...string) string {
	var params []string

	if noteID != nil {
		params = append(params, fmt.Sprintf("id=%d", *noteID))
	}

	if len(linkToTables) > 0 {
		params = append(params, fmt.Sprintf(
			"link_to_tables=%s", strings.Join(linkToTables, ","),
		))
	}

	url := fmt.Sprintf("%s/htmx/notes/edit", env.ServerPathPrefix)
	if len(params) > 0 {
		url += "?" + strings.Join(params, "&")
	}

	return url
}

func URLPostNotesEdit(linkToTables ...string) string {
	if len(linkToTables) == 0 {
		return fmt.Sprintf("%s/htmx/notes/edit", env.ServerPathPrefix)
	}

	return fmt.Sprintf(
		"%s/htmx/notes/edit?link_to_tables=%s",
		env.ServerPathPrefix, strings.Join(linkToTables, ","),
	)
}

func URLPutNotesEdit(noteID int64, linkToTables ...string) string {
	if len(linkToTables) == 0 {
		return fmt.Sprintf("%s/htmx/notes/edit?id=%d", env.ServerPathPrefix, noteID)
	}

	return fmt.Sprintf(
		"%s/htmx/notes/edit?id=%d&link_to_tables=%s",
		env.ServerPathPrefix, noteID, strings.Join(linkToTables, ","),
	)
}

func URLDeleteNote(noteID int64) string {
	return fmt.Sprintf("%s/htmx/notes/delete?id=%d", env.ServerPathPrefix, noteID)
}

// TODO: Category: Press Page

// Get
// "%s/htmx/tools/press/%d/notes"

// Get
// "%s/htmx/tools/press/%d/active-tools"

// Get
// "%s/htmx/tools/press/%d/metal-sheets"

// Get
// "%s/htmx/tools/press/%d/cycle-summary-pdf"

//.Get
// "%s/htmx/tools/press/%d/cycles"
