package main

import (
	"fmt"
	"os"

	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
	"github.com/labstack/gommon/color"

	"github.com/SuperPaintman/nice/cli"
)

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

				fmt.Printf("%-15s %s\n", "Telegram ID", "User Name")
				fmt.Printf("%-15s %s\n", "-----------", "---------")
				for _, u := range users {
					fmt.Printf("%-15d %s\n", u.TelegramID, u.Name)
				}

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
					if utils.IsNotFoundError(err) {
						os.Exit(exitCodeNotFound)
					}

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

			telegramID := cli.Int64Arg(cmd, "telegram-id", cli.Required)
			userName := cli.StringArg(cmd, "user-name", cli.Required)
			apiKey := cli.StringArg(cmd, "api-key", cli.Required)

			return func(cmd *cli.Command) error {
				db, err := openDB(*customDBPath)
				if err != nil {
					return err
				}

				user := models.NewUser(*telegramID, *userName, *apiKey)
				if _, err = db.Users.Add(user, nil); utils.IsAlreadyExistsError(err) {
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

				db.Users.Delete(*telegramID, nil)

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
					user.Name = *userName
				}

				if *apiKey != "" {
					user.ApiKey = *apiKey
				}

				err = db.Users.Update(user, nil)
				return err
			}
		}),
	}
}
