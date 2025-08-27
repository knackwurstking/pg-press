package utils

import (
	"strings"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/labstack/echo/v4"
)

type EchoRoute struct {
	Method  string
	Path    string
	Handler echo.HandlerFunc
}

func NewEchoRoute(method string, path string, handler echo.HandlerFunc) *EchoRoute {
	return &EchoRoute{
		Method:  method,
		Path:    path,
		Handler: handler,
	}
}

func RegisterEchoRoutes(e *echo.Echo, routes []*EchoRoute) {
	for _, route := range routes {
		e.GET(constants.ServerPathPrefix+route.Path, route.Handler)
		if !strings.HasSuffix(route.Path, "/") {
			e.GET(constants.ServerPathPrefix+route.Path+"/", route.Handler)
		}
	}
}
