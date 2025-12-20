package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/shared"

	"github.com/SuperPaintman/nice/cli"
	"github.com/labstack/gommon/color"
)

func listUserCommand() cli.Command {
	return createSimpleCommand("list", "List all users", func() error {
		users, merr := db.ListUsers()
		if merr != nil {
			return merr
		}

		fmt.Printf("%-15s %s\n", "Telegram ID", "User Name")
		fmt.Printf("%-15s %s\n", "-----------", "---------")
		for _, u := range users {
			fmt.Printf("%-15d %s\n", u.ID, u.Name)
		}

		return nil
	})
}

func showUserCommand() cli.Command {
	return cli.Command{
		Name: "show",
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := createDBPathOption(cmd)
			flagApiKey := cli.Bool(cmd, "api-key", cli.Optional)
			telegramIDArg := cli.Int64Arg(cmd, "telegram-id", cli.Required)

			return func(cmd *cli.Command) error {
				return withDBOperation(*customDBPath, func(r *common.DB) error {
					telegramID := shared.TelegramID(*telegramIDArg)

					user, merr := r.User.Users.GetByID(telegramID)
					if merr != nil {
						fmt.Fprintf(os.Stderr, "Failed to get user (%d): %v\n", telegramID, merr)

						if merr.Code == http.StatusNotFound {
							os.Exit(exitCodeNotFound)
						}

						os.Exit(exitCodeGeneric)
					}

					if *flagApiKey {
						fmt.Fprint(os.Stdout, user.ApiKey)
						return nil
					}

					fmt.Printf("%-15s %-20s %s\n", "Telegram ID", "User Name", "Api Key")
					fmt.Printf("%-15s %-20s %s\n", "-----------", "---------", "-------")
					fmt.Printf("%-15d %-20s %s\n", user.ID, user.Name, user.ApiKey)

					if cookies, err := r.User.Cookies.List(); err != nil {
						fmt.Fprintf(os.Stderr, "Get cookies from the database: %v\n", err)
					} else {
						if len(cookies) > 0 {
							fmt.Printf("\n%s <last-login> - <user> - <value> - <user-agent>\n\n",
								color.Underline(color.Bold("Cookies:")))

							for _, c := range cookies {
								fmt.Printf("%s - %s - %s - \"%s\"\n",
									c.LastLogin.FormatDateTime(),
									color.Bold(c.UserID.String()),
									c.Value,
									color.Italic(c.UserAgent),
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

func addUserCommand() cli.Command {
	return cli.Command{
		Name: "add",
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := createDBPathOption(cmd)
			telegramIDArg := cli.Int64Arg(cmd, "telegram-id", cli.Required)
			userName := cli.StringArg(cmd, "user-name", cli.Required)
			apiKey := cli.StringArg(cmd, "api-key", cli.Required)

			return func(cmd *cli.Command) error {
				return withDBOperation(*customDBPath, func(r *common.DB) error {
					telegramID := shared.TelegramID(*telegramIDArg)

					user := &shared.User{
						ID:     telegramID,
						Name:   *userName,
						ApiKey: *apiKey,
					}
					merr := r.User.Users.Create(user)
					if merr != nil && merr.IsExistsError() {
						return fmt.Errorf("user already exists: %d (%s)", telegramID, *userName)
					}
					return nil
				})
			}
		}),
	}
}

func removeUserCommand() cli.Command {
	return cli.Command{
		Name: "remove",
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := createDBPathOption(cmd)
			telegramIDArg := cli.Int64Arg(cmd, "telegram-id", cli.Required)

			return func(cmd *cli.Command) error {
				return withDBOperation(*customDBPath, func(r *common.DB) error {
					return r.User.Users.Delete(shared.TelegramID(*telegramIDArg))
				})
			}
		}),
	}
}

func modUserCommand() cli.Command {
	return cli.Command{
		Name: "mod",
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := createDBPathOption(cmd)
			userName := cli.String(cmd, "name", cli.WithShort("n"), cli.Optional)
			apiKey := cli.String(cmd, "api-key", cli.Optional)
			telegramID := cli.Int64Arg(cmd, "telegram-id", cli.Required)

			return func(cmd *cli.Command) error {
				return withDBOperation(*customDBPath, func(r *common.DB) error {
					user, err := r.User.Users.GetByID(shared.TelegramID(*telegramID))
					if err != nil {
						return err
					}

					if *userName != "" {
						user.Name = *userName
					}

					if *apiKey != "" {
						user.ApiKey = *apiKey
					}

					return r.User.Users.Update(user)
				})
			}
		}),
	}
}
