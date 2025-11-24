// Package main command line interface for pgpress server.
//
// This file implements the server command which starts the HTTP server
// for the pgpress web application. It handles database initialization,
// middleware setup, error handling, and route configuration.
package main

import (
	"log/slog"
	"os"

	"github.com/SuperPaintman/nice/cli"
	"github.com/knackwurstking/ui"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// serverCommand creates the CLI command for starting the HTTP server.
func serverCommand() cli.Command {
	return cli.Command{
		Name:  "server",
		Usage: cli.Usage("Start the HTTP server for the pgpress web application."),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := createDBPathOption(cmd, "Custom database file path (defaults to standard location)")

			addr := cli.String(cmd, "addr",
				cli.WithShort("a"),
				cli.Usage("Set server address in format <host>:<port> (e.g., localhost:8080)"))
			*addr = serverAddress

			return func(cmd *cli.Command) error {
				initializeLogging()

				r, err := openDB(*customDBPath, true)
				if err != nil {
					slog.Error("Failed to open database", "error", err)
					return err
				}

				e := echo.New()
				e.HideBanner = true

				e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
					Output:           os.Stderr,
					Format:           "${time_custom} ${method} ${status} ${uri} ${latency_human} ${remote_ip} ${error}\n",
					CustomTimeFormat: "2006-01-02 15:04:05",
				}))

				e.Use(middlewareKeyAuth(r))
				e.Use(ui.EchoMiddlewareCache())

				Serve(e, r)

				slog.Info("Starting HTTP server", "address", *addr)
				if err := e.Start(*addr); err != nil {
					slog.Error("Server startup failed", "address", *addr, "error", err)
					slog.Error("Common causes: port already in use, permission denied, invalid address format")
					os.Exit(exitCodeServerStart)
				}

				return nil
			}
		}),
	}
}
