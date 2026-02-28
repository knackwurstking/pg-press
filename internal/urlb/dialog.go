package urlb

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/internal/shared"
)

// DialogEditToolRegenerationGet constructs edit tool regeneration dialog GET URL
func DialogEditToolRegenerationGet(toolID shared.EntityID) templ.SafeURL {
	return BuildURLWithParams(
		"/dialog/edit-tool-regeneration",
		map[string]string{
			"tool_id": toolID.String(),
		},
	)
}

// DialogEditToolRegenerationPost constructs edit tool regeneration dialog POST URL
func DialogEditToolRegenerationPost(toolID shared.EntityID) templ.SafeURL {
	return DialogEditToolRegenerationGet(toolID)
}

// DialogEditToolRegenerationPut constructs edit tool regeneration dialog PUT URL
func DialogEditToolRegenerationPut(toolRegenerationID shared.EntityID) templ.SafeURL {
	return BuildURLWithParams(
		"/dialog/edit-tool-regeneration",
		map[string]string{
			"id": toolRegenerationID.String(),
		},
	)
}

// DialogEditCycleGet constructs edit cycle dialog GET URL
func DialogEditCycleGet(cycleID shared.EntityID, toolID shared.EntityID, toolChangeMode bool) templ.SafeURL {
	params := map[string]string{}
	if cycleID != 0 {
		params["id"] = fmt.Sprintf("%d", cycleID)
	}
	if toolID != 0 {
		params["tool_id"] = fmt.Sprintf("%d", toolID)
	}
	if toolChangeMode {
		params["tool_change_mode"] = "true"
	}
	return BuildURLWithParams("/dialog/edit-cycle", params)
}

// DialogEditCyclePost constructs edit cycle dialog POST URL
func DialogEditCyclePost() templ.SafeURL {
	return BuildURL("/dialog/edit-cycle")
}

// DialogEditCyclePut constructs edit cycle dialog PUT URL
func DialogEditCyclePut() templ.SafeURL {
	return BuildURL("/dialog/edit-cycle")
}

// DialogEditPressGet constructs edit press dialog GET URL
func DialogEditPress(pressID shared.EntityID) templ.SafeURL {
	params := map[string]string{}
	if pressID != 0 {
		params["id"] = fmt.Sprintf("%d", pressID)
	}

	return BuildURLWithParams("/dialog/edit-press", params)
}

func DialogEditTool(toolID shared.EntityID) templ.SafeURL {
	params := map[string]string{}
	if toolID != 0 {
		params["id"] = fmt.Sprintf("%d", toolID)
	}

	return BuildURLWithParams("/dialog/edit-tool", params)
}

func DialogEditCassette(cassetteID shared.EntityID) templ.SafeURL {
	params := map[string]string{}
	if cassetteID != 0 {
		params["id"] = fmt.Sprintf("%d", cassetteID)
	}

	return BuildURLWithParams("/dialog/edit-cassette", params)
}

func DialogEditMetalSheet(metalSheetID shared.EntityID, toolID shared.EntityID, position shared.Slot) templ.SafeURL {
	params := map[string]string{}
	if metalSheetID != 0 {
		params["id"] = fmt.Sprintf("%d", metalSheetID)
		params["position"] = fmt.Sprintf("%d", position)
	}
	if toolID != 0 {
		params["tool_id"] = fmt.Sprintf("%d", toolID)
	}

	return BuildURLWithParams("/dialog/edit-metal-sheet", params)
}

func DialogEditNote(noteID shared.EntityID, linked string) templ.SafeURL {
	params := map[string]string{
		"linked": linked,
	}
	if noteID != 0 {
		params["id"] = fmt.Sprintf("%d", noteID)
	}

	return BuildURLWithParams("/dialog/edit-note", params)
}
