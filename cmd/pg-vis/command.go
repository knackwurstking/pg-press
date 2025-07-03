package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/SuperPaintman/nice/cli"
	"github.com/charmbracelet/log"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/knackwurstking/pg-vis/html"
	"github.com/knackwurstking/pg-vis/pkg/pgvis"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

				// NOTE: Here i could print out some more user related stuff
				// 		 like last activity, or whatever

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

				// FIXME: Find a better way to to this
				skipperRegExp := regexp.MustCompile(
					`(.*/login.*|.*pico.lime.min.css|manifest.json|.*\.png|.*\.ico)`,
				)

				e := echo.New()

				// Init logger
				log.SetLevel(log.DebugLevel)

				e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
					Format: "${custom} ---> ${status} ${method} ${uri} (${remote_ip}) ${latency_human}\n",
					Output: os.Stderr,
					CustomTagFunc: func(c echo.Context, buf *bytes.Buffer) (int, error) {
						t := time.Now()
						buf.Write(fmt.Appendf(nil,
							"%d/%02d/%02d %02d:%02d:%02d",
							t.Year(), int(t.Month()), t.Day(),
							t.Hour(), t.Minute(), t.Second(),
						))

						return 0, nil
					},
				}))

				e.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
					Skipper: func(c echo.Context) bool {
						url := c.Request().URL.String()
						log.Debugf("Auth: Skipper: %s", url)

						return skipperRegExp.MatchString(url)
					},

					KeyLookup: "header:" + echo.HeaderAuthorization + ",query:access_token,cookie:" + html.CookieName,

					AuthScheme: "Bearer",

					Validator: func(auth string, c echo.Context) (bool, error) {
						log.Debugf("Auth: Validator: %s", c.Request().UserAgent())

						if cookie, err := c.Cookie(html.CookieName); err == nil {
							c, err := db.Cookies.Get(cookie.Value)
							if err == nil {
								log.Debugf("Auth: Validator: cookie found")
								auth = c.ApiKey
							}
						}

						user, err := db.Users.GetUserFromApiKey(auth)
						if err != nil {
							return false, fmt.Errorf("get user from db: %s (%#v)", err.Error(), auth)
						}

						return user.ApiKey == auth, nil
					},

					ErrorHandler: func(err error, c echo.Context) error {
						log.Debugf("Auth ErrorHandler: %s", err.Error())

						if err != nil {
							return c.Redirect(http.StatusSeeOther, serverPathPrefix+"/login")
						}

						return nil
					},
				}))

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
