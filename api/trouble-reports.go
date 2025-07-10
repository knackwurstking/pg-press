package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func serverTroubleReports(e *echo.Echo, options Options) {
	e.POST(options.ServerPathPrefix+"/trouble-reports", func(c echo.Context) error {
		return postTroubleReports()
	})

	// TODO: DELETE /api/trouble-reports/:id
}

func postTroubleReports() *echo.HTTPError {
	// TODO: POST /api/trouble-reports

	return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("under construction"))
}
