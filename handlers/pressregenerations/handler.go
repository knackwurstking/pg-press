package pressregenerations

import (
	"net/http"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/handlers/pressregenerations/components"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	registry *services.Registry
}

func NewHandler(r *services.Registry) *Handler {
	return &Handler{
		registry: r,
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	utils.RegisterEchoRoutes(e, []*utils.EchoRoute{
		// Press regeneration page
		utils.NewEchoRoute(
			http.MethodGet,
			"/tools/press/:press/regenerations",
			h.GetPressRegenerationsPage,
		),

		utils.NewEchoRoute(
			http.MethodPost,
			"/tools/press/:press/regenerations",
			h.PostPressRegenerationsPage,
		),

		// TODO: Remove
		// HTMX endpoints for press regeneration content
		//utils.NewEchoRoute(
		//	http.MethodGet,
		//	"/htmx/tools/press/:press/regenerations/history",
		//	h.HTMXGetPressRegenerationHistory,
		//),
		//utils.NewEchoRoute(
		//	http.MethodPost,
		//	"/htmx/tools/press/:press/regenerations/start",
		//	h.HTMXStartPressRegeneration,
		//),
		//utils.NewEchoRoute(
		//	http.MethodPost,
		//	"/htmx/tools/press/:press/regenerations/stop",
		//	h.HTMXStopPressRegeneration,
		//),
		//utils.NewEchoRoute(
		//	http.MethodPost,
		//	"/htmx/tools/press/:press/regenerations/add",
		//	h.HTMXAddPressRegeneration,
		//),
	})
}

func (h *Handler) GetPressRegenerationsPage(c echo.Context) error {
	press, eerr := h.parseParamPress(c)
	if eerr != nil {
		return eerr
	}

	if err := components.PagePressRegenerations(components.PagePressRegenerationsProps{
		Press: press,
	}).Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render press page template")
	}

	return nil
}

func (h *Handler) PostPressRegenerationsPage(c echo.Context) error {
	// TODO: Parse form...

	return errors.BadRequest(nil, "Under Construction")
}

func (h *Handler) parseParamPress(c echo.Context) (models.PressNumber, *echo.HTTPError) {
	pressNum, err := utils.ParseParamInt8(c, "press")
	if err != nil {
		return -1, errors.BadRequest(err, "invalid or missing press parameter")
	}

	press := models.PressNumber(pressNum)
	if !models.IsValidPressNumber(&press) {
		return -1, errors.BadRequest(err, "invalid press number")
	}

	return press, nil
}

//func (h *Handler) HTMXGetPressRegenerationHistory(c echo.Context) error {
//	press, err := h.getPressNumberFromParam(c)
//	if err != nil {
//		return err
//	}
//
//	// For now, just return a simple response to ensure the handler works
//	return c.String(
//		http.StatusOK,
//		"Regeneration History for Press "+
//			string(rune(press+'0')),
//	)
//}
//
//func (h *Handler) HTMXStartPressRegeneration(c echo.Context) error {
//	press, err := h.getPressNumberFromParam(c)
//	if err != nil {
//		return err
//	}
//
//	reason := c.FormValue("reason")
//	if reason == "" {
//		return errors.BadRequest(nil, "reason is required")
//	}
//
//	// For now, just return a simple response to ensure the handler works
//	return c.String(
//		http.StatusOK,
//		"Started regeneration for Press "+
//			string(rune(press+'0'))+
//			" with reason: "+reason,
//	)
//}
//
//func (h *Handler) HTMXStopPressRegeneration(c echo.Context) error {
//	press, err := h.getPressNumberFromParam(c)
//	if err != nil {
//		return err
//	}
//
//	// For now, just return a simple response to ensure the handler works
//	return c.String(
//		http.StatusOK,
//		"Stopped regeneration for Press "+
//			string(rune(press+'0')),
//	)
//}
//
//func (h *Handler) HTMXAddPressRegeneration(c echo.Context) error {
//	press, err := h.getPressNumberFromParam(c)
//	if err != nil {
//		return err
//	}
//
//	startedAtStr := c.FormValue("started_at")
//	if startedAtStr == "" {
//		return errors.BadRequest(nil, "started_at is required")
//	}
//
//	reason := c.FormValue("reason")
//	if reason == "" {
//		return errors.BadRequest(nil, "reason is required")
//	}
//
//	// For now, just return a simple response to ensure the handler works
//	return c.String(
//		http.StatusOK,
//		"Added regeneration for Press "+
//			string(rune(press+'0'))+
//			" with reason: "+reason,
//	)
//}
//
