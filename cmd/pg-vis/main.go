package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SuperPaintman/nice/cli"
)

const (
	appName      = "pg-vis"
	version      = "v0.0.1"
	databaseFile = "pgvis.db"

	exitCodeNotFound = 10
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
			{
				Name: "api-key",
				Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
					return func(cmd *cli.Command) error {
						// TODO: Generate a new unique api key
						fmt.Fprintf(os.Stdout, "<api-key>")

						return nil
					}
				}),
			},

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
