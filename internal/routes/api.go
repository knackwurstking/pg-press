// API Routes:
//   - apiSetupPing: 	GET 	- "/api/ping"
//   - apiSetupDevices: GET 	- "/api/devices?cache=true"
//   - apiSetupDevices: POST 	- "/api/devices/color?force=true" <- { devices: Device[]; color: number[] }
//   - apiSetupColors: 	GET 	- "/api/colors"
//   - apiSetupColors: 	GET 	- "/api/colors/:index"
//   - apiSetupColors 	POST 	- "/api/colors:index" <- `number[]`
//   - apiSetupColors: 	DELETE 	- "/api/colors/:index"
package routes

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"picow-led/internal/api"
	"strconv"

	"github.com/labstack/echo/v4"
)

type apiOptions struct {
	ServerPathPrefix string
	Config           *api.Config
}

func apiRoutes(e *echo.Echo, o apiOptions) {
	apiSetupPing(e, o)
	apiSetupDevices(e, o)
	apiSetupColor(e, o)
}

func apiSetupPing(e *echo.Echo, o apiOptions) {
	e.GET(o.ServerPathPrefix+"/api/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})
}

func apiSetupDevices(e *echo.Echo, o apiOptions) {
	e.GET(o.ServerPathPrefix+"/api/devices", func(c echo.Context) error {
		if c.QueryParam("cache") == "true" {
			return c.JSON(http.StatusOK, cache.Devices())
		}

		devices := api.GetDevices(o.Config.Devices...)
		err := c.JSON(http.StatusOK, devices)
		if err != nil {
			slog.Warn("Parse JSON", "error", err, "path", c.Request().URL.Path)
			return err
		}

		go cache.SetDevices(devices...)
		return err
	})

	e.POST(o.ServerPathPrefix+"/api/devices/color", func(c echo.Context) error {
		var reqData struct {
			Devices []*api.Device  `json:"devices"`
			Color   api.MicroColor `json:"color"`
		}
		err := json.NewDecoder(c.Request().Body).Decode(&reqData)
		if err != nil {
			slog.Warn("Decode client data", "error", err, "path", c.Request().URL.Path)
			return c.String(http.StatusBadRequest, err.Error())
		}

		handle := func() error {
			reqData.Devices = api.SetColor(reqData.Color, reqData.Devices...)

			for i, d := range reqData.Devices {
				d, err := cache.UpdateDevice(d.Server.Addr, d)
				if err != nil {
					slog.Warn("Update cached device", "error", err, "path", c.Request().URL.Path)
					return c.String(http.StatusNotFound, err.Error())
				}

				reqData.Devices[i] = d
			}

			if c.QueryParam("force") == "true" {
				return nil
			}

			err = c.JSON(http.StatusOK, reqData.Devices)
			if err != nil {
				slog.Warn("Parse JSON", "error", err, "path", c.Request().URL.Path)
				return c.String(http.StatusInternalServerError, err.Error())
			}

			return nil
		}

		if c.QueryParam("force") == "true" {
			go handle()
			return nil
		}

		return handle()
	})

}

func apiSetupColor(e *echo.Echo, o apiOptions) {
	e.GET(o.ServerPathPrefix+"/api/colors", func(c echo.Context) error {
		return c.JSON(http.StatusOK, cache.Colors())
	})

	e.GET(o.ServerPathPrefix+"/api/colors/:index", func(c echo.Context) error {
		index, err := strconv.Atoi(c.Param("index"))
		if err != nil {
			slog.Warn("Parse param \":index\"", "error", err, "path", c.Request().URL.Path)
			return c.String(http.StatusBadRequest, err.Error())
		}

		cacheColor := cache.Colors()
		if len(cacheColor)-1 < index {
			return c.String(http.StatusBadRequest, fmt.Sprintf("color index %d not found", index))
		}
		color := cacheColor[index]

		err = c.JSON(http.StatusOK, color)
		if err != nil {
			slog.Warn("Parse JSON", "error", err, "path", c.Request().URL.Path)
			return c.String(http.StatusBadRequest, err.Error())
		}

		return nil
	})

	e.POST(o.ServerPathPrefix+"/api/colors/:index", func(c echo.Context) error {
		index, err := strconv.Atoi(c.Param("index"))
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}

		var reqData api.MicroColor
		err = json.NewDecoder(c.Request().Body).Decode(&reqData)
		if err != nil {
			slog.Warn("Decode client data", "error", err, "path", c.Request().URL.Path)
			return c.String(http.StatusBadRequest, err.Error())
		}

		err = cache.UpdateColor(index, reqData)
		if err != nil {
			slog.Warn("Update (cache) color", "index", index, "data", reqData, "error", err, "path", c.Request().URL.Path)
			return c.String(http.StatusBadRequest, err.Error())
		}

		return nil
	})

	e.DELETE(o.ServerPathPrefix+"/api/colors/:index", func(c echo.Context) error {
		index, err := strconv.Atoi(c.Param("index"))
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}

		cache.DeleteColor(index)

		return nil
	})
}
