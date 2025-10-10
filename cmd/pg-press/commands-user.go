package main

import (
	"fmt"
	"os"

	"github.com/knackwurstking/pgpress/internal/services"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
	"github.com/labstack/gommon/color"

	"github.com/SuperPaintman/nice/cli"
)

func listUserCommand() cli.Command {
	return createSimpleCommand("list", "List all users", func(db *services.Registry) error {
		users, err := db.Users.List()
		if err != nil {
			return err
		}

		fmt.Printf("%-15s %s\n", "Telegram ID", "User Name")
		fmt.Printf("%-15s %s\n", "-----------", "---------")
		for _, u := range users {
			fmt.Printf("%-15d %s\n", u.TelegramID, u.Name)
		}

		return nil
	})
}

func showUserCommand() cli.Command {
	return cli.Command{
		Name: "show",
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := createDBPathOption(cmd, "")
			flagApiKey := cli.Bool(cmd, "api-key", cli.Optional)
			telegramID := cli.Int64Arg(cmd, "telegram-id", cli.Required)

			return func(cmd *cli.Command) error {
				return withDBOperation(customDBPath, func(db *services.Registry) error {
					user, err := db.Users.Get(*telegramID)
					if err != nil {
						handleNotFoundError(err)
						handleGenericError(err, fmt.Sprintf("Get user \"%d\" failed", *telegramID))
						return err
					}

					if *flagApiKey {
						fmt.Fprint(os.Stdout, user.ApiKey)
						return nil
					}

					fmt.Printf("%-15s %-20s %s\n", "Telegram ID", "User Name", "Api Key")
					fmt.Printf("%-15s %-20s %s\n", "-----------", "---------", "-------")
					fmt.Printf("%-15d %-20s %s\n", user.TelegramID, user.Name, user.ApiKey)

					if cookies, err := db.Cookies.ListApiKey(user.ApiKey); err != nil {
						fmt.Fprintf(os.Stderr, "Failed to get cookies from the database: %s\n", err.Error())
					} else {
						if len(cookies) > 0 {
							fmt.Printf("\n%s <last-login> - <api-key> - <value> - <user-agent>\n\n", color.Underline(color.Bold("Cookies:")))

							for _, c := range models.SortCookies(cookies) {
								fmt.Printf("%s - %s - %s - \"%s\"\n", c.TimeString(), color.Bold(c.ApiKey), c.Value, color.Italic(c.UserAgent))
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
			customDBPath := createDBPathOption(cmd, "")
			telegramID := cli.Int64Arg(cmd, "telegram-id", cli.Required)
			userName := cli.StringArg(cmd, "user-name", cli.Required)
			apiKey := cli.StringArg(cmd, "api-key", cli.Required)

			return func(cmd *cli.Command) error {
				return withDBOperation(customDBPath, func(db *services.Registry) error {
					user := models.NewUser(*telegramID, *userName, *apiKey)
					if _, err := db.Users.Add(user); utils.IsAlreadyExistsError(err) {
						return fmt.Errorf("user already exists: %d (%s)", *telegramID, *userName)
					} else {
						return err
					}
				})
			}
		}),
	}
}

func removeUserCommand() cli.Command {
	return cli.Command{
		Name: "remove",
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := createDBPathOption(cmd, "")
			telegramID := cli.Int64Arg(cmd, "telegram-id", cli.Required)

			return func(cmd *cli.Command) error {
				return withDBOperation(customDBPath, func(db *services.Registry) error {
					return db.Users.Delete(*telegramID)
				})
			}
		}),
	}
}

func modUserCommand() cli.Command {
	return cli.Command{
		Name: "mod",
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := createDBPathOption(cmd, "")
			userName := cli.String(cmd, "name", cli.WithShort("n"), cli.Optional)
			apiKey := cli.String(cmd, "api-key", cli.Optional)
			telegramID := cli.Int64Arg(cmd, "telegram-id", cli.Required)

			return func(cmd *cli.Command) error {
				return withDBOperation(customDBPath, func(db *services.Registry) error {
					user, err := db.Users.Get(*telegramID)
					if err != nil {
						return err
					}

					if *userName != "" {
						user.Name = *userName
					}

					if *apiKey != "" {
						user.ApiKey = *apiKey
					}

					return db.Users.Update(user)
				})
			}
		}),
	}
}
