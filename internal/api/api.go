// Routes:
//   - /api/trouble-reports     - GET/POST
//   - /api/trouble-reports/:id - GET/PUT/DELETE
package api

import "github.com/labstack/echo/v4"

type Options struct {
	ServerPathPrefix string
}

func Serve(e *echo.Echo, options Options) {
	serverTroubleReports(e, options)
}
