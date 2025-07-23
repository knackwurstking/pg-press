package troublereports

import (
	"net/http"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/pgvis/logger"
	"github.com/knackwurstking/pg-vis/routes/constants"
	"github.com/knackwurstking/pg-vis/routes/internal/utils"
	"github.com/labstack/echo/v4"
)

type TroubleReportsTemplateData struct {
	TroubleReports []*pgvis.TroubleReport `json:"trouble_reports"`
	User           *pgvis.User            `json:"user"`
}

func (h *Handler) handleGetData(c echo.Context) error {
	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	trs, err := h.db.TroubleReports.List()
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	return utils.HandleTemplate(
		c,
		TroubleReportsTemplateData{
			TroubleReports: trs,
			User:           user,
		},
		h.templates,
		[]string{
			constants.TroubleReportsDataComponentTemplatePath,
		},
	)
}

func (h *Handler) handleDeleteData(c echo.Context) error {
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

	if err := h.db.TroubleReports.Remove(id); err != nil {
		return utils.HandlePgvisError(c, err)
	}

	return h.handleGetData(c)
}
