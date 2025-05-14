package main

import (
	"embed"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"picow-led/internal/api"
	"picow-led/internal/routes"
	"time"

	"github.com/SuperPaintman/nice/cli"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/lmittmann/tint"
)

var (
	serverPathPrefix      = os.Getenv("SERVER_PATH_PREFIX")
	version               = "v0.11.0"
	apiConfigPath         = "api.yaml"
	apiConfigFallbackPath = ""
)

func init() {
	if d, err := os.UserConfigDir(); err == nil {
		apiConfigFallbackPath = filepath.Join(d, "picow-led", "api.yaml")
	}
}

//go:embed templates
var templates embed.FS

//go:embed public
var public embed.FS

func templatesFS() fs.FS {
	fs, err := fs.Sub(templates, "templates")
	if err != nil {
		panic(err)
	}

	return fs
}

func publicFS() fs.FS {
	fs, err := fs.Sub(public, "public")
	if err != nil {
		panic(err)
	}

	return fs
}

func main() {
	app := cli.App{
		Name:  "picow-led",
		Usage: cli.Usage("PicoW LED server and client."),
		Commands: []cli.Command{
			{
				Name:  "server",
				Usage: cli.Usage("Start the Server"),
				Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
					addr := cli.String(
						cmd, "addr",
						cli.WithShort("a"),
						cli.Usage("Set server address (<host>:<port>)"),
					)

					*addr = os.Getenv("SERVER_ADDR")

					return cliServerAction(addr)
				}),
			},
		},
		CommandFlags: []cli.CommandFlag{
			cli.HelpCommandFlag(),
			cli.VersionCommandFlag(version),
		},
	}

	app.HandleError(app.Run())
}

func cliServerAction(addr *string) cli.ActionRunner {
	return func(cmd *cli.Command) error {
		e := echo.New()

		logger := slog.New(tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.DateTime,
			AddSource:  true,
		}))
		slog.SetDefault(logger)

		// Echo: Middleware
		e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: "[${time_rfc3339}] ${status} ${method} ${path} (${remote_ip}) ${latency_human}\n",
			Output: os.Stderr,
		}))

		// Echo: Static File Server
		e.GET(serverPathPrefix+"/*", echo.StaticDirectoryHandler(publicFS(), false))

		// Api Configuration
		slog.Info("Load API configuration", "path", apiConfigPath, "fallbackPath", apiConfigFallbackPath)
		apiConfig, err := api.GetConfig(
			apiConfigPath, apiConfigFallbackPath,
		)
		if err != nil {
			slog.Warn("Read API configuration failed!")
			slog.Warn(err.Error())
		}

		routes.Create(e, routes.Options{
			ServerPathPrefix: serverPathPrefix,
			Version:          version,
			Templates:        templatesFS(),
			Config:           apiConfig,
		})

		return e.Start(*addr)
	}
}
