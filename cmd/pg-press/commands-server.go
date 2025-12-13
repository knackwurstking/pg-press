// Package main command line interface for pgpress server.
//
// This file implements the server command which starts the HTTP server
// for the pgpress web application. It handles database initialization,
// middleware setup, error handling, and route configuration.
package main

import (
	"log/slog"
	"os"

	"github.com/knackwurstking/pg-press/internal/assets"
	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/handlers"

	ui "github.com/knackwurstking/ui/ui-templ"

	"github.com/SuperPaintman/nice/cli"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// serverCommand creates the CLI command for starting the HTTP server.
func serverCommand() cli.Command {
	return cli.Command{
		Name:  "server",
		Usage: cli.Usage("Start the HTTP server for the pgpress web application."),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := createDBPathOption(cmd)

			cli.StringVar(cmd, &env.ServerAddress, "addr",
				cli.WithShort("a"),
				cli.Usage("Set server address in format <host>:<port> (e.g., localhost:8080)"))

			return func(cmd *cli.Command) error {
				initializeLogging()

				r, err := openDB(*customDBPath, true)
				if err != nil {
					slog.Error("Failed to open database", "error", err)
					return err
				}

				e := echo.New()
				e.HideBanner = true

				middlewareConfiguration(e, r)
				setupRouter(e, r, env.ServerPathPrefix)
				startServer(e, r, env.ServerAddress)

				return nil
			}
		}),
	}
}

/*******************************************************************************
 * Server Middleware Configuration
 ******************************************************************************/

func middlewareConfiguration(e *echo.Echo, r *common.DB) {
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Output:           os.Stderr,
		Format:           "${time_custom} ${method} ${status} ${uri} ${latency_human} ${remote_ip} ${error}\n",
		CustomTimeFormat: "2006-01-02 15:04:05",
	}))

	e.Use(middlewareKeyAuth(r))
	e.Use(ui.EchoMiddlewareCache(pages))
}

/*******************************************************************************
 * Server Route Configuration
 ******************************************************************************/

func setupRouter(e *echo.Echo, r *common.DB, prefix string) {
	// Static File Server
	e.StaticFS(prefix+"/", assets.GetAssets())
	handlers.RegisterAll(r, e)
}

/*******************************************************************************
 * Server Startup
 ******************************************************************************/

func startServer(e *echo.Echo, r *common.DB, address string) {
	slog.Info("Starting HTTP server", "address", address)
	if err := e.Start(address); err != nil {
		slog.Error("Server startup failed", "address", address, "error", err)
		slog.Error("Common causes: port already in use, permission denied, invalid address format")
		os.Exit(exitCodeServerStart)
	}
}
