package main

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/knackwurstking/pgpress/internal/services"
	"github.com/knackwurstking/pgpress/pkg/models"

	"github.com/SuperPaintman/nice/cli"
)

func toolsCommand() cli.Command {
	return cli.Command{
		Name:  "tools",
		Usage: cli.Usage("Handle tools database table, list all tools"),
		Commands: []cli.Command{
			listToolsCommand(),
			listDeadToolsCommand(),
			markDeadCommand(),
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
		fmt.Fprintln(w, "ID\tFORMAT\tCODE\tTYPE\tPOSITION\tPRESS\tREGEN\tSTATUS")
		fmt.Fprintln(w, "----\t------\t----\t----\t--------\t-----\t-----\t------")

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

			statusStr := "Alive"
			if tool.IsDead {
				statusStr = "Dead"
			}

			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				tool.ID,
				tool.Format.String(),
				tool.Code,
				tool.Type,
				tool.Position.GermanString(),
				pressStr,
				regenStr,
				statusStr,
			)
		}

		// Flush the tabwriter
		w.Flush()

		fmt.Printf("\nTotal tools: %d\n", len(tools))

		return nil
	})
}

func listDeadToolsCommand() cli.Command {
	return createSimpleCommand("list-dead", "List all dead tools from the database", func(registry *services.Registry) error {
		// Get all dead tools from database
		tools, err := registry.Tools.ListDeadTools()
		if err != nil {
			return fmt.Errorf("failed to retrieve dead tools: %v", err)
		}

		if len(tools) == 0 {
			fmt.Println("No dead tools found in database.")
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

		fmt.Printf("\nTotal dead tools: %d\n", len(tools))

		return nil
	})
}

func markDeadCommand() cli.Command {
	return cli.Command{
		Name:  "mark-dead",
		Usage: cli.Usage("Mark a tool as dead by ID"),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := createDBPathOption(cmd, "")
			toolID := cli.Int64Arg(cmd, "tool-id", cli.Required)

			return func(cmd *cli.Command) error {
				return withDBOperation(customDBPath, func(registry *services.Registry) error {
					// Get tool first to check if it exists
					tool, err := registry.Tools.Get(*toolID)
					if err != nil {
						return fmt.Errorf("failed to find tool with ID %d: %v", *toolID, err)
					}

					if tool.IsDead {
						fmt.Printf("Tool %d (%s %s) is already marked as dead.\n", tool.ID, tool.Format.String(), tool.Code)
						return nil
					}

					// Create a dummy user for CLI operations (you might want to make this configurable)
					user := &models.User{
						TelegramID: 1,
						Name:       "cli-user",
					}

					// Mark tool as dead
					err = registry.Tools.MarkAsDead(*toolID, user)
					if err != nil {
						return fmt.Errorf("failed to mark tool as dead: %v", err)
					}

					fmt.Printf("Successfully marked tool %d (%s %s) as dead.\n", tool.ID, tool.Format.String(), tool.Code)
					return nil
				})
			}
		}),
	}
}
