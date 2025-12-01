package utils

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/knackwurstking/pg-press/env"

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
	registerFn := func(fn func(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route, path string, handler echo.HandlerFunc) {
		fn(path, handler)
		if !strings.HasSuffix(path, "/") {
			fn(path+"/", handler)
		}
	}

	for _, route := range routes {
		path := filepath.Join(env.ServerPathPrefix, route.Path)

		var fn func(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
		switch route.Method {
		case http.MethodGet:
			fn = e.GET
		case http.MethodPost:
			fn = e.POST
		case http.MethodPut:
			fn = e.PUT
		case http.MethodDelete:
			fn = e.DELETE
		case http.MethodPatch:
			fn = e.PATCH
		default:
			panic("unhandled method: " + route.Method)
		}

		registerFn(fn, path, route.Handler)
	}
}
