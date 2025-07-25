package troublereports

import (
	"errors"
	"io/fs"
	"slices"
	"time"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/pgvis/logger"
	"github.com/knackwurstking/pg-vis/routes/constants"
	"github.com/knackwurstking/pg-vis/routes/internal/utils"
	"github.com/labstack/echo/v4"
)

type ModificationsTemplateData struct {
	User              *pgvis.User
	TroubleReport     *pgvis.TroubleReport
	LoadedAttachments []*pgvis.Attachment
	Mods              pgvis.Mods[pgvis.TroubleReportMod]
	ModAttachments    map[int64][]*pgvis.Attachment // map modification time to its attachments
}

func (mtd *ModificationsTemplateData) FirstModified() *pgvis.Modified[pgvis.TroubleReportMod] {
	if len(mtd.TroubleReport.Mods) == 0 {
		return nil
	}
	return mtd.TroubleReport.Mods[0]
}

type ModificationsHandler struct {
	db               *pgvis.DB
	serverPathPrefix string
	templates        fs.FS
}

func (h *ModificationsHandler) RegisterRoutes(e *echo.Echo) {
	modificationsPath := h.serverPathPrefix + "/trouble-reports/modifications/:id"

	e.GET(modificationsPath, func(c echo.Context) error {
		return h.handleGetModifications(c, nil)
	})

	e.POST(modificationsPath, h.handlePostModifications)
}

func (h *ModificationsHandler) handleGetModifications(c echo.Context, tr *pgvis.TroubleReport) *echo.HTTPError {
	id, herr := utils.ParseInt64Param(c, constants.QueryParamID)
	if herr != nil {
		return herr
	}

	if tr == nil {
		var err error
		tr, err = h.db.TroubleReports.Get(id)
		if err != nil {
			return utils.HandlePgvisError(c, err)
		}
	}

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	// Load attachments for the trouble report
	loadedAttachments, err := h.db.TroubleReportService.LoadAttachments(tr)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	mods := slices.Clone(tr.Mods)
	slices.Reverse(mods)

	// Load attachments for each modification
	modAttachments := make(map[int64][]*pgvis.Attachment)
	for _, mod := range tr.Mods {
		if len(mod.Data.LinkedAttachments) > 0 {
			attachments, err := h.db.Attachments.GetByIDs(mod.Data.LinkedAttachments)
			if err != nil {
				return utils.HandlePgvisError(c, err)
			}
			modAttachments[mod.Time] = attachments
		}
	}

	data := &ModificationsTemplateData{
		User:              user,
		TroubleReport:     tr,
		LoadedAttachments: loadedAttachments,
		Mods:              mods,
		ModAttachments:    modAttachments,
	}

	return utils.HandleTemplate(
		c,
		data,
		h.templates,
		[]string{
			constants.TroubleReportsModificationsComponentTemplatePath,
		},
	)
}

func (h *ModificationsHandler) handlePostModifications(c echo.Context) error {
	id, herr := utils.ParseInt64Param(c, constants.QueryParamID)
	if herr != nil {
		return herr
	}

	timeQuery, herr := utils.ParseInt64Query(c, constants.QueryParamTime)
	if herr != nil {
		return herr
	}

	tr, err := h.db.TroubleReports.Get(id)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	// Move modification to the top (last item in list)
	newMods := []*pgvis.Modified[pgvis.TroubleReportMod]{}
	var mod *pgvis.Modified[pgvis.TroubleReportMod]
	for _, m := range tr.Mods {
		if m.Time == timeQuery {
			if mod != nil {
				// NOTE: Should never happen, but it is possible theoretically
				logger.TroubleReport().Warn("Multiple modifications with the same time, mod: %+v, m: %+v", mod, m)
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

	// Update mods with a new order
	tr.Mods = append(newMods, mod)

	// Update trouble reports data
	tr.Title = mod.Data.Title
	tr.Content = mod.Data.Content
	tr.LinkedAttachments = mod.Data.LinkedAttachments

	// Update trouble reports database
	if err = h.db.TroubleReports.Update(id, tr); err != nil {
		return utils.HandlePgvisError(c, pgvis.WrapError(err, "failed to update trouble report"))
	}

	return h.handleGetModifications(c, tr)
}
