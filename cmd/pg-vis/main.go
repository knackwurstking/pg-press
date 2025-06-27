package main

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/SuperPaintman/nice/cli"
)

const (
	appName      = "pg-vis"
	version      = "v0.0.1"
	databaseFile = "pgvis.db"
)

var (
	errUnderConstruction = errors.New("under construction")
	configPath           string
)

func init() {
	if p, err := os.UserConfigDir(); err != nil {
		panic(err)
	} else {
		configPath = filepath.Join(p, appName)
	}
}

func main() {
	a := cli.App{
		Name: appName,
		Commands: []cli.Command{
			{
				Name: "api-key",
				Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
					return func(cmd *cli.Command) error {
						// TODO: Generate a new unique api key

						return errUnderConstruction
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
