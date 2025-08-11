package htmxhandler

import (
	"errors"
	"net/http"
	"slices"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/templates/components"
	"github.com/knackwurstking/pgpress/internal/utils"
)

func (h *TroubleReports) handleGetData(c echo.Context) error {
	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	trs, err := h.DB.TroubleReportsHelper.ListWithAttachments()
	if err != nil {
		return utils.HandlepgpressError(c, err)
	}

	troubleReportsList := components.TroubleReportsList(user, trs)
	if err := troubleReportsList.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to render trouble reports list component: "+err.Error())
	}
	return nil
}

func (h *TroubleReports) handleDeleteData(c echo.Context) error {
	id, herr := utils.ParseInt64Query(c, "id")
	if herr != nil {
		return herr
	}

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	if !user.IsAdmin() {
		return echo.NewHTTPError(
			http.StatusForbidden,
			adminPrivilegesRequiredMessage,
		)
	}

	logger.TroubleReport().Info("Administrator %s (Telegram ID: %d) is deleting trouble report %d",
		user.UserName, user.TelegramID, id)

	if removedReport, err := h.DB.TroubleReportsHelper.RemoveWithAttachments(id); err != nil {
		return utils.HandlepgpressError(c, err)
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
	id, herr := utils.ParseInt64Query(c, "id")
	if herr != nil {
		return herr
	}

	tr, err := h.DB.TroubleReportsHelper.GetWithAttachments(id)
	if err != nil {
		return utils.HandlepgpressError(c, err)
	}

	// If "time" query is provided, return modified data attachments for "id"
	timeQuery, herr := utils.ParseInt64Query(c, constants.QueryParamTime)
	if herr == nil {
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
					return utils.HandlepgpressError(c, err)
				}

				// Use the modified trouble report with loaded attachments
				tr.TroubleReport = modifiedTr
				tr.LoadedAttachments = loadedAttachments
				break
			}
		}
	}

	// TODO: Migrate to templ
	return utils.HandleTemplate(
		c,
		attachmentsPreviewTemplateData{
			TroubleReport: tr,
		},
		h.Templates,
		[]string{
			//constants.HTMXTroubleReportsAttachmentsPreviewTemplatePath,
		},
	)
}

func (h *TroubleReports) handleGetModifications(c echo.Context, tr *database.TroubleReport) *echo.HTTPError {
	id, herr := utils.ParseInt64Param(c, constants.QueryParamID)
	if herr != nil {
		return herr
	}

	if tr == nil {
		var err error
		tr, err = h.DB.TroubleReports.Get(id)
		if err != nil {
			return utils.HandlepgpressError(c, err)
		}
	}

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	mods := slices.Clone(tr.Mods)
	slices.Reverse(mods)

	data := &modificationsTemplateData{
		User:          user,
		TroubleReport: tr,
		Mods:          mods,
	}

	// TODO: Migrate to templ
	return utils.HandleTemplate(
		c,
		data,
		h.Templates,
		[]string{
			//constants.HTMXTroubleReportsModificationsTemplatePath,
		},
	)
}

func (h *TroubleReports) handlePostModifications(c echo.Context) error {
	id, herr := utils.ParseInt64Param(c, constants.QueryParamID)
	if herr != nil {
		return herr
	}

	timeQuery, herr := utils.ParseInt64Query(c, constants.QueryParamTime)
	if herr != nil {
		return herr
	}

	tr, err := h.DB.TroubleReports.Get(id)
	if err != nil {
		return utils.HandlepgpressError(c, err)
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
		return utils.HandlepgpressError(c, errors.New("modification not found"))
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
		return utils.HandlepgpressError(c, database.WrapError(err, "failed to update trouble report"))
	}

	return h.handleGetModifications(c, tr)
}
