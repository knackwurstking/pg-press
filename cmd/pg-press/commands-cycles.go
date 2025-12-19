package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/knackwurstking/pg-press/internal/common"
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
				return withDBOperation(*argCustomDBPath, func(db *common.DB) error {
					// Validate press number
					pressNumber := shared.PressNumber(*argPressNumber)
					if !pressNumber.IsValid() {
						return fmt.Errorf("invalid press number: %d (must be 0-5)", pressNumber)
					}

					// Get all cycles and filter by press
					allCycles, err := db.Press.Cycle.List()
					if err != nil {
						return fmt.Errorf("retrieve cycles: %v", err)
					}

					// Filter cycles by press number
					var cycles []*shared.Cycle
					for _, cycle := range allCycles {
						if cycle.PressNumber == pressNumber {
							cycles = append(cycles, cycle)
						}
					}

					if len(cycles) == 0 {
						fmt.Fprintf(os.Stderr, "No cycles found for press %d.\n", pressNumber)
						return nil
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
				return withDBOperation(*customDBPath, func(db *common.DB) error {
					// Delete cycle
					err := db.Press.Cycle.Delete(shared.EntityID(*cycleIDArg))
					if err != nil {
						return fmt.Errorf("delete cycle: %v", err)
					}
					return nil
				})
			}
		}),
	}
}
