package main

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/SuperPaintman/nice/cli"
	"github.com/goforj/godump"
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
					{
						Name: "list",
						Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
							customDBPath := cli.String(cmd, "db",
								cli.Optional,
							)

							return func(cmd *cli.Command) error {
								dbPath := filepath.Join(configPath, databaseFile)

								if customDBPath != nil {
									var err error
									dbPath, err = filepath.Abs(*customDBPath)
									if err != nil {
										return err
									}
								}

								db, err := openDB(dbPath)
								if err != nil {
									return err
								}

								users, err := db.Users.List()
								if err != nil {
									return err
								}

								// TODO: Print out all users
								for _, u := range users {
									godump.Dump(u)
								}

								return errUnderConstruction
							}
						}),
					},

					{
						Name: "show",
						Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
							return func(cmd *cli.Command) error {
								// TODO: Show user information

								return errUnderConstruction
							}
						}),
					},

					{
						Name: "add",
						Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
							return func(cmd *cli.Command) error {
								// TODO: Add a new user

								return errUnderConstruction
							}
						}),
					},

					{
						Name: "remove",
						Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
							return func(cmd *cli.Command) error {
								// TODO: Remove user

								return errUnderConstruction
							}
						}),
					},

					{
						Name: "mod",
						Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
							return func(cmd *cli.Command) error {
								// TODO: Modify user

								return errUnderConstruction
							}
						}),
					},
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
