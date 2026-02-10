package main

import (
	"os"
	"path/filepath"

	"github.com/knackwurstking/pg-press/internal/logger"

	"github.com/SuperPaintman/nice/cli"
	"github.com/knackwurstking/ui/templ/ui"
)

const (
	version = "v0.0.1"

	exitCodeGeneric     = 1
	exitCodeNotFound    = 10
	exitCodeServerStart = 20
)

var (
	appName    string
	configPath string
	log        *ui.Logger
)

func init() {
	log = logger.New("main")

	appName = filepath.Base(os.Args[0])

	p, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	configPath = filepath.Join(p, appName)
	if err := os.MkdirAll(configPath, 0700); err != nil {
		panic(err)
	}
}

func main() {
	a := cli.App{
		Name: appName,
		Usage: cli.Usage(`Exit Codes:
  1   Generic
  10  Not Found
`),
		Commands: []cli.Command{
			apiKeyCommand(),

			{
				Name: "user",
				Usage: cli.Usage(
					"Handle users database table, add, remove or modify user data"),
				Commands: []cli.Command{
					listUserCommand(),
					showUserCommand(),
					addUserCommand(),
					removeUserCommand(),
					modUserCommand(),
				},
			},

			{
				Name: "cookies",
				Usage: cli.Usage(
					"Handle cookies database table, remove cookies data"),
				Commands: []cli.Command{
					removeCookiesCommand(),
					autoCleanCookiesCommand(),
				},
			},

			toolsCommand(),

			serverCommand(),

			cli.CompletionCommand(),
		},
		CommandFlags: []cli.CommandFlag{
			cli.HelpCommandFlag(),
			cli.VersionCommandFlag(version),
		},
	}

	a.HandleError(a.Run())
}
