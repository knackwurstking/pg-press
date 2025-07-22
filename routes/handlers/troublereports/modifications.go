package troublereports

import (
	"errors"
	"slices"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/constants"
	"github.com/knackwurstking/pg-vis/routes/internal/utils"
	"github.com/labstack/echo/v4"
)

type ModificationsTemplateData struct {
	User          *pgvis.User
	TroubleReport *pgvis.TroubleReport
	Mods          pgvis.Mods[pgvis.TroubleReportMod]
}

func (mtd *ModificationsTemplateData) FirstModified() *pgvis.Modified[pgvis.TroubleReportMod] {
	if len(mtd.TroubleReport.Mods) == 0 {
		return nil
	}
	return mtd.TroubleReport.Mods[0]
}

func (h *Handler) handleGetModifications(c echo.Context) error {
	id, herr := utils.ParseInt64Param(c, constants.IDQueryParam)
	if herr != nil {
		return herr
	}

	tr, err := h.db.TroubleReports.Get(id)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	mods := slices.Clone(tr.Mods)
	slices.Reverse(mods)

	data := &ModificationsTemplateData{
		User:          user,
		TroubleReport: tr,
		Mods:          mods,
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

func (h *Handler) handlePostModifications(c echo.Context) error {
	// TODO: ...

	return errors.New("under construction")
}
