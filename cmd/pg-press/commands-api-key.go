package main

import (
	"fmt"

	"github.com/SuperPaintman/nice/cli"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/williepotgieter/keymaker"
)

func apiKeyCommand() cli.Command {
	return cli.Command{
		Name:  "api-key",
		Usage: cli.Usage("Generating a new Api Key"),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			return func(cmd *cli.Command) error {
				apiKey, err := keymaker.NewApiKey("pgp", 32)
				if err != nil {
					return errors.Wrap(err, "generate api key")
				}

				fmt.Print(apiKey) // Yes, no newline at the end
				return nil
			}
		}),
	}
}
