package main

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/knackwurstking/pgpress/internal/services"

	"github.com/SuperPaintman/nice/cli"
)

func toolsCommand() cli.Command {
	return cli.Command{
		Name:  "tools",
		Usage: cli.Usage("Handle tools database table, list all tools"),
		Commands: []cli.Command{
			listToolsCommand(),
		},
	}
}

func listToolsCommand() cli.Command {
	return createSimpleCommand("list", "List all tools from the database", func(registry *services.Registry) error {
		// Get all tools from database
		tools, err := registry.Tools.List()
		if err != nil {
			return fmt.Errorf("failed to retrieve tools: %v", err)
		}

		if len(tools) == 0 {
			fmt.Println("No tools found in database.")
			return nil
		}

		// Create tabwriter for nice formatting
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		// Print header
		fmt.Fprintln(w, "ID\tFORMAT\tCODE\tTYPE\tPOSITION\tPRESS\tREGEN")
		fmt.Fprintln(w, "----\t------\t----\t----\t--------\t-----\t-----")

		// Print each tool
		for _, tool := range tools {
			pressStr := "None"
			if tool.Press != nil {
				pressStr = strconv.Itoa(int(*tool.Press))
			}

			regenStr := "No"
			if tool.Regenerating {
				regenStr = "Yes"
			}

			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\t%s\n",
				tool.ID,
				tool.Format.String(),
				tool.Code,
				tool.Type,
				tool.Position.GermanString(),
				pressStr,
				regenStr,
			)
		}

		// Flush the tabwriter
		w.Flush()

		fmt.Printf("\nTotal tools: %d\n", len(tools))

		return nil
	})
}
