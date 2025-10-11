package helpers

import (
	"net/http"
	"strings"

	"github.com/knackwurstking/pgpress/internal/env"
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
		switch route.Method {
		case http.MethodGet:
			e.GET(env.ServerPathPrefix+route.Path, route.Handler)
			if !strings.HasSuffix(route.Path, "/") {
				e.GET(env.ServerPathPrefix+route.Path+"/", route.Handler)
			}
		case http.MethodPost:
			e.POST(env.ServerPathPrefix+route.Path, route.Handler)
			if !strings.HasSuffix(route.Path, "/") {
				e.POST(env.ServerPathPrefix+route.Path+"/", route.Handler)
			}
		case http.MethodPut:
			e.PUT(env.ServerPathPrefix+route.Path, route.Handler)
			if !strings.HasSuffix(route.Path, "/") {
				e.PUT(env.ServerPathPrefix+route.Path+"/", route.Handler)
			}
		case http.MethodDelete:
			e.DELETE(env.ServerPathPrefix+route.Path, route.Handler)
			if !strings.HasSuffix(route.Path, "/") {
				e.DELETE(env.ServerPathPrefix+route.Path+"/", route.Handler)
			}
		case http.MethodPatch:
			e.PATCH(env.ServerPathPrefix+route.Path, route.Handler)
			if !strings.HasSuffix(route.Path, "/") {
				e.PATCH(env.ServerPathPrefix+route.Path+"/", route.Handler)
			}
		default:
			panic("unhandled method: " + route.Method)
		}
	}
}
