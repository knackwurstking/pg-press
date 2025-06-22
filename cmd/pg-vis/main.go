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

			{
				Name:  "server",
				Usage: cli.Usage("Start the server."),
				Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
					return func(cmd *cli.Command) error {
						// TODO: Server backend

						return errUnderConstruction
					}
				}),
			},

			cli.CompletionCommand(),
		},
		CommandFlags: []cli.CommandFlag{
			cli.HelpCommandFlag(),
			cli.VersionCommandFlag(version),
		},
	}

	a.HandleError(a.Run())
}
