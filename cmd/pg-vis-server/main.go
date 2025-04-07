package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"

	"github.com/knackwurstking/pg-vis/internal/constants"
)

func main() {
	e := echo.New()

	e.Logger.SetLevel(log.DEBUG)

	setHandlers(e)

	if err := e.Start(constants.ServerAddr); err != nil {
		e.Logger.Fatal(err)
	}
}
