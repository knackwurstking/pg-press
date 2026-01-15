package editor

import (
	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/editor/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"
	"github.com/labstack/echo/v4"
)

func Page(c echo.Context) *echo.HTTPError {
	editorType, eerr := getQueryEditorType(c)
	if eerr != nil {
		return eerr
	}

	id, eerr := getQueryID(c)
	if eerr != nil {
		return eerr
	}

	returnURL, eerr := getQueryReturnURL(c)
	if eerr != nil {
		return eerr
	}

	attachments := make([]string, 0) // TODO: Get existing attachments if editing already exists

	t := templates.Page(&templates.PageProps{
		Type:        editorType,
		ID:          id,
		ReturnURL:   returnURL,
		Title:       "", // TODO: Get title if editing already exists
		Content:     "", // TODO: Get content if editing already exists
		Attachments: attachments,
		UseMarkdown: false, // TODO: Get use_markdown from existing data
	})
	if err := t.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.NewRenderError(err, "editor page")
	}
	return nil
}

func getQueryEditorType(c echo.Context) (shared.EditorType, *echo.HTTPError) {
	t, merr := utils.GetQueryString(c, "type")
	if merr != nil {
		return "", merr.WrapEcho("editor type")
	}
	return shared.EditorType(t), nil
}

func getQueryID(c echo.Context) (shared.EntityID, *echo.HTTPError) {
	id, merr := utils.GetQueryInt64(c, "id")
	if merr != nil && !merr.IsNotFoundError() {
		return 0, merr.WrapEcho("editor id")
	}
	return shared.EntityID(id), nil
}

func getQueryReturnURL(c echo.Context) (templ.SafeURL, *echo.HTTPError) {
	u, merr := utils.GetQueryString(c, "return_url")
	if merr != nil {
		return "", merr.WrapEcho("editor return_url")
	}
	return templ.SafeURL(u), nil
}
