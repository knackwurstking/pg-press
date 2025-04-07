package main

import (
	"github.com/knackwurstking/pg-vis/internal/constants"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func main() {
	e := echo.New()

	e.Logger.SetLevel(log.DEBUG)

	setHandlers(e)

	if err := e.Start(constants.ServerAddr); err != nil {
		e.Logger.Fatal(err)
	}
}
