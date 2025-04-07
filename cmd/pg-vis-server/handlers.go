package main

import (
	"github.com/knackwurstking/pg-vis/internal/constants"
	"github.com/knackwurstking/pg-vis/internal/httperrors"
	"github.com/labstack/echo/v4"
)

func setHandlers(e *echo.Echo) {
	// TODO: Set Api Routes here
	//     - "/registration" => Api key input and info how to get a new api key
	//     - "/" => Home page showing the current api key in use, the user
	//              name and the telegram id and api key permissions setup
	//     - "/metal-sheets" => List all existing sheets here, also allow to add a new one
	//     - "/metal-sheets/:format/:toolID" => Show entries, allow add, modify and deletion

	e.GET(constants.BaseServerPath+"/", func(c echo.Context) error {
		// TODO: Test this first
		// TODO: Deliver the home page here
		return httperrors.NewUnderConstruction()
	})
}
