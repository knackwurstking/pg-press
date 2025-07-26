package main

import (
	"os"
	"path/filepath"

	"github.com/SuperPaintman/nice/cli"

	"github.com/knackwurstking/pg-vis/internal/logger"
)

const (
	appName      = "pg-vis"
	version      = "v0.0.1"
	databaseFile = "pgvis.db"

	exitCodeGeneric     = 1
	exitCodeNotFound    = 10
	exitCodeServerStart = 20 // exitCodeServerStart failed
)

var (
	configPath       string
	serverPathPrefix = os.Getenv("SERVER_PATH_PREFIX")
	serverAddress    = os.Getenv("SERVER_ADDR")
)

func init() {
	if p, err := os.UserConfigDir(); err != nil {
		panic(err)
	} else {
		configPath = filepath.Join(p, appName)
		if err := os.MkdirAll(configPath, 0700); err != nil {
			panic(err)
		}
	}
}

func main() {
	// Initialize colored logger
	logger.Initialize()

	// Configure logger based on environment
	if os.Getenv("DEBUG") != "" {
		logger.SetupDevelopment()
	} else if os.Getenv("PRODUCTION") != "" {
		logger.SetupProduction()
	}

	a := cli.App{
		Name: appName,
		Usage: cli.Usage(`Exit Codes:
  1   Generic
  10  Not Found
`),
		Commands: []cli.Command{
			apiKeyCommand(),

			{
				Name:  "user",
				Usage: cli.Usage("Handle users database table, add, remove or modify user data"),
				Commands: []cli.Command{
					listUserCommand(),
					showUserCommand(),
					addUserCommand(),
					removeUserCommand(),
					modUserCommand(),
				},
			},

			{
				Name:  "cookies",
				Usage: cli.Usage("Handle cookies database table, remove cookies data"),
				Commands: []cli.Command{
					removeCookiesCommand(),
					autoCleanCookiesCommand(),
				},
			},

			serverCommand(),

			cli.CompletionCommand(),
		},
		CommandFlags: []cli.CommandFlag{
			cli.HelpCommandFlag(),
			cli.VersionCommandFlag(version),
		},
	}

	a.HandleError(a.Run())
}
