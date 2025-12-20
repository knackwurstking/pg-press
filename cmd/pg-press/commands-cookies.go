package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/knackwurstking/pg-press/internal/shared"

	"github.com/SuperPaintman/nice/cli"
)

func removeCookiesCommand() cli.Command {
	return cli.Command{
		Name: "remove",
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := createDBPathOption(cmd)

			// TODO: Switch this to user_id instead, this will make things a lot easier
			useApiKey := cli.Bool(cmd, "api-key",
				cli.Usage("Remove all entries containing the api-key"),
				cli.Optional)

			value := cli.StringArg(cmd, "value",
				cli.Usage("Remove entry containing the cookie value, only if `--api-key` is not set"),
				cli.Required)

			return func(cmd *cli.Command) error {
				return withDBOperation(*customDBPath, func(r *common.DB) error {
					var err error
					if *useApiKey {
						// Get users, we need to find the user ID for the api key
						users, merr := r.User.Users.List()
						if merr != nil {
							fmt.Fprintf(os.Stderr, "Failed to list users: %v\n", merr)
							os.Exit(exitCodeGeneric)
						}

						// Get cookies, we need to find all cookies for the user
						// ID which matches the api key
						cookies, merr := r.User.Cookies.List()
						if merr != nil {
							fmt.Fprintf(os.Stderr, "Failed to list cookies: %v\n", merr)
							os.Exit(exitCodeGeneric)
						}

						for _, u := range users {
							// Only process users which match the api key
							if u.ApiKey != *value {
								continue
							}

							// Remove all cookies matching the current user ID
							// (user_id and api_key are unique)
							for _, c := range cookies {
								if c.UserID != u.ID {
									continue
								}

								merr := r.User.Cookies.Delete(c.Value)
								if merr != nil {
									// Ignore not found errors, continue with others
									if merr.Code == http.StatusNotFound {
										continue
									}
									fmt.Fprintf(os.Stderr, "Failed to remove cookie entry: %v\n", merr)
									os.Exit(exitCodeGeneric)
								}
							}
						}
					} else {
						err = r.User.Cookies.Delete(*value)
					}

					if err != nil {
						fmt.Fprintf(os.Stderr, "Failed to remove cookie entry: %v\n", err)
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
				return withDBOperation(*argCustomDBPath, func(db *common.DB) error {
					telegramID := shared.TelegramID(*argTelegramIDArg)

					// Clean up cookies for a specific telegram user
					if telegramID != 0 {
						cleanUpCookiesForUser(db, telegramID)
						os.Exit(0)
					}

					// Clean up all cookies
					cookies, merr := db.User.Cookies.List()
					if merr != nil {
						fmt.Fprintf(os.Stderr,
							"List cookies from database failed: %v\n",
							merr,
						)
						os.Exit(exitCodeGeneric)
					}

					for _, c := range cookies {
						if c.IsExpired() {
							merr = db.User.Cookies.Delete(c.Value)
							if merr != nil {
								// Print out error and continue
								fmt.Fprintf(os.Stderr,
									"Removing cookie with value \"%s\": %v\n",
									c.Value, merr,
								)
							}
						}
					}

					return nil
				})
			}
		}),
	}
}

func cleanUpCookiesForUser(db *common.DB, telegramID shared.TelegramID) {
	cookies, merr := db.User.Cookies.List()
	if merr != nil {
		fmt.Fprintf(os.Stderr, "Failed to get the cookies: %v\n", merr)
		os.Exit(exitCodeGeneric)
	}

	user, merr := db.User.Users.GetByID(telegramID)
	if merr != nil {
		fmt.Fprintf(os.Stderr, "Failed to get user (%d): %v\n", telegramID, merr)

		if merr.Code == http.StatusNotFound {
			os.Exit(exitCodeNotFound)
		}

		os.Exit(exitCodeGeneric)
	}

	for _, c := range cookies {
		if c.UserID != user.ID {
			continue
		}

		if c.IsExpired() {
			merr = db.User.Cookies.Delete(c.Value)
			if merr != nil {
				// Print out error and continue
				fmt.Fprintf(os.Stderr,
					"Removing cookie for user %d with value \"%s\": %v\n",
					telegramID, c.Value, merr,
				)
			}
		}
	}
}
