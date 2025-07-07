package main

import (
	"fmt"
	"os"

	"github.com/SuperPaintman/nice/cli"
	"github.com/williepotgieter/keymaker"
)

func apiKeyCommand() cli.Command {
	return cli.Command{
		Name: "api-key",
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			return func(cmd *cli.Command) error {
				apiKey, err := keymaker.NewApiKey("pgp", 32)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Generating a new api key failed: %s\n", err.Error())
				}

				fmt.Println(apiKey)

				return nil
			}
		}),
	}
}
