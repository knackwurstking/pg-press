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

	"github.com/knackwurstking/pgpress/internal/web"
	"github.com/knackwurstking/pgpress/pkg/utils"

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

			logFile := cli.String(cmd, "log-file",
				cli.WithShort("l"),
				cli.Usage("Path to log file"),
				cli.Optional)

			return func(cmd *cli.Command) error {
				if *logFile != "" {
					f, err := os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						log().Error("Failed to open log file %s: %v", *logFile, err)
						return err
					} else {
						log().SetOutput(f)
						log().Info("Redirected logs to file: %s", *logFile)
					}
				}

				db, err := openDB(*customDBPath)
				if err != nil {
					log().Error("Failed to open database: %v", err)
					return err
				}

				e := echo.New()
				e.HideBanner = true
				e.HTTPErrorHandler = createHTTPErrorHandler()

				e.Use(middlewareLogger())
				e.Use(conditionalCacheMiddleware())
				e.Use(staticCacheMiddleware())
				e.Use(middlewareKeyAuth(db))

				web.Serve(e, db)

				log().Info("Starting HTTP server on %s", *addr)
				if err := e.Start(*addr); err != nil &&
					err != http.ErrServerClosed {
					log().Error("Server startup failed on %s: %v", *addr, err)
					log().Error("Common causes: port already in use, permission denied, invalid address format")
					os.Exit(exitCodeServerStart)
				}

				return nil
			}
		}),
	}
}

// createHTTPErrorHandler creates a custom HTTP error handler with comprehensive logging.
//
// The handler provides:
// - Detailed error logging with request context (method, URI, remote IP, user agent)
// - Status-appropriate log levels (ERROR for 5xx, INFO/WARN for 4xx, DEBUG for 404 GET)
// - Smart client error filtering to reduce log noise
// - JSON or plain text responses based on Accept header
// - Protection against double-header writing
func createHTTPErrorHandler() echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		// Handle nil error case (should never happen, but be defensive)
		if err == nil {
			log().Error("HTTP error handler received nil error - " +
				"this indicates a bug in the application")

			err = errors.New("unexpected nil error")
		}

		// Extract error details
		var code int
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
			code = utils.GetHTTPStatusCode(err)
			message = err.Error()
		}

		// Get request context for logging
		req := c.Request()
		remoteIP := c.RealIP()
		method := req.Method
		uri := req.RequestURI
		userAgent := req.UserAgent()

		if code >= 500 {
			log().Error("HTTP %d: %s [%s %s] from %s (UA: %s)",
				code, message, method, uri, remoteIP, userAgent)
		} else if code >= 400 {
			switch code {
			case 401, 403:
				log().Warn(
					"HTTP %d: Authentication/Authorization failed [%s %s] from %s",
					code, method, uri, remoteIP)
			case 404:
				log().Warn("HTTP %d: Not found [%s %s] from %s",
					code, method, uri, remoteIP)
			default:
				log().Warn("HTTP %d: %s [%s %s] from %s",
					code, message, method, uri, remoteIP)
			}
		} else {
			log().Warn("HTTP %d: %s [%s %s] from %s",
				code, message, method, uri, remoteIP)
		}

		// Send error response if headers haven't been written yet
		if !c.Response().Committed {
			if c.Request().Header.Get("Accept") == "application/json" {
				if err := c.JSON(code, map[string]any{
					"error":  message,
					"code":   code,
					"status": http.StatusText(code),
				}); err != nil {
					log().Error("Failed to send JSON error response: %v", err)
				}
			} else {
				if err := c.String(code, message); err != nil {
					log().Error("Failed to send string error response: %v", err)
				}
			}
		} else {
			log().Error(
				"Cannot send HTTP %d response - "+
					"headers already committed [%s %s] from %s",
				code, method, uri, remoteIP)
		}
	}
}
