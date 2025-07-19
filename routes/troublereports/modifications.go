package troublereports

import (
	"io/fs"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/shared"
	"github.com/knackwurstking/pg-vis/routes/utils"
	"github.com/labstack/echo/v4"
)

type ModificationsPageData struct {
	User *pgvis.User
}

func GETModifications(templates fs.FS, c echo.Context, db *pgvis.DB) *echo.HTTPError {
	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	return utils.HandleTemplate(
		c,
		ModificationsPageData{
			User: user,
		},
		templates,
		[]string{
			shared.TroubleReportsModificationsTemplatePath,
		},
	)
}
