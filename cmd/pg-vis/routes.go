package main

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

func initRouter(e *echo.Echo) {
	e.GET(serverPathPrefix+"/", func(c echo.Context) error {
		// TODO: ...

		return echo.NewHTTPError(
			http.StatusInternalServerError,
			errors.New("under construction"),
		)
	})

	e.GET(serverPathPrefix+"/signup", func(c echo.Context) error {
		// TODO: ...

		return echo.NewHTTPError(
			http.StatusInternalServerError,
			errors.New("under construction"),
		)
	})

	e.GET(serverPathPrefix+"/feed", func(c echo.Context) error {
		// TODO: ...

		return echo.NewHTTPError(
			http.StatusInternalServerError,
			errors.New("under construction"),
		)
	})
}
