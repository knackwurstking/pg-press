package htmxhandler

import (
	"errors"
	"net/http"
	"slices"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/internal/constants"
	"github.com/knackwurstking/pg-vis/internal/database"
	"github.com/knackwurstking/pg-vis/internal/logger"
	"github.com/knackwurstking/pg-vis/internal/utils"
)

func (h *TroubleReports) handleGetData(c echo.Context) error {
	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	trs, err := h.DB.TroubleReportService.ListWithAttachments()
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	return utils.HandleTemplate(
		c,
		troubleReportsDataTemplateData{
			TroubleReports: trs,
			User:           user,
		},
		h.Templates,
		[]string{
			constants.HTMXTroubleReportsDataTemplatePath,
		},
	)
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

	if err := h.DB.TroubleReportService.RemoveWithAttachments(id); err != nil {
		return utils.HandlePgvisError(c, err)
	}

	return h.handleGetData(c)
}

func (h *TroubleReports) handleGetAttachmentsPreview(c echo.Context) error {
	id, herr := utils.ParseInt64Query(c, "id")
	if herr != nil {
		return herr
	}

	tr, err := h.DB.TroubleReportService.GetWithAttachments(id)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	return utils.HandleTemplate(
		c,
		attachmentsPreviewTemplateData{
			TroubleReport: tr,
		},
		h.Templates,
		[]string{
			constants.HTMXTroubleReportsAttachmentsPreviewTemplatePath,
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
			return utils.HandlePgvisError(c, err)
		}
	}

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	loadedAttachments, err := h.DB.TroubleReportService.LoadAttachments(tr)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	mods := slices.Clone(tr.Mods)
	slices.Reverse(mods)

	data := &modificationsTemplateData{
		User:              user,
		TroubleReport:     tr,
		LoadedAttachments: loadedAttachments,
		Mods:              mods,
	}

	return utils.HandleTemplate(
		c,
		data,
		h.Templates,
		[]string{
			constants.HTMXTroubleReportsModificationsTemplatePath,
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
		return utils.HandlePgvisError(c, err)
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
		return utils.HandlePgvisError(c, errors.New("modification not found"))
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
		return utils.HandlePgvisError(c, database.WrapError(err, "failed to update trouble report"))
	}

	return h.handleGetModifications(c, tr)
}
