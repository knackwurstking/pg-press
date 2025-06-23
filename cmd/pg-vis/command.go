package main

import (
	"os"

	"github.com/SuperPaintman/nice/cli"
	"github.com/jedib0t/go-pretty/v6/table"
)

func listUserCommand() cli.Command {
	return cli.Command{
		Name: "list",
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := cli.String(cmd, "db",
				cli.WithShort("d"),
				cli.Optional,
			)

			return func(cmd *cli.Command) error {
				db, err := openDB(customDBPath)
				if err != nil {
					return err
				}

				users, err := db.Users.List()
				if err != nil {
					return err
				}

				t := table.NewWriter()
				t.SetOutputMirror(os.Stdout)
				t.AppendHeader(table.Row{"Telegram ID", "User Name"})

				rows := []table.Row{}
				for _, u := range users {
					rows = append(rows, table.Row{u.TelegramID, u.UserName})
				}

				t.AppendRows(rows)
				t.SetStyle(table.StyleLight)
				t.Render()

				return nil
			}
		}),
	}
}

func showUserCommand() cli.Command {
	return cli.Command{
		Name: "show",
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := cli.String(cmd, "db",
				cli.WithShort("d"),
				cli.Optional,
			)

			return func(cmd *cli.Command) error {
				db, err := openDB(customDBPath)
				if err != nil {
					return err
				}

				// TODO: Show user information

				return errUnderConstruction
			}
		}),
	}
}

func addUserCommand() cli.Command {
	return cli.Command{
		Name: "add",
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			return func(cmd *cli.Command) error {
				// TODO: Add a new user

				return errUnderConstruction
			}
		}),
	}
}

func removeUserCommand() cli.Command {
	return cli.Command{
		Name: "remove",
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			return func(cmd *cli.Command) error {
				// TODO: Remove user

				return errUnderConstruction
			}
		}),
	}
}

func modUserCommand() cli.Command {
	return cli.Command{
		Name: "mod",
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			return func(cmd *cli.Command) error {
				// TODO: Modify user

				return errUnderConstruction
			}
		}),
	}
}

func serverCommand() cli.Command {
	return cli.Command{
		Name:  "server",
		Usage: cli.Usage("Start the server."),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			return func(cmd *cli.Command) error {
				// TODO: Server backend

				return errUnderConstruction
			}
		}),
	}
}
