package htmxhandler

import (
	"net/http"
	"time"

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

	trs, err := h.DB.TroubleReportsHelper.ListWithAttachments()
	if err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to load trouble reports: "+err.Error())
	}

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

	logger.TroubleReport().Info("Administrator %s (Telegram ID: %d) is deleting trouble report %d",
		user.UserName, user.TelegramID, id)

	if removedReport, err := h.DB.TroubleReportsHelper.RemoveWithAttachments(id); err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to delete trouble report: "+err.Error())
	} else {
		feed := database.NewFeed(
			database.FeedTypeTroubleReportRemove,
			&database.FeedTroubleReportRemove{
				ID:        removedReport.ID,
				Title:     removedReport.Title,
				RemovedBy: user,
			},
		)
		if err := h.DB.Feeds.Add(feed); err != nil {
			return database.WrapError(err, "failed to add feed entry")
		}
	}

	return h.handleGetData(c)
}

func (h *TroubleReports) handleGetAttachmentsPreview(c echo.Context) error {
	id, err := utils.ParseInt64Query(c, constants.QueryParamID)
	if err != nil {
		return err
	}

	tr, err := h.DB.TroubleReportsHelper.GetWithAttachments(id)
	if err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to load trouble report: "+err.Error())
	}

	// If "time" query is provided, return modified data attachments for "id"
	timeQuery, err := utils.ParseInt64Query(c, constants.QueryParamTime)
	if err == nil {
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
					return echo.NewHTTPError(database.GetHTTPStatusCode(err),
						"failed to load attachments for modified trouble report: "+err.Error())
				}

				// Use the modified trouble report with loaded attachments
				tr.TroubleReport = modifiedTr
				tr.LoadedAttachments = loadedAttachments
				break
			}
		}
	}

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

	var firstMod *database.Modified[database.TroubleReportMod]
	if len(tr.Mods) > 0 {
		firstMod = tr.Mods[len(tr.Mods)-1]
	}

	// FIXME: Mods needs to be fixed...
	logger.TroubleReport().Debug("Trouble report %d has %d modifications", id, len(tr.Mods))
	for i, mod := range tr.Mods {
		logger.TroubleReport().Debug("Mod %d: Time=%d, Title=%s", i, mod.Time, mod.Data.Title)
	}

	trModifications := components.TroubleReportModifications(
		user,
		tr,
		firstMod,
		tr.Mods,
	)
	if err := trModifications.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render trouble report modifications component: "+err.Error())
	}
	return nil
}

func (h *TroubleReports) handlePostModifications(c echo.Context) error {
	id, err := utils.ParseInt64Param(c, constants.QueryParamID)
	if err != nil {
		return err
	}

	timeQuery, err := utils.ParseInt64Query(c, constants.QueryParamTime)
	if err != nil {
		return err
	}

	tr, err := h.DB.TroubleReports.Get(id)
	if err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to load trouble report: "+err.Error())
	}

	// Move modification to the top
	newMods := []*database.Modified[database.TroubleReportMod]{}
	var mod *database.Modified[database.TroubleReportMod]
	for _, m := range tr.Mods {
		if m.Time == timeQuery {
			if mod != nil {
				logger.TroubleReport().Warn(
					"Multiple modifications with the same time, mod: %+v, m: %+v", mod, m)
				newMods = append(newMods, m)
			} else {
				mod = m
			}
		} else {
			newMods = append(newMods, m)
		}
	}

	if mod == nil {
		return echo.NewHTTPError(http.StatusNotFound, "modification not found")
	}

	mod.Time = time.Now().UnixMilli()

	// Update mods with new order
	tr.Mods = append(newMods, mod)

	// Update trouble reports data
	tr.Title = mod.Data.Title
	tr.Content = mod.Data.Content
	tr.LinkedAttachments = mod.Data.LinkedAttachments

	// Update database
	if err = h.DB.TroubleReports.Update(id, tr); err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to update trouble report: "+err.Error())
	}

	return h.handleGetModifications(c, tr)
}
