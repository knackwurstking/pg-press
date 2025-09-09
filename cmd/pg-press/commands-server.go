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

	"github.com/SuperPaintman/nice/cli"
	"github.com/knackwurstking/pgpress/internal/database/dberror"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/router"
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

			logFile := cli.String(cmd, "log-file",
				cli.WithShort("l"),
				cli.Usage("Path to log file"),
				cli.Optional)

			return func(cmd *cli.Command) error {
				if *logFile != "" {
					logger.Server().Info("Opening log file: %s", *logFile)
					f, err := os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						logger.AppLogger.Error("Failed to open log file %s: %v", *logFile, err)
						return err
					} else {
						logger.SetOutput(f)
						logger.Server().Info("Successfully redirected logs to file: %s", *logFile)
					}
				}

				dbPath := *customDBPath
				if dbPath == "" {
					dbPath = "default location"
				}
				logger.Server().Info("Opening database at: %s", dbPath)

				db, err := openDB(*customDBPath)
				if err != nil {
					logger.Server().Error("Failed to open database at %s: %v", dbPath, err)
					return err
				}
				logger.Server().Info("Database opened successfully")

				logger.Server().Info("Initializing Echo web server")
				e := echo.New()
				e.HideBanner = true
				e.HTTPErrorHandler = createHTTPErrorHandler()

				logger.Server().Debug("Configuring middleware stack")
				e.Use(middlewareLogger())
				e.Use(conditionalCacheMiddleware())
				e.Use(staticCacheMiddleware())
				e.Use(middlewareKeyAuth(db))

				logger.Server().Debug("Configuring routes")
				router.Serve(e, db)

				logger.Server().Info("Starting HTTP server on %s", *addr)
				logger.Server().Info("Server ready to accept connections")
				if err := e.Start(*addr); err != nil &&
					err != http.ErrServerClosed {
					logger.Server().Error("Server startup failed on %s: %v", *addr, err)
					logger.Server().Error("Common causes: port already in use, permission denied, invalid address format")
					os.Exit(exitCodeServerStart)
				}
				logger.Server().Info("Server shutdown complete")

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
			logger.Server().Error("HTTP error handler called with nil error - this should not happen")
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
			logger.Server().Debug("Non-HTTP error converted to status %d: %v", code, err)
		}

		// Note: Request logging is handled by the logger middleware
		// Only log here if we need additional context beyond the standard request log
		if code >= 500 {
			logger.Server().Error("Internal server error (%d): %s", code, message)
			// Log additional context for server errors
			if he, ok := err.(*echo.HTTPError); ok && he.Internal != nil {
				logger.Server().Error("Internal error cause: %v", he.Internal)
			}
		} else if code >= 400 {
			logger.Server().Debug("Client error (%d): %s", code, message)
		}

		// This line checks if the HTTP response headers have already been written and sent to the client.
		// The `c.Response().Committed` property is true after the response headers are sent.
		// This check is a crucial safeguard to prevent the server from trying to send a new
		// error response if another response has already been started or completed.
		// Attempting to write headers twice would cause a panic.
		if !c.Response().Committed {
			if c.Request().Header.Get("Accept") == "application/json" {
				c.JSON(code, map[string]any{
					"error":  message,
					"code":   code,
					"status": http.StatusText(code),
				})
			} else {
				c.String(code, message)
			}
		} else {
			logger.Server().Error("Could not send error response because response was already committed (status: %d, error: %s)", code, message)
		}
	}
}
