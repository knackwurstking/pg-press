package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/SuperPaintman/nice/cli"
	"github.com/knackwurstking/pg-vis/pgvis"
)

func removeCookiesCommand() cli.Command {
	return cli.Command{
		Name: "remove",
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := cli.String(cmd, "db",
				cli.WithShort("d"),
				cli.Optional,
			)

			useApiKey := cli.Bool(cmd, "api-key",
				cli.Usage("Remove all entries containing the api-key"),
				cli.Optional)

			value := cli.StringArg(cmd, "value",
				cli.Usage("Remove entry containing the cookie value, only if `--api-key` is not set"),
				cli.Required)

			return func(cmd *cli.Command) error {
				db, err := openDB(*customDBPath)
				if err != nil {
					return err
				}

				if *useApiKey {
					err = db.Cookies.RemoveApiKey(*value)
				} else {
					err = db.Cookies.Remove(*value)
				}

				if err != nil {
					fmt.Fprintf(os.Stderr, "Removing cookies from database failed: %s", err.Error())
					os.Exit(exitCodeGeneric)
				}

				return nil
			}
		}),
	}
}

func autoCleanCookiesCommand() cli.Command {
	return cli.Command{
		Name: "auto-clean",
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := cli.String(cmd, "db",
				cli.WithShort("d"),
				cli.Optional,
			)

			telegramID := cli.Int64(cmd, "user",
				cli.WithShort("u"),
				cli.Optional,
			)

			return func(cmd *cli.Command) error {
				db, err := openDB(*customDBPath)
				if err != nil {
					return err
				}


				if *telegramID != 0 {
					u, err := db.Users.Get(*telegramID)
					if err != nil {
						if errors.Is(err, pgvis.ErrNotFound) {
							os.Exit(exitCodeNotFound)
						}

						fmt.Fprintf(os.Stderr, "Get user \"%d\" failed: %s", *telegramID, err.Error())
						os.Exit(exitCodeGeneric)
					}

					cookies, err := db.Cookies.ListApiKey(u.ApiKey)
					if err != nil {
						fmt.Fprintf(os.Stderr, "List cookies for user \"%d\" failed: %s", *telegramID, err.Error())
						os.Exit(exitCodeGeneric)
					}

					now := time.Now().UnixMilli()
					for _, cookie := range cookies {
						if now >= cookie.LastLogin {
							if err = db.Cookies.Remove(cookie.Value); err != nil {
								// Print out error and continue
								fmt.Fprintf(os.Stderr, "Removing cookie with value \"%s\" failed: %s", cookie.Value, err.Error())
							}
						}
					}

					return nil
				}

				// TODO: Auto clean up expired cookies for all users
				//db.Cookies.List()

				return errors.New("under construction")
			}
		}),
	}
}
