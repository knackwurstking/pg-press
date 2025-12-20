package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/shared"

	"github.com/SuperPaintman/nice/cli"
)

func removeCookiesCommand() cli.Command {
	return cli.Command{
		Name: "remove",
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := createDBPathOption(cmd)

			useApiKey := cli.Bool(cmd, "api-key",
				cli.Usage("Remove all entries containing the api-key"),
				cli.Optional)

			value := cli.StringArg(cmd, "value",
				cli.Usage("Remove entry containing the cookie value, only if `--api-key` is not set"),
				cli.Required)

			return func(cmd *cli.Command) error {
				return withDBOperation(*customDBPath, false, func() error {
					if *useApiKey {
						user, merr := db.GetUserByApiKey(*value)
						if merr != nil {
							fmt.Fprintf(os.Stderr, "Failed to get user by api key: %v\n", merr)
							if merr.Code == http.StatusNotFound {
								os.Exit(exitCodeNotFound)
							}
							os.Exit(exitCodeGeneric)
						}

						merr = db.DeleteCookiesByUserID(user.ID)
						if merr != nil {
							fmt.Fprintf(os.Stderr, "Failed to remove cookies for user: %v\n", merr)
							if merr.Code == http.StatusNotFound {
								os.Exit(exitCodeNotFound)
							}
							os.Exit(exitCodeGeneric)
						}
					}

					merr := db.DeleteCookie(*value)
					if merr != nil {
						fmt.Fprintf(os.Stderr, "Failed to remove cookie: %v\n", merr)
						if merr.Code == http.StatusNotFound {
							os.Exit(exitCodeNotFound)
						}
						os.Exit(exitCodeGeneric)
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
			argCustomDBPath := createDBPathOption(cmd)
			argTelegramIDArg := cli.Int64(cmd, "user", cli.WithShort("u"), cli.Optional)

			return func(cmd *cli.Command) error {
				return withDBOperation(*argCustomDBPath, false, func() error {
					telegramID := shared.TelegramID(*argTelegramIDArg)

					// Clean up cookies for a specific telegram user
					if telegramID != 0 {
						cleanUpCookiesForUser(telegramID)
						os.Exit(0)
					}

					// Clean up all cookies
					cookies, merr := db.ListCookies()
					if merr != nil {
						fmt.Fprintf(os.Stderr, "List cookies from database failed: %v\n", merr)
						os.Exit(exitCodeGeneric)
					}

					for _, c := range cookies {
						if c.IsExpired() {
							if merr = db.DeleteCookie(c.Value); merr != nil {
								// Print out error and continue
								fmt.Fprintf(os.Stderr, "Removing cookie with value \"%s\": %v\n", c.Value, merr)
								os.Exit(exitCodeGeneric)
							}
						}
					}

					return nil
				})
			}
		}),
	}
}

func cleanUpCookiesForUser(telegramID shared.TelegramID) {
	cookies, merr := db.ListCookiesByUserID(telegramID)
	if merr != nil {
		fmt.Fprintf(os.Stderr, "Failed to get the cookies: %v\n", merr)
		os.Exit(exitCodeGeneric)
	}

	for _, c := range cookies {
		if c.UserID != telegramID {
			continue
		}

		if c.IsExpired() {
			if merr = db.DeleteCookie(c.Value); merr != nil {
				// Print out error and continue
				fmt.Fprintf(os.Stderr, "Removing cookie for user %d with value \"%s\": %v\n",
					telegramID, c.Value, merr)
				os.Exit(exitCodeGeneric)
			}
		}
	}
}
