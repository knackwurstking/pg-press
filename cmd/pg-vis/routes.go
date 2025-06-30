package main

import (
	"errors"

	"github.com/labstack/echo/v4"
)

func initRouter(e *echo.Echo) {
	e.GET(serverPathPrefix+"/", func(c echo.Context) error {
		// TODO: ...

		return errors.New("under construction")
	})

	e.GET(serverPathPrefix+"/feed", func(c echo.Context) error {
		// TODO: ...

		return errors.New("under construction")
	})
}
