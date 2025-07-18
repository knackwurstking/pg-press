// NOTE: Cleaned up by AI
//
// Package main command line interface for pg-vis server.
//
// This file implements the server command which starts the HTTP server
// for the pg-vis web application. It handles database initialization,
// middleware setup, error handling, and route configuration.
package main

import (
	"net/http"
	"os"

	"github.com/SuperPaintman/nice/cli"
	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes"
)

// serverCommand creates the CLI command for starting the HTTP server.
func serverCommand() cli.Command {
	return cli.Command{
		Name:  "server",
		Usage: cli.Usage("Start the HTTP server for the pg-vis web application."),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := cli.String(cmd, "db",
				cli.WithShort("d"),
				cli.Usage("Custom database file path (defaults to standard location)"),
				cli.Optional,
			)

			addr := cli.String(cmd, "addr",
				cli.WithShort("a"),
				cli.Usage("Set server address in format <host>:<port> (e.g., localhost:8080)"),
			)
			*addr = serverAddress

			return func(cmd *cli.Command) error {
				db, err := openDB(*customDBPath)
				if err != nil {
					log.Errorf("Failed to open database: %v", err)
					return err
				}

				e := echo.New()
				e.HideBanner = true

				log.SetLevel(log.DebugLevel)
				log.SetReportCaller(true)
				log.SetReportTimestamp(false)

				e.Use(middlewareLogger())
				e.Use(middlewareKeyAuth(db))
				e.HTTPErrorHandler = createHTTPErrorHandler()

				routes.Serve(e, routes.Options{
					ServerPathPrefix: serverPathPrefix,
					DB:               db,
				})

				log.Infof("Server listening on %s", *addr)
				if err := e.Start(*addr); err != nil && err != http.ErrServerClosed {
					log.Errorf("Server startup failed: %v", err)
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
		if err == nil {
			return
		}

		code := http.StatusInternalServerError
		message := "Internal server error"

		if herr, ok := err.(*echo.HTTPError); ok {
			if herr == nil {
				return
			}

			code = herr.Code
			switch msg := herr.Message.(type) {
			case string:
				message = msg
			case error:
				message = msg.Error()
			default:
				message = http.StatusText(code)
			}
		} else {
			code = pgvis.GetHTTPStatusCode(err)
			message = err.Error()
		}

		if code >= 500 {
			log.Errorf("Server error (%d): %v", code, err)
		} else if code >= 400 {
			log.Warnf("Client error (%d): %v", code, err)
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
