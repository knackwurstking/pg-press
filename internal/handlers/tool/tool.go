package tool

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/tool/templates"
	"github.com/knackwurstking/pg-press/internal/helper"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"

	ui "github.com/knackwurstking/ui/ui-templ"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	db *common.DB
}

func NewHandler(db *common.DB) *Handler {
	return &Handler{
		db: db,
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo, path string) {
	ui.RegisterEchoRoutes(e, env.ServerPathPrefix, []*ui.EchoRoute{
		// Main Page
		ui.NewEchoRoute(http.MethodGet, path+"/:id", h.GetToolPage), // "is_cassette" defines the tool type

		// Regenerations Table
		ui.NewEchoRoute(http.MethodDelete,
			path+"/:id/delete-regeneration", h.HTMXDeleteRegeneration), // "id" is regeneration ID

		// Tool status and regenerations management
		ui.NewEchoRoute(http.MethodGet,
			path+"/:id/status-edit", h.HTMXGetStatusEdit),
		ui.NewEchoRoute(http.MethodGet,
			path+"/:id/status-display", h.HTMXGetStatusDisplay),
		ui.NewEchoRoute(http.MethodPut,
			path+"/:id/status", h.HTMXUpdateToolStatus),

		// Section loading
		ui.NewEchoRoute(http.MethodGet,
			path+"/:id/notes", h.HTMXGetToolNotes),
		ui.NewEchoRoute(http.MethodGet,
			path+"/:id/metal-sheets", h.HTMXGetToolMetalSheets),

		// Cycles table rows
		ui.NewEchoRoute(http.MethodGet,
			path+"/:id/cycles", h.GetCyclesSectionContent),
		ui.NewEchoRoute(http.MethodGet,
			path+"/:id/total-cycles", h.HTMXGetToolTotalCycles),

		// Delete a cycle table entry
		ui.NewEchoRoute(http.MethodDelete,
			path+"/cycle/delete", h.HTMXDeleteToolCycle),

		// Update tools binding data
		ui.NewEchoRoute(http.MethodPatch,
			path+"/:id/bind", h.HTMXPatchToolBinding),
		ui.NewEchoRoute(http.MethodPatch,
			path+"/:id/unbind", h.HTMXPatchToolUnBinding),
	})
}

func (h *Handler) GetToolPage(c echo.Context) *echo.HTTPError {
	user, merr := shared.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	id, merr := shared.ParseParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}

	var tool shared.ModelTool
	tool, merr = h.db.Tool.Tool.GetByID(shared.EntityID(id))
	if merr != nil {
		if merr.Code == http.StatusNotFound {
			tool, merr = h.db.Tool.Cassette.GetByID(shared.EntityID(id))
			if merr != nil {
				return merr.Echo()
			}
		}
		return merr.Echo()
	}

	t := templates.Page(&templates.PageProps{
		Tool: tool,
		User: user,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Tool Page")
	}

	return nil
}

func (h *Handler) HTMXPatchToolBinding(c echo.Context) *echo.HTTPError {
	user, merr := shared.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	id, merr := shared.ParseParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	tool, merr := h.db.Tool.Tool.GetByID(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	// Get target_id from form value
	targetIDString := c.FormValue("target_id")
	if targetIDString == "" {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			fmt.Sprintf("failed to parse target_id: %#v", targetIDString),
		)
	}
	id, err := strconv.ParseInt(targetIDString, 10, 64)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"invalid target_id",
		)
	}
	cassette, merr := h.db.Tool.Cassette.GetByID(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	// Bind tool to target, this will get an error if target already has a binding
	merr = helper.BindCassetteToTool(h.db, tool.ID, cassette.ID)
	if merr != nil {
		return merr.Echo()
	}

	return h.renderBindingSection(c, tool, user)
}

