package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/shared"

	"github.com/SuperPaintman/nice/cli"
)

func cyclesCommand() cli.Command {
	return cli.Command{
		Name:  "cycles",
		Usage: cli.Usage("Handle press cycles database table, list and delete cycles"),
		Commands: []cli.Command{
			listCyclesAllCommand(),
			deleteCycleCommand(),
		},
	}
}

func listCyclesAllCommand() cli.Command {
	return cli.Command{
		Name:  "list",
		Usage: cli.Usage("List all cycles for a specific press number"),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			argCustomDBPath := createDBPathOption(cmd)
			argPressNumber := cli.Int64Arg(cmd, "press-number", cli.Required)

			return func(cmd *cli.Command) error {
				return withDBOperation(*argCustomDBPath, false, func() error {
					// Validate press number
					pressNumber := shared.PressNumber(*argPressNumber)
					if !pressNumber.IsValid() {
						return fmt.Errorf("invalid press number: %d (must be 0-5)", pressNumber)
					}

					// Get all cycles and filter by press
					cycles, err := db.ListCyclesByPressNumber(pressNumber)
					if err != nil {
						return fmt.Errorf("retrieve cycles: %v", err)
					}

					// Create tabwriter for nice formatting
					w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

					// Print header
					fmt.Fprintln(w, "ID\tTOOL ID\tPRESS\tTOTAL CYCLES\tPARTIAL CYCLES\tSTART\tSTOP")
					fmt.Fprintln(w, "--\t-------\t-----\t------------\t--------------\t-----\t----")

					// Print each cycle
					for _, cycle := range cycles {
						fmt.Fprintf(w, "%d\t%d\t%d\t%d\t%d\t%d\t%d\n",
							cycle.ID,
							cycle.ToolID,
							cycle.PressNumber,
							cycle.PressCycles,
							cycle.PartialCycles,
							cycle.Start,
							cycle.Stop,
						)
					}

					// Flush the tabwriter
					w.Flush()

					fmt.Printf("\nTotal cycles: %d\n", len(cycles))

					return nil
				})
			}
		}),
	}
}

func deleteCycleCommand() cli.Command {
	return cli.Command{
		Name:  "delete",
		Usage: cli.Usage("Delete a cycle by ID"),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := createDBPathOption(cmd)
			cycleIDArg := cli.Int64Arg(cmd, "cycle-id", cli.Required)

			return func(cmd *cli.Command) error {
				return withDBOperation(*customDBPath, false, func() error {
					// Delete cycle
					if merr := db.DeleteCycle(shared.EntityID(*cycleIDArg)); merr != nil {
						return fmt.Errorf("delete cycle: %v", merr)
					}
					return nil
				})
			}
		}),
	}
}
