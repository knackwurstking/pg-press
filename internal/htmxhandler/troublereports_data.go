package htmxhandler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/templates/components"
	"github.com/knackwurstking/pgpress/internal/utils"
)

func (h *TroubleReports) handleGetData(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return err
	}

	logger.HTMXHandlerTroubleReports().Debug("User %s fetching trouble reports list", user.UserName)

	trs, err := h.DB.TroubleReportsHelper.ListWithAttachments()
	if err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to load trouble reports: "+err.Error())
	}

	logger.HTMXHandlerTroubleReports().Debug("Found %d trouble reports for user %s", len(trs), user.UserName)

	troubleReportsList := components.TroubleReportsList(user, trs)
	if err := troubleReportsList.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to render trouble reports list component: "+err.Error())
	}
	return nil
}

func (h *TroubleReports) handleDeleteData(c echo.Context) error {
	id, err := utils.ParseInt64Query(c, constants.QueryParamID)
	if err != nil {
		return err
	}

	user, err := utils.GetUserFromContext(c)
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
		user.UserName, user.TelegramID, id)

	if removedReport, err := h.DB.TroubleReportsHelper.RemoveWithAttachments(id); err != nil {
		logger.HTMXHandlerTroubleReports().Error("Failed to delete trouble report %d: %v", id, err)
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to delete trouble report: "+err.Error())
	} else {
		logger.HTMXHandlerTroubleReports().Info("Successfully deleted trouble report %d (%s)", removedReport.ID, removedReport.Title)
		feed := database.NewFeed(
			database.FeedTypeTroubleReportRemove,
			&database.FeedTroubleReportRemove{
				ID:        removedReport.ID,
				Title:     removedReport.Title,
				RemovedBy: user,
			},
		)
		if err := h.DB.Feeds.Add(feed); err != nil {
			logger.HTMXHandlerTroubleReports().Error("Failed to add feed entry for deleted trouble report %d: %v", id, err)
			return database.WrapError(err, "failed to add feed entry")
		}
		logger.HTMXHandlerTroubleReports().Debug("Feed entry created for deleted trouble report %d", id)
	}

	return h.handleGetData(c)
}

func (h *TroubleReports) handleGetAttachmentsPreview(c echo.Context) error {
	id, err := utils.ParseInt64Query(c, constants.QueryParamID)
	if err != nil {
		return err
	}

	logger.HTMXHandlerTroubleReports().Debug("Fetching attachments preview for trouble report %d", id)

	tr, err := h.DB.TroubleReportsHelper.GetWithAttachments(id)
	if err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to load trouble report: "+err.Error())
	}

	// If "time" query is provided, return modified data attachments for "id"
	timeQuery, err := utils.ParseInt64Query(c, constants.QueryParamTime)
	if err == nil {
		logger.HTMXHandlerTroubleReports().Debug("Loading modified attachments for trouble report %d at time %d", id, timeQuery)
		for _, mod := range tr.Mods {
			if mod.Time == timeQuery {
				// Create a temporary trouble report with the modified data
				modifiedTr := &database.TroubleReport{
					ID:                tr.ID,
					Title:             mod.Data.Title,
					Content:           mod.Data.Content,
					LinkedAttachments: mod.Data.LinkedAttachments,
				}

				// Load attachments for the modified data
				loadedAttachments, err := h.DB.TroubleReportsHelper.LoadAttachments(modifiedTr)
				if err != nil {
					logger.HTMXHandlerTroubleReports().Error("Failed to load attachments for modified trouble report %d: %v", id, err)
					return echo.NewHTTPError(database.GetHTTPStatusCode(err),
						"failed to load attachments for modified trouble report: "+err.Error())
				}

				// Use the modified trouble report with loaded attachments
				tr.TroubleReport = modifiedTr
				tr.LoadedAttachments = loadedAttachments
				logger.HTMXHandlerTroubleReports().Debug("Loaded %d attachments for modified trouble report %d", len(loadedAttachments), id)
				break
			}
		}
	}

	logger.HTMXHandlerTroubleReports().Debug("Rendering attachments preview with %d attachments", len(tr.LoadedAttachments))
	attachmentsPreview := components.AttachmentsPreview(tr.LoadedAttachments)
	if err := attachmentsPreview.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render attachments preview component: "+err.Error())
	}
	return nil
}

func (h *TroubleReports) handleGetModifications(c echo.Context, tr *database.TroubleReport) error {
	id, err := utils.ParseInt64Param(c, constants.QueryParamID)
	if err != nil {
		return err
	}

	logger.HTMXHandlerTroubleReports().Debug("Loading modifications for trouble report %d", id)

	if tr == nil {
		var err error
		tr, err = h.DB.TroubleReports.Get(id)
		if err != nil {
			return echo.NewHTTPError(database.GetHTTPStatusCode(err),
				"failed to load trouble report: "+err.Error())
		}
	}

	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return err
	}

	// Create a reversed copy of tr.Mods
	trModifications := components.TroubleReportModifications(
		user,
		tr.ID,
		tr.Mods.First(),
		tr.Mods.Current(),
		tr.Mods,
	)
	if err := trModifications.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render trouble report modifications component: "+err.Error())
	}
	return nil
}

func (h *TroubleReports) handlePostModifications(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return err
	}

	id, err := utils.ParseInt64Param(c, constants.QueryParamID)
	if err != nil {
		return err
	}

	logger.HTMXHandlerTroubleReports().Info("Processing modification restore for trouble report %d", id)

	timeQuery, err := utils.ParseInt64Query(c, constants.QueryParamTime)
	if err != nil {
		return err
	}

	logger.HTMXHandlerTroubleReports().Debug("Restoring modification at time %d for trouble report %d", timeQuery, id)

	tr, err := h.DB.TroubleReports.Get(id)
	if err != nil {
		logger.HTMXHandlerTroubleReports().Error("Failed to load trouble report %d: %v", id, err)
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to load trouble report: "+err.Error())
	}

	// Move modification to the top
	mod, err := tr.Mods.Get(timeQuery)
	if err != nil {
		logger.HTMXHandlerTroubleReports().Error(
			"Failed to get modification at time %d for trouble report %d: %v",
			timeQuery, id, err,
		)

		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to get modification: "+err.Error())
	}

	tr.Title = mod.Data.Title
	tr.Content = mod.Data.Content
	tr.LinkedAttachments = mod.Data.LinkedAttachments

	// Update database
	logger.HTMXHandlerTroubleReports().Debug(
		"Updating trouble report %d with restored modification",
		id,
	)

	if err = h.DB.TroubleReports.Update(id, tr, user); err != nil {
		logger.HTMXHandlerTroubleReports().Error(
			"Failed to update trouble report %d with restored modification: %v",
			id, err,
		)

		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to update trouble report: "+err.Error())
	}

	logger.HTMXHandlerTroubleReports().Info(
		"Successfully restored modification for trouble report %d",
		id,
	)

	return h.handleGetModifications(c, tr)
}
