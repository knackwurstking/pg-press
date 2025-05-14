// Frontend Routes:
//   - GET - "/"
//   - GET - "/devices/:addr"
//   - GET - "/settings"
//   - GET - "/manifest.json"
package routes

import (
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"picow-led/internal/api"

	"github.com/labstack/echo/v4"
)

const (
	contentDevices     content = "devices"
	contentDevicesAddr content = "devices-addr"
	contentSettings    content = "settings"
)

type content string

type frontendOptions struct {
	ServerPathPrefix string
	Version          string
	Templates        fs.FS
}

func (f *frontendOptions) BasicPatterns() []string {
	return []string{
		"components/online-indicator.go.html",
		"components/svg-power.go.html",
	}
}

// serve template data
func (f *frontendOptions) serve(c echo.Context, pattern string, mimeType string, data frontendTemplateData) error {
	t, err := template.ParseFS(f.Templates, pattern)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	c.Response().Header().Add("Content-Type", mimeType)
	err = t.Execute(c.Response().Writer, data)
	if err != nil {
		slog.Error("Execute template",
			"error", err, "path", c.Request().URL.Path)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

// servePage template data
func (f *frontendOptions) servePage(c echo.Context, content content, data frontendTemplateData) error {
	patterns := []string{
		"main.go.html",         // There is only one page for now
		"layouts/base.go.html", // There is also only on layout for now
		fmt.Sprintf("pages/%s.go.html", content),
	}
	patterns = append(patterns, f.BasicPatterns()...)

	t, err := template.ParseFS(f.Templates, patterns...)
	if err != nil {
		slog.Error("ParseFS (template) patterns", "error", err, "path", c.Request().URL.Path)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	c.Response().Header().Add("Content-Type", "text/html; charset=utf-8")
	err = t.Execute(c.Response().Writer, data)
	if err != nil {
		slog.Error("Execute template", "error", err, "path", c.Request().URL.Path)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

type frontendTemplateData struct {
	ServerPathPrefix string
	Version          string
	Title            string
}

func frontendRoutes(e *echo.Echo, o frontendOptions) {
	e.GET(o.ServerPathPrefix+"/", func(c echo.Context) error {
		err := o.servePage(c, contentDevices, frontendTemplateData{
			ServerPathPrefix: o.ServerPathPrefix,
			Version:          o.Version,
			Title:            "PicoW LED | Devices",
		})
		if err != nil {
			slog.Error(err.Error(), "path", c.Request().URL.Path)
		}
		return err
	})

	e.GET(o.ServerPathPrefix+"/devices/:addr", func(c echo.Context) error {
		addr, err := url.QueryUnescape(c.Param("addr"))
		if err != nil {
			slog.Warn("Query unescape", "error", err, "path", c.Request().URL.Path)
			return c.String(http.StatusBadRequest, err.Error())
		}

		var device *api.Device
		for _, d := range cache.Devices() {
			if d.Server.Addr == addr {
				device = d
				break
			}
		}

		if device == nil {
			msg := fmt.Sprintf("Device \"%s\" not found", addr)
			slog.Warn(msg, "error", err, "path", c.Request().URL.Path)
			return c.String(http.StatusNotFound, msg)
		}

		return o.servePage(c, contentDevicesAddr, frontendTemplateData{
			ServerPathPrefix: o.ServerPathPrefix,
			Version:          o.Version,
			Title:            fmt.Sprintf("PicoW LED | %s", addr),
		})
	})

	e.GET(o.ServerPathPrefix+"/settings", func(c echo.Context) error {
		return o.servePage(c, contentSettings, frontendTemplateData{
			ServerPathPrefix: o.ServerPathPrefix,
			Version:          o.Version,
			Title:            "PicoW LED | Settings",
		})
	})

	// PWA Stuff here
	e.GET(o.ServerPathPrefix+"/manifest.json", func(c echo.Context) error {
		return o.serve(c, "pwa/manifest.json", "application/json", frontendTemplateData{
			ServerPathPrefix: o.ServerPathPrefix,
			Version:          o.Version,
			Title:            "PicoW LED | Settings",
		})
	})
}
