package main

import (
	"errors"

	"github.com/SuperPaintman/nice/cli"
)

const (
	version = "v0.0.1"
)

var errUnderConstruction = errors.New("under construction")

func main() {
	a := cli.App{
		Name: "pg-vis",
		Commands: []cli.Command{
			{
				Name:  "user",
				Usage: cli.Usage("Handle users, add, delete or modify user data in the database"),
				Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
					// TODO: Flags here...

					return func(cmd *cli.Command) error {
						// TODO: ...

						return errUnderConstruction
					}
				}),
			},

			{
				Name:  "server",
				Usage: cli.Usage("Start the server."),
				Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
					// TODO: Flags here...

					return func(cmd *cli.Command) error {
						// TODO: ...

						return errUnderConstruction
					}
				}),
			},
		},
		CommandFlags: []cli.CommandFlag{
			cli.HelpCommandFlag(),
			cli.VersionCommandFlag(version),
		},
	}

	a.HandleError(a.Run())
}
