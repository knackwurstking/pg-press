package htmx

import (
	"net/http"

	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/helpers"
	"github.com/knackwurstking/pgpress/internal/web/templates/components"
	"github.com/knackwurstking/pgpress/internal/web/templates/troublereportspage"
	"github.com/knackwurstking/pgpress/pkg/utils"

	"github.com/labstack/echo/v4"
)

func (h *TroubleReports) handleGetData(c echo.Context) error {
	user, err := helpers.GetUserFromContext(c)
	if err != nil {
		return err
	}

	logger.HTMXHandlerTroubleReports().Debug("User %s fetching trouble reports list", user.Name)

	trs, err := h.DB.TroubleReports.ListWithAttachments()
	if err != nil {
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to load trouble reports: "+err.Error())
	}

	logger.HTMXHandlerTroubleReports().Debug("Found %d trouble reports for user %s", len(trs), user.Name)

	troubleReportsList := troublereportspage.List(user, trs)
	if err := troubleReportsList.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to render trouble reports list component: "+err.Error())
	}
	return nil
}

func (h *TroubleReports) handleDeleteData(c echo.Context) error {
	id, err := helpers.ParseInt64Query(c, "id")
	if err != nil {
		return err
	}

	user, err := helpers.GetUserFromContext(c)
	if err != nil {
		return err
	}

	if !user.IsAdmin() {
		return echo.NewHTTPError(
			http.StatusForbidden,
			"administrator privileges required",
		)
	}

	logger.HTMXHandlerTroubleReports().Info("Administrator %s (Telegram ID: %d) is deleting trouble report %d",
		user.Name, user.TelegramID, id)

	if removedReport, err := h.DB.TroubleReports.RemoveWithAttachments(id, user); err != nil {
		logger.HTMXHandlerTroubleReports().Error("Failed to delete trouble report %d: %v", id, err)
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to delete trouble report: "+err.Error())
	} else {
		logger.HTMXHandlerTroubleReports().Info("Successfully deleted trouble report %d (%s)", removedReport.ID, removedReport.Title)
	}

	return h.handleGetData(c)
}

func (h *TroubleReports) handleGetAttachmentsPreview(c echo.Context) error {
	id, err := helpers.ParseInt64Query(c, "id")
	if err != nil {
		return err
	}

	logger.HTMXHandlerTroubleReports().Debug("Fetching attachments preview for trouble report %d", id)

	tr, err := h.DB.TroubleReports.GetWithAttachments(id)
	if err != nil {
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to load trouble report: "+err.Error())
	}

	logger.HTMXHandlerTroubleReports().Debug(
		"Rendering attachments preview with %d attachments", len(tr.LoadedAttachments),
	)

	attachmentsPreview := components.AttachmentsPreview(tr.LoadedAttachments)
	if err := attachmentsPreview.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render attachments preview component: "+err.Error())
	}
	return nil
}
