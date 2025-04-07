package main

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/internal/constants"
	"github.com/knackwurstking/pg-vis/internal/httperrors"
)

func setHandlers(e *echo.Echo) {
	// TODO: Set Api Routes here
	//     - "/registration" => Api key input and info how to get a new api key
	//     - "/" => Home page showing the current api key in use, the user
	//              name and the telegram id and api key permissions setup
	//     - "/metal-sheets" => List all existing sheets here, also allow to add a new one
	//     - "/metal-sheets/:format/:toolID" => Show entries, allow add, modify and deletion

	e.GET(constants.BaseServerPath+"", func(c echo.Context) error {
		return c.Redirect(http.StatusSeeOther, constants.BaseServerPath+"/")
	})

	e.GET(constants.BaseServerPath+"/", func(c echo.Context) error {
		// TODO: Serve the app here, but only with a valid api key, so i need to add
		//       a auth middleware here

		return httperrors.NewUnderConstruction()
	})
}
