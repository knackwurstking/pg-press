package main

import (
	"fmt"
	"os"
	"time"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/pkg/models"

	"github.com/SuperPaintman/nice/cli"
)

func removeCookiesCommand() cli.Command {
	return cli.Command{
		Name: "remove",
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := createDBPathOption(cmd, "")
			useApiKey := cli.Bool(cmd, "api-key",
				cli.Usage("Remove all entries containing the api-key"),
				cli.Optional)
			value := cli.StringArg(cmd, "value",
				cli.Usage("Remove entry containing the cookie value, only if `--api-key` is not set"),
				cli.Required)

			return func(cmd *cli.Command) error {
				return withDBOperation(customDBPath, func(db *database.DB) error {
					var err error
					if *useApiKey {
						err = db.Cookies.RemoveApiKey(*value)
					} else {
						err = db.Cookies.Remove(*value)
					}

					if err != nil {
						handleGenericError(err, "Error")
					}

					return nil
				})
			}
		}),
	}
}

func autoCleanCookiesCommand() cli.Command {
	return cli.Command{
		Name: "auto-clean",
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := createDBPathOption(cmd, "")
			telegramID := cli.Int64(cmd, "user",
				cli.WithShort("u"),
				cli.Optional,
			)

			return func(cmd *cli.Command) error {
				return withDBOperation(customDBPath, func(db *database.DB) error {
					t := time.Now().Add(0 - constants.CookieExpirationDuration).UnixMilli()
					isExpired := func(cookie *models.Cookie) bool {
						return t >= cookie.LastLogin
					}

					// Clean up cookies for a specific telegram user
					if *telegramID != 0 {
						u, err := db.Users.Get(*telegramID)
						if err != nil {
							handleNotFoundError(err)
							handleGenericError(err, fmt.Sprintf("Get user \"%d\" failed", *telegramID))
						}

						cookies, err := db.Cookies.ListApiKey(u.ApiKey)
						if err != nil {
							handleGenericError(err, fmt.Sprintf("List cookies for user \"%d\" failed", *telegramID))
						}

						for _, cookie := range cookies {
							if isExpired(cookie) {
								if err = db.Cookies.Remove(cookie.Value); err != nil {
									// Print out error and continue
									fmt.Fprintf(os.Stderr, "Removing cookie with value \"%s\" failed: %s\n", cookie.Value, err.Error())
								}
							}
						}

						return nil
					}

					// Clean up all cookies
					cookies, err := db.Cookies.List()
					if err != nil {
						handleGenericError(err, "List cookies from database failed")
					}

					for _, cookie := range cookies {
						if isExpired(cookie) {
							if err = db.Cookies.Remove(cookie.Value); err != nil {
								// Print out error and continue
								fmt.Fprintf(os.Stderr, "Removing cookie with value \"%s\" failed: %s\n", cookie.Value, err.Error())
							}
						}
					}

					return nil
				})
			}
		}),
	}
}
