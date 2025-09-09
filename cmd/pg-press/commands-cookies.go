package main

import (
	"errors"
	"os"
	"time"

	"github.com/SuperPaintman/nice/cli"
	"github.com/knackwurstking/pgpress/internal/database/dberror"
	cookiemodels "github.com/knackwurstking/pgpress/internal/database/models/cookie"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/constants"
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
					logger.AppLogger.Error("Removing cookies from database failed: %s", err.Error())
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

				t := time.Now().Add(0 - constants.CookieExpirationDuration).UnixMilli()
				isExpired := func(cookie *cookiemodels.Cookie) bool {
					return t >= cookie.LastLogin
				}

				// Clean up cookies for a specific telegram user
				if *telegramID != 0 {
					u, err := db.Users.Get(*telegramID)
					if err != nil {
						if errors.Is(err, dberror.ErrNotFound) {
							os.Exit(exitCodeNotFound)
						}

						logger.AppLogger.Error("Get user \"%d\" failed: %s", *telegramID, err.Error())
						os.Exit(exitCodeGeneric)
					}

					cookies, err := db.Cookies.ListApiKey(u.ApiKey)
					if err != nil {
						logger.AppLogger.Error("List cookies for user \"%d\" failed: %s", *telegramID, err.Error())
						os.Exit(exitCodeGeneric)
					}

					for _, cookie := range cookies {
						if isExpired(cookie) {
							if err = db.Cookies.Remove(cookie.Value); err != nil {
								// Print out error and continue
								logger.AppLogger.Error(
									"Removing cookie with value \"%s\" failed: %s",
									cookie.Value, err.Error())
							}
						}
					}

					return nil
				}

				// Clean up all cookies
				cookies, err := db.Cookies.List()
				if err != nil {
					logger.AppLogger.Error("List cookies from database failed: %s", err.Error())
					os.Exit(exitCodeGeneric)
				}

				for _, cookie := range cookies {
					if isExpired(cookie) {
						if err = db.Cookies.Remove(cookie.Value); err != nil {
							// Print out error and continue
							logger.AppLogger.Error(
								"Removing cookie with value \"%s\" failed: %s",
								cookie.Value, err.Error())
						}
					}
				}

				return nil
			}
		}),
	}
}
