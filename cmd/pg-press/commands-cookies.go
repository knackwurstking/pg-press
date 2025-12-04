package main

import (
	"fmt"
	"os"

	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/services"

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
				return withDBOperation(customDBPath, func(r *services.Registry) error {
					var err error
					if *useApiKey {
						err = r.Cookies.RemoveApiKey(*value)
					} else {
						err = r.Cookies.Remove(*value)
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
			telegramIDArg := cli.Int64(cmd, "user", cli.WithShort("u"), cli.Optional)

			return func(cmd *cli.Command) error {
				return withDBOperation(customDBPath, func(r *services.Registry) error {
					telegramID := models.TelegramID(*telegramIDArg)

					// Clean up cookies for a specific telegram user
					if telegramID != 0 {
						u, dberr := r.Users.Get(telegramID)
						if dberr != nil {
							handleNotFoundError(dberr)
							handleGenericError(dberr, fmt.Sprintf("Get user \"%d\" failed", telegramID))
						}

						cookies, err := r.Cookies.ListApiKey(u.ApiKey)
						if err != nil {
							handleGenericError(err, fmt.Sprintf("List cookies for user \"%d\" failed", telegramID))
						}

						for _, cookie := range cookies {
							if cookie.IsExpired() {
								if err = r.Cookies.Remove(cookie.Value); err != nil {
									// Print out error and continue
									fmt.Fprintf(os.Stderr, "Removing cookie with value \"%s\": %v\n", cookie.Value, err)
								}
							}
						}

						return nil
					}

					// Clean up all cookies
					cookies, err := r.Cookies.List()
					if err != nil {
						handleGenericError(err, "List cookies from database failed")
					}

					for _, cookie := range cookies {
						if cookie.IsExpired() {
							if err = r.Cookies.Remove(cookie.Value); err != nil {
								// Print out error and continue
								fmt.Fprintf(os.Stderr, "Removing cookie with value \"%s\": %v\n", cookie.Value, err)
							}
						}
					}

					return nil
				})
			}
		}),
	}
}