func (h *Handler) HTMXPatchToolUnBinding(c echo.Context) *echo.HTTPError {
	slog.Info("Initiating tool unbinding operation")

	user, merr := shared.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	id, merr := shared.ParseParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	merr = helper.UnbindCassetteFromTool(h.db, shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	return h.renderBindingSection(c, nil, user)
}

func (h *Handler) GetCyclesSectionContent(c echo.Context) *echo.HTTPError {
	return h.renderCyclesSectionContent(c)
}

func (h *Handler) HTMXGetToolTotalCycles(c echo.Context) *echo.HTTPError {
	id, merr := shared.ParseParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}

	totalCycles, merr := helper.GetTotalCyclesForTool(h.db, shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	t := templates.TotalCycles(totalCycles, shared.ParseQueryBool(c, "input"))
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "TotalCycles")
	}
	return nil
}

func (h *Handler) HTMXDeleteToolCycle(c echo.Context) *echo.HTTPError {
	id, merr := shared.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	cycleID := shared.EntityID(id)

	merr = h.db.Press.Cycle.Delete(cycleID)
	if merr != nil {
		return merr.Echo()
	}

	urlb.SetHXTrigger(c, "reload-cycles-section")

	return nil
}

func (h *Handler) HTMXGetToolMetalSheets(c echo.Context) *echo.HTTPError {
	slog.Info("Retrieving metal sheet entries for tool")

	user, merr := shared.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	id, merr := shared.ParseParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	var tool shared.ModelTool
	tool, merr = h.db.Tool.Tool.GetByID(shared.EntityID(id))
	// If not found, try cassette
	if merr != nil {
		if merr.Code == http.StatusNotFound {
			tool, merr = h.db.Tool.Cassette.GetByID(shared.EntityID(id))
			if merr != nil {
				return merr.Echo()
			}
		} else {
			return merr.Echo()
		}
	}

	var t templ.Component

	// Fetch metal sheets for tool
	switch p := tool.GetBase().Position; p {
	case shared.SlotUpper:
		metalSheets, merr := helper.ListUpperMetalSheetsForTool(h.db, tool.GetID())
		if merr != nil {
			return merr.Echo()
		}
		t = templates.UppperMetalSheets(metalSheets, tool.(*shared.Tool), user)
	case shared.SlotLower:
		metalSheets, merr := helper.ListLowerMetalSheetsForTool(h.db, tool.GetID())
		if merr != nil {
			return merr.Echo()
		}
		t = templates.LowerMetalSheets(metalSheets, tool.(*shared.Tool), user)
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "Tool is not supported for metal sheets")
	}

	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "MetalSheets")
	}
	return nil
}

func (h *Handler) HTMXGetToolNotes(c echo.Context) *echo.HTTPError {
	id, merr := shared.ParseParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	toolID := shared.EntityID(id)
	notes, merr := helper.ListNotesForLinked(h.db, "tool", toolID)
	if merr != nil {
		return merr.Echo()
	}

	t := templates.Notes(toolID, notes)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Notes")
	}
	return nil
}

func (h *Handler) HTMXDeleteRegeneration(c echo.Context) *echo.HTTPError {
	id, merr := shared.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	regeneration, merr := h.db.Tool.Regeneration.GetByID(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	merr = h.db.Tool.Regeneration.Delete(regeneration.ID)
	if merr != nil {
		return merr.Echo()
	}

	urlb.SetHXTrigger(c, "reload-cycles-section")
	return nil
}

func (h *Handler) HTMXGetStatusEdit(c echo.Context) *echo.HTTPError {
	id, merr := shared.ParseParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	tool, merr := h.db.Tool.Tool.GetByID(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	eerr := h.renderRegenerationEdit(c, tool, true, nil)
	if eerr != nil {
		return eerr
	}

	return nil
}

func (h *Handler) HTMXGetStatusDisplay(c echo.Context) *echo.HTTPError {
	id, merr := shared.ParseParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	tool, merr := h.db.Tool.Tool.GetByID(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	eerr := h.renderRegenerationEdit(c, tool, false, nil)
	if eerr != nil {
		return eerr
	}

	return nil
}

func (h *Handler) HTMXUpdateToolStatus(c echo.Context) *echo.HTTPError {
	slog.Info("Change the tool status")

	user, merr := shared.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	id, merr := shared.ParseParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	tool, merr := h.db.Tool.Tool.GetByID(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	statusStr := c.FormValue("status")
	if statusStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "status is required")
	}

	// Handle regeneration start/stop/abort only
	switch statusStr {
	case "regenerating":
		merr = helper.StartToolRegeneration(h.db, tool.ID)
		if merr != nil {
			return merr.Echo()
		}

	case "active":
		merr = helper.StopToolRegeneration(h.db, tool.ID)
		if merr != nil {
			return merr.Echo()
		}

	case "abort":
		merr := helper.AbortToolRegeneration(h.db, tool.ID)
		if merr != nil {
			return merr.Echo()
		}

	default:
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"invalid status: must be 'regenerating', 'active', or 'abort'",
		)
	}

	// Get updated tool and render status display
	tool, merr = h.db.Tool.Tool.GetByID(tool.ID)
	if merr != nil {
		return merr.Echo()
	}

	// Render the updated status component
	eerr := h.renderRegenerationEdit(c, tool, false, user)
	if eerr != nil {
		return eerr
	}

	return h.renderCyclesSection(c, tool)
}

func (h *Handler) renderCyclesSection(c echo.Context, tool *shared.Tool) *echo.HTTPError {
	// Render out-of-band swap for cycles section to trigger reload
	t := templates.CyclesSection(true, tool.GetID(), !tool.IsCassette())
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "CyclesSection")
	}
	return nil
}

