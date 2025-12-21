package main

import (
	"fmt"
	"os"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/shared"

	"github.com/SuperPaintman/nice/cli"
)

func listUserCommand() cli.Command {
	return cli.Command{
		Name:  "list",
		Usage: cli.Usage("List all users"),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := createDBPathOption(cmd)

			return func(cmd *cli.Command) error {
				return withDBOperation(*customDBPath, false, func() error {
					users, merr := db.ListUsers()
					if merr != nil {
						return merr
					}

					fmt.Printf("TELEGRAM ID\tUSER NAME\n")
					fmt.Printf("-----------\t---------\n")
					for _, u := range users {
						fmt.Printf("%d\t%s\n", u.ID, u.Name)
					}

					return nil
				})
			}
		}),
	}
}

func showUserCommand() cli.Command {
	return cli.Command{
		Name: "show",
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := createDBPathOption(cmd)
			flagApiKey := cli.Bool(cmd, "api-key", cli.Optional)
			telegramIDArg := cli.Int64Arg(cmd, "telegram-id", cli.Required)

			return func(cmd *cli.Command) error {
				return withDBOperation(*customDBPath, false, func() error {
					user, merr := db.GetUser(shared.TelegramID(*telegramIDArg))
					if merr != nil {
						return merr.Wrap("get user")
					}

					if *flagApiKey {
						fmt.Fprint(os.Stdout, user.ApiKey)
						return nil
					}

					fmt.Printf("Telegram ID\tUser Name\tApi Key\n")
					fmt.Printf("-----------\t---------\t-------\n")
					fmt.Printf("%d\t%s\t%s\n", user.ID, user.Name, user.ApiKey)

					cookies, merr := db.ListCookiesByUserID(user.ID)
					if merr != nil {
						return merr.Wrap("list cookies for user ID %d", user.ID)
					}

					if len(cookies) > 0 {
						fmt.Printf("\n=== Cookies ===n")
						fmt.Printf("LAST LOGIN\tUSER\tVALUE\tUSER AGENT>\n")

						for _, c := range cookies {
							fmt.Printf("%s\t%s\t%s\t%s\n",
								c.LastLogin.FormatDateTime(), c.UserID.String(),
								c.Value, c.UserAgent)
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
				return withDBOperation(*customDBPath, false, func() error {
					telegramID := shared.TelegramID(*telegramIDArg)

					merr := db.AddUser(&shared.User{
						ID:     telegramID,
						Name:   *userName,
						ApiKey: *apiKey,
					})
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
				return withDBOperation(*customDBPath, false, func() error {
					return db.DeleteUser(shared.TelegramID(*telegramIDArg))
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
				return withDBOperation(*customDBPath, false, func() error {
					user, err := db.GetUser(shared.TelegramID(*telegramID))
					if err != nil {
						return err
					}

					if *userName != "" {
						user.Name = *userName
					}

					if *apiKey != "" {
						user.ApiKey = *apiKey
					}

					return db.UpdateUser(user)
				})
			}
		}),
	}
}
