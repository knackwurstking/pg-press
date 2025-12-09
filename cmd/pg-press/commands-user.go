package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/services"
	"github.com/labstack/gommon/color"

	"github.com/SuperPaintman/nice/cli"
)

func listUserCommand() cli.Command {
	return createSimpleCommand("list", "List all users", func(r *services.Registry) error {
		users, err := r.Users.List()
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
			customDBPath := createDBPathOption(cmd)
			flagApiKey := cli.Bool(cmd, "api-key", cli.Optional)
			telegramIDArg := cli.Int64Arg(cmd, "telegram-id", cli.Required)

			return func(cmd *cli.Command) error {
				return withDBOperation(customDBPath, func(r *services.Registry) error {
					telegramID := models.TelegramID(*telegramIDArg)

					user, merr := r.Users.Get(telegramID)
					if merr != nil {
						fmt.Fprintf(os.Stderr, "Failed to get user (%d): %v\n", telegramID, merr)

						if merr != nil && merr.Code == http.StatusNotFound {
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
					fmt.Printf("%-15d %-20s %s\n", user.TelegramID, user.Name, user.ApiKey)

					if cookies, err := r.Cookies.ListApiKey(user.ApiKey); err != nil {
						fmt.Fprintf(os.Stderr, "Get cookies from the database: %v\n", err)
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
			customDBPath := createDBPathOption(cmd)
			telegramIDArg := cli.Int64Arg(cmd, "telegram-id", cli.Required)
			userName := cli.StringArg(cmd, "user-name", cli.Required)
			apiKey := cli.StringArg(cmd, "api-key", cli.Required)

			return func(cmd *cli.Command) error {
				return withDBOperation(customDBPath, func(r *services.Registry) error {
					telegramID := models.TelegramID(*telegramIDArg)

					user := models.NewUser(telegramID, *userName, *apiKey)
					_, merr := r.Users.Add(user)
					if merr.IsExistsError() {
						return fmt.Errorf("user already exists: %d (%s)", telegramID, *userName)
					}
					return merr
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
				return withDBOperation(customDBPath, func(r *services.Registry) error {
					return r.Users.Delete(models.TelegramID(*telegramIDArg))
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
				return withDBOperation(customDBPath, func(r *services.Registry) error {
					user, err := r.Users.Get(models.TelegramID(*telegramID))
					if err != nil {
						return err
					}

					if *userName != "" {
						user.Name = *userName
					}

					if *apiKey != "" {
						user.ApiKey = *apiKey
					}

					return r.Users.Update(user)
				})
			}
		}),
	}
}
