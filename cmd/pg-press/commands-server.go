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
	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/handlers"

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
			customDBPath := createDBPathOption(cmd)

			_ = cli.StringVar(cmd, &env.ServerAddress, "addr",
				cli.WithShort("a"),
				cli.Usage("Set server address in format <host>:<port> (e.g., localhost:8080)"))

			return func(cmd *cli.Command) error {
				return withDBOperation(*customDBPath, true, func() error {
					e := echo.New()
					e.HideBanner = true
					e.HidePort = true

					middlewareConfiguration(e)
					setupRouter(e, env.ServerPathPrefix)
					startServer(e, env.ServerAddress)

					return nil
				})
			}
		}),
	}
}

/*******************************************************************************
 * Server Middleware Configuration
 ******************************************************************************/

func middlewareConfiguration(e *echo.Echo) {
	e.Use(middleware.RequestLogger())
	e.Use(middlewareKeyAuth())
	e.Use(ui.EchoMiddlewareCache(pages))
}

/*******************************************************************************
 * Server Route Configuration
 ******************************************************************************/

func setupRouter(e *echo.Echo, prefix string) {
	e.StaticFS(prefix+"/", assets.GetPublic())
	e.Static(prefix+"/images", env.ServerPathImages)
	handlers.RegisterAll(e)
}

/*******************************************************************************
 * Server Startup
 ******************************************************************************/

func startServer(e *echo.Echo, address string) {
	slog.Info("Starting HTTP server", "address", address)

	if err := e.Start(address); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(exitCodeServerStart)
	}
}