func (h *Handler) renderCyclesSectionContent(c echo.Context) *echo.HTTPError {
	// Get tool from URL param "id"
	id, merr := shared.ParseParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	toolID := shared.EntityID(id)

	tool, merr := h.db.Tool.Tool.GetByID(toolID)
	if merr != nil {
		return merr.Echo()
	}

	// Get cycles for this specific tool
	toolCycles, merr := helper.ListCyclesForTool(h.db, toolID)
	if merr != nil {
		return merr.Echo()
	}

	// Get active press number for this tool, -1 if none
	activePressNumber := helper.GetPressNumberForTool(h.db, toolID)

	// Get bindable cassettes for this tool, if it is a tool and not a cassette
	cassettesForBinding, merr := helper.ListAvailableCassettesForBinding(h.db, toolID)
	if merr != nil {
		return merr.Echo()
	}

	// Get regenerations for this tool
	regenerations, merr := helper.GetRegenerationsForTool(h.db, toolID)
	if merr != nil {
		return merr.Echo()
	}

	// Get user from context
	user, merr := shared.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	t := templates.CyclesSectionContent(templates.CyclesSectionContentProps{
		Tool:                tool,
		ToolCycles:          toolCycles,
		ActivePressNumber:   activePressNumber,
		CassettesForBinding: cassettesForBinding,
		Regenerations:       regenerations,
		User:                user,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Cycles")
	}

	return nil
}

func (h *Handler) renderRegenerationEdit(c echo.Context, tool shared.ModelTool, editable bool, user *shared.User) *echo.HTTPError {
	if user == nil {
		var merr *errors.MasterError
		user, merr = shared.GetUserFromContext(c)
		if merr != nil {
			return merr.Echo()
		}
	}

	regenerations, merr := helper.GetRegenerationsForTool(h.db, tool.GetID())
	if merr != nil && merr.Code != http.StatusNotFound {
		return merr.Echo()
	}
	isRegenerating := false
	for _, r := range regenerations {
		if r.Stop == 0 {
			isRegenerating = true
			break
		}
	}

	t := templates.RegenerationEdit(templates.RegenerationEditProps{
		Tool:              tool,
		ActivePressNumber: helper.GetPressNumberForTool(h.db, tool.GetID()),
		IsRegenerating:    isRegenerating,
		Editable:          editable,
		UserHasPermission: user.IsAdmin(),
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "ToolStatusEdit")
	}
	return nil
}

func (h *Handler) renderBindingSection(c echo.Context, tool *shared.Tool, user *shared.User) *echo.HTTPError {
	cassettesForBinding, merr := helper.ListAvailableCassettesForBinding(h.db, tool.ID)
	if merr != nil {
		return merr.Echo()
	}

	// Render the template
	t := templates.BindingSection(templates.BindingSectionProps{
		Tool:                tool,
		CassettesForBinding: cassettesForBinding,
		IsAdmin:             user.IsAdmin(),
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "BindingSection")
	}
	return nil
}
