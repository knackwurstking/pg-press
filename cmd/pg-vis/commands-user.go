package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/SuperPaintman/nice/cli"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/knackwurstking/pg-vis/pkg/pgvis"
	"github.com/labstack/gommon/color"
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
				db, err := openDB(*customDBPath)
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
				cli.Optional)

			flagApiKey := cli.Bool(cmd, "api-key",
				cli.Optional)

			telegramID := cli.Int64Arg(cmd, "telegram-id", cli.Required)

			return func(cmd *cli.Command) error {
				db, err := openDB(*customDBPath)
				if err != nil {
					return err
				}

				user, err := db.Users.Get(*telegramID)
				if err != nil {
					if errors.Is(err, pgvis.ErrNotFound) {
						os.Exit(exitCodeNotFound)
					}

					return err
				}

				if *flagApiKey {
					fmt.Fprint(os.Stdout, user.ApiKey)
					return nil
				}

				t := table.NewWriter()
				t.SetOutputMirror(os.Stdout)

				t.AppendHeader(table.Row{"Telegram ID", "User Name", "Api Key"})

				row := table.Row{user.TelegramID, user.UserName, user.ApiKey}

				t.AppendRows([]table.Row{row})
				t.SetStyle(table.StyleLight)
				t.Render()

				if cookies, err := db.Cookies.GetForApiKey(user.ApiKey); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to get cookies from the database: %s\n", err.Error())
				} else {
					if len(cookies) > 0 {
						fmt.Printf(
							"\n%s <last-login> - <api-key> - <value> - <user-agent>\n\n",
							color.Underline(color.Bold("Cookies:")),
						)

						for _, c := range cookies {
							fmt.Printf(
								"%d - %s - %s - \"%s\"\n",
								c.LastLogin,
								color.Bold(c.ApiKey),
								c.Value,
								color.Italic(c.UserAgent),
							)
						}
					}
				}

				return nil
			}
		}),
	}
}

func addUserCommand() cli.Command {
	return cli.Command{
		Name: "add",
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := cli.String(cmd, "db",
				cli.WithShort("d"),
				cli.Optional,
			)

			apiKey := cli.String(cmd, "api-key",
				cli.Optional)

			telegramID := cli.Int64Arg(cmd, "telegram-id", cli.Required)
			userName := cli.StringArg(cmd, "user-name", cli.Optional)

			return func(cmd *cli.Command) error {
				db, err := openDB(*customDBPath)
				if err != nil {
					return err
				}

				user := pgvis.NewUser(*telegramID, *userName, *apiKey)

				if *apiKey != "" {
					user.ApiKey = *apiKey
				}

				err = db.Users.Add(user)
				if errors.Is(err, pgvis.ErrAlreadyExists) {
					return fmt.Errorf("user already exists: %d (%s)",
						*telegramID, *userName)
				}

				return err
			}
		}),
	}
}

func removeUserCommand() cli.Command {
	return cli.Command{
		Name: "remove",
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := cli.String(cmd, "db",
				cli.WithShort("d"),
				cli.Optional,
			)

			telegramID := cli.Int64Arg(cmd, "telegram-id", cli.Required)

			return func(cmd *cli.Command) error {
				db, err := openDB(*customDBPath)
				if err != nil {
					return err
				}

				db.Users.Remove(*telegramID)

				return nil
			}
		}),
	}
}

func modUserCommand() cli.Command {
	return cli.Command{
		Name: "mod",
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := cli.String(cmd, "db",
				cli.WithShort("d"),
				cli.Optional,
			)

			userName := cli.String(cmd, "name",
				cli.WithShort("n"),
				cli.Optional)

			apiKey := cli.String(cmd, "api-key",
				cli.Optional)

			deleteApiKey := cli.Bool(cmd, "delete-api-key",
				cli.Optional)

			telegramID := cli.Int64Arg(cmd, "telegram-id", cli.Required)

			return func(cmd *cli.Command) error {
				db, err := openDB(*customDBPath)
				if err != nil {
					return err
				}

				user, err := db.Users.Get(*telegramID)
				if err != nil {
					return err
				}

				if *userName != "" {
					user.UserName = *userName
				}

				if *apiKey != "" {
					user.ApiKey = *apiKey
				}

				if *deleteApiKey {
					user.ApiKey = ""
				}

				err = db.Users.Update(*telegramID, user)
				return err
			}
		}),
	}
}
