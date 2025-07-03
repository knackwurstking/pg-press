package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/SuperPaintman/nice/cli"
	"github.com/charmbracelet/log"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/knackwurstking/pg-vis/html"
	"github.com/knackwurstking/pg-vis/pkg/pgvis"
	"github.com/labstack/echo/v4"
)

func apiKeyCommand() cli.Command {
	return cli.Command{
		Name: "api-key",
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			return func(cmd *cli.Command) error {
				// TODO: Generate a new unique api key
				fmt.Fprintf(os.Stdout, "<api-key>")

				return nil
			}
		}),
	}
}

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

				cookies, err := db.Cookies.GetForApiKey(user.ApiKey)
				if err != nil {
					t := table.NewWriter()
					t.SetOutputMirror(os.Stdout)
					t.AppendHeader(table.Row{"Api Key", "User Agent"})

					rows := []table.Row{}
					for _, c := range cookies {
						rows = append(rows, table.Row{c.ApiKey, c.UserAgent})
					}

					t.AppendRows([]table.Row{row})
					t.SetStyle(table.StyleLight)
					t.Render()
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

func serverCommand() cli.Command {
	return cli.Command{
		Name:  "server",
		Usage: cli.Usage("Start the server."),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := cli.String(cmd, "db",
				cli.WithShort("d"),
				cli.Optional,
			)

			addr := cli.String(
				cmd, "addr",
				cli.WithShort("a"),
				cli.Usage("Set server address (<host>:<port>)"),
			)
			*addr = serverAddress

			return func(cmd *cli.Command) error {
				db, err := openDB(*customDBPath)
				if err != nil {
					return err
				}

				e := echo.New()

				// Init logger
				log.SetLevel(log.DebugLevel)

				e.Use(middlewareLogger())
				e.Use(middlewareKeyAuth(db))

				e.HTTPErrorHandler = func(err error, c echo.Context) {
					log.Warnf("HTTPErrorHandler: %s", err.Error())

					// NOTE: Maybe serve an error page here instead
					if herr, ok := err.(*echo.HTTPError); ok {
						message := http.StatusText(herr.Code)

						if m, ok := herr.Message.(string); ok {
							message = m
						} else if e, ok := herr.Message.(error); ok {
							message = e.Error()
						}

						c.JSON(herr.Code, message)
						return
					}

					c.JSON(500, err.Error())
				}

				html.Serve(e, html.Options{
					ServerPathPrefix: serverPathPrefix,
					DB:               db,
				})

				if err := e.Start(*addr); err != nil {
					log.Errorf("Starting the server failed: %s", err.Error())
					os.Exit(exitCodeServerStart)
				}

				return nil
			}
		}),
	}
}
