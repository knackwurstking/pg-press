// Package troublereports provides HTTP route handlers for trouble report management.
package troublereports

import (
	"io/fs"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/constants"
	"github.com/knackwurstking/pg-vis/routes/internal/utils"
)

const (
	InvalidContentFormFieldMessage = "invalid content form value"
	InvalidTitleFormFieldMessage   = "invalid title form value"
)

// PUTDialogEdit handles PUT requests to update existing trouble reports.
func PUTDialogEdit(templates fs.FS, c echo.Context, db *pgvis.DB) *echo.HTTPError {
	id, herr := utils.ParseRequiredIDQuery(c, constants.IDQueryParam)
	if herr != nil {
		return herr
	}

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	title, content, herr := extractAndValidateFormData(c)
	if herr != nil {
		return herr
	}

	dialogEditData := &EditDialogPageData{
		Submitted:      true,
		ID:             int(id),
		Title:          title,
		Content:        content,
		InvalidTitle:   title == "",
		InvalidContent: content == "",
	}

	if !dialogEditData.InvalidTitle && !dialogEditData.InvalidContent {
		trOld, err := db.TroubleReports.Get(id)
		if err != nil {
			return utils.HandlePgvisError(c, err)
		}

		modified := pgvis.NewModified(user, trOld)
		trNew := pgvis.NewTroubleReport(modified, title, content)

		if err := db.TroubleReports.Update(id, trNew); err != nil {
			return utils.HandlePgvisError(c, err)
		}
	} else {
		dialogEditData.Submitted = false
	}

	return GETDialogEdit(templates, c, db, dialogEditData)
}

// extractAndValidateFormData extracts and validates form data.
func extractAndValidateFormData(ctx echo.Context) (title, content string, httpErr *echo.HTTPError) {
	var err error

	title, err = url.QueryUnescape(ctx.FormValue(constants.TitleFormField))
	if err != nil {
		return "", "", echo.NewHTTPError(http.StatusBadRequest,
			pgvis.WrapError(err, InvalidTitleFormFieldMessage))
	}
	title = utils.SanitizeInput(title)

	content, err = url.QueryUnescape(ctx.FormValue(constants.ContentFormField))
	if err != nil {
		return "", "", echo.NewHTTPError(http.StatusBadRequest,
			pgvis.WrapError(err, InvalidContentFormFieldMessage))
	}
	content = utils.SanitizeInput(content)

	if httpErr := utils.ValidateStringLength(title, constants.TitleFormField, constants.TitleMinLength, constants.TitleMaxLength); httpErr != nil {
		return title, content, httpErr
	}

	if httpErr := utils.ValidateStringLength(content, constants.ContentFormField, constants.ContentMinLength, constants.ContentMaxLength); httpErr != nil {
		return title, content, httpErr
	}

	return title, content, nil
}
