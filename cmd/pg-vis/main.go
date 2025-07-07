package main

import (
	"os"
	"path/filepath"

	"github.com/SuperPaintman/nice/cli"
)

const (
	appName      = "pg-vis"
	version      = "v0.0.1"
	databaseFile = "pgvis.db"

	exitCodeNotFound    = 10
	exitCodeServerStart = 20
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
				Usage: cli.Usage("Handle users, add, delete or modify user data in the database"),
				Commands: []cli.Command{
					listUserCommand(),
					showUserCommand(),
					addUserCommand(),
					removeUserCommand(),
					modUserCommand(),
				},
			},

			// TODO: Cookies command(s): "cookies" remove --api-key <api-key>
			// TODO: Cookies command(s): "cookies" remove --value <value>

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
