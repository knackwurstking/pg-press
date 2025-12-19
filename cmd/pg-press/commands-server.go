// Package main command line interface for pgpress server.
//
// This file implements the server command which starts the HTTP server
// for the pgpress web application. It handles database initialization,
// middleware setup, error handling, and route configuration.
package main

import (
	"bytes"
	"fmt"
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

			_ = cli.StringVar(cmd, &env.ServerAddress, "addr",
				cli.WithShort("a"),
				cli.Usage("Set server address in format <host>:<port> (e.g., localhost:8080)"))

			return func(cmd *cli.Command) error {
				db, err := openDB(*customDBPath, true)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
					return err
				}

				e := echo.New()
				e.HideBanner = true

				middlewareConfiguration(e, db)
				setupRouter(e, db, env.ServerPathPrefix)
				startServer(e, env.ServerAddress)

				return nil
			}
		}),
	}
}

/*******************************************************************************
 * Server Middleware Configuration
 ******************************************************************************/

func middlewareConfiguration(e *echo.Echo, db *common.DB) {
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Output:           os.Stderr,
		Format:           "${time_custom} ${method} ${status} ${uri} ${latency_human} ${remote_ip} ${error} ${custom}\n",
		CustomTimeFormat: "2006-01-02 15:04:05",
		CustomTagFunc: func(c echo.Context, b *bytes.Buffer) (int, error) {
			if c.Get("user-name") == nil || c.Get("user-name").(string) == "" {
				return b.WriteString("user-name=\"anonymous\"")
			}
			return fmt.Fprintf(b, "user-name=\"%s\" ", c.Get("user-name").(string))
		},
	}))

	e.Use(middlewareKeyAuth(db))
	e.Use(ui.EchoMiddlewareCache(pages))
}

/*******************************************************************************
 * Server Route Configuration
 ******************************************************************************/

func setupRouter(e *echo.Echo, db *common.DB, prefix string) {
	// Static File Server
	e.StaticFS(prefix+"/", assets.GetAssets())
	handlers.RegisterAll(db, e)
}

/*******************************************************************************
 * Server Startup
 ******************************************************************************/

func startServer(e *echo.Echo, address string) {
	log.Info("Starting HTTP server at %#v", address)

	if err := e.Start(address); err != nil {
		log.Error("Failed to start server: %v", err)
		os.Exit(exitCodeServerStart)
	}
}
