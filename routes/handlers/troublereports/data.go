package troublereports

import (
	"io/fs"
	"net/http"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/pgvis/logger"
	"github.com/knackwurstking/pg-vis/routes/constants"
	"github.com/knackwurstking/pg-vis/routes/internal/utils"
	"github.com/labstack/echo/v4"
)

type TemplateData struct {
	TroubleReports []*pgvis.TroubleReportWithAttachments `json:"trouble_reports"`
	User           *pgvis.User                           `json:"user"`
}

type DataHandler struct {
	db               *pgvis.DB
	serverPathPrefix string
	templates        fs.FS
}

func (h *DataHandler) RegisterRoutes(e *echo.Echo) {
	dataPath := h.serverPathPrefix + "/trouble-reports/data"
	attachmentsPreviewPath := h.serverPathPrefix + "/trouble-reports/attachments-preview"

	e.GET(dataPath, h.handleGetData)
	e.DELETE(dataPath, h.handleDeleteData)
	e.GET(attachmentsPreviewPath, h.handleGetAttachmentsPreview)
}

func (h *DataHandler) handleGetData(c echo.Context) error {
	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	trs, err := h.db.TroubleReportService.ListWithAttachments()
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	return utils.HandleTemplate(
		c,
		TemplateData{
			TroubleReports: trs,
			User:           user,
		},
		h.templates,
		[]string{
			constants.TroubleReportsDataComponentTemplatePath,
		},
	)
}

func (h *DataHandler) handleDeleteData(c echo.Context) error {
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

	if err := h.db.TroubleReportService.RemoveWithAttachments(id); err != nil {
		return utils.HandlePgvisError(c, err)
	}

	return h.handleGetData(c)
}

type AttachmentsPreviewTemplateData struct {
	TroubleReport *pgvis.TroubleReportWithAttachments `json:"trouble_report"`
}

func (h *DataHandler) handleGetAttachmentsPreview(c echo.Context) error {
	id, herr := utils.ParseInt64Query(c, "id")
	if herr != nil {
		return herr
	}

	tr, err := h.db.TroubleReportService.GetWithAttachments(id)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	return utils.HandleTemplate(
		c,
		AttachmentsPreviewTemplateData{
			TroubleReport: tr,
		},
		h.templates,
		[]string{
			constants.TroubleReportsAttachmentsPreviewComponentTemplatePath,
		},
	)
}
