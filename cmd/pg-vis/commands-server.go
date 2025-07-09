package main

import (
	"net/http"
	"os"

	"github.com/SuperPaintman/nice/cli"
	"github.com/charmbracelet/log"
	"github.com/knackwurstking/pg-vis/html"
	"github.com/labstack/echo/v4"
)

func serverCommand() cli.Command {
	return cli.Command{
		Name:  "server",
		Usage: cli.Usage("Start the server."),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := cli.String(cmd, "db",
				cli.WithShort("d"),
				cli.Optional,
			)

			addr := cli.String(
				cmd, "addr",
				cli.WithShort("a"),
				cli.Usage("Set server address (<host>:<port>)"),
			)
			*addr = serverAddress

			return func(cmd *cli.Command) error {
				db, err := openDB(*customDBPath)
				if err != nil {
					return err
				}

				e := echo.New()

				// Init logger
				log.SetLevel(log.DebugLevel)
				log.SetReportCaller(true)

				e.Use(middlewareLogger())
				e.Use(middlewareKeyAuth(db))

				e.HTTPErrorHandler = func(err error, c echo.Context) {
					log.Debugf("HTTPErrorHandler -> err=%#v", err)

					if err == nil {
						return 
					}

					code := 500
					message := ""

					// NOTE: Maybe serve an error page here instead
					if err, ok := err.(*echo.HTTPError); ok {
						if err == nil {
							return
						}

						code = err.Code
						message = http.StatusText(err.Code)

						if m, ok := err.Message.(string); ok {
							message = m
						} else if e, ok := err.Message.(error); ok {
							message = e.Error()
						}
					} else {
						message = err.Error()
					}

					log.Warnf("HTTPErrorHandler -> %s", err.Error())
					c.JSON(code, message)
				}

				html.Serve(e, html.Options{
					ServerPathPrefix: serverPathPrefix,
					DB:               db,
				})

				if err := e.Start(*addr); err != nil {
					log.Errorf("Starting the server failed: %s", err.Error())
					os.Exit(exitCodeServerStart)
				}

				return nil
			}
		}),
	}
}
