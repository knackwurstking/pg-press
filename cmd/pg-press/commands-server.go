// Package main command line interface for pgpress server.
//
// This file implements the server command which starts the HTTP server
// for the pgpress web application. It handles database initialization,
// middleware setup, error handling, and route configuration.
package main

import (
	"errors"
	"net/http"
	"os"

	"github.com/knackwurstking/pgpress/internal/database/dberror"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/router"

	"github.com/SuperPaintman/nice/cli"
	"github.com/labstack/echo/v4"
)

// serverCommand creates the CLI command for starting the HTTP server.
func serverCommand() cli.Command {
	return cli.Command{
		Name:  "server",
		Usage: cli.Usage("Start the HTTP server for the pgpress web application."),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := cli.String(cmd, "db",
				cli.WithShort("d"),
				cli.Usage("Custom database file path (defaults to standard location)"),
				cli.Optional)

			addr := cli.String(cmd, "addr",
				cli.WithShort("a"),
				cli.Usage("Set server address in format <host>:<port> (e.g., localhost:8080)"))
			*addr = serverAddress

			return func(cmd *cli.Command) error {
				db, err := openDB(*customDBPath)
				if err != nil {
					logger.Server().Error("Failed to open database: %v", err)
					return err
				}

				e := echo.New()
				e.HideBanner = true

				e.Use(middlewareLogger())
				e.Use(conditionalCacheMiddleware())
				e.Use(staticCacheMiddleware())
				e.Use(middlewareKeyAuth(db))
				e.HTTPErrorHandler = createHTTPErrorHandler()

				router.Serve(e, db)

				logger.Server().Info("Server listening on %s", *addr)
				if err := e.Start(*addr); err != nil &&
					err != http.ErrServerClosed {
					logger.Server().Error("Server startup failed: %v", err)
					os.Exit(exitCodeServerStart)
				}

				return nil
			}
		}),
	}
}

// createHTTPErrorHandler creates a custom HTTP error handler.
func createHTTPErrorHandler() echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		// NOTE: I hope there will never be a nil error again, but if it does, we'll handle it gracefully
		if err == nil {
			err = errors.New("unexpected nil error")
		}

		code := http.StatusInternalServerError
		var message string

		if herr, ok := err.(*echo.HTTPError); ok {
			code = herr.Code
			switch m := herr.Message.(type) {
			case string:
				message = m
			case error:
				message = m.Error()
			default:
				message = http.StatusText(code)
			}
		} else {
			code = dberror.GetHTTPStatusCode(err)
			message = err.Error()
		}

		// Note: Request logging is handled by the logger middleware
		// Only log here if we need additional context beyond the standard request log
		if code >= 500 {
			logger.Server().Debug("Error details: %s", message)
		}

		if !c.Response().Committed {
			if c.Request().Header.Get("Accept") == "application/json" {
				c.JSON(code, map[string]any{
					"error":  message,
					"code":   code,
					"status": http.StatusText(code),
				})
			} else {
				c.JSON(code, message)
			}
		}
	}
}
