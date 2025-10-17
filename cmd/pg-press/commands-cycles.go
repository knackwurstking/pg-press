package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/services"
	"github.com/knackwurstking/pgpress/pkg/models"

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
			customDBPath := createDBPathOption(cmd, "")
			pressNumber := cli.Int64Arg(cmd, "press-number", cli.Required)

			return func(cmd *cli.Command) error {
				return withDBOperation(customDBPath, func(registry *services.Registry) error {
					// Validate press number
					press := models.PressNumber(*pressNumber)
					if !models.IsValidPressNumber(&press) {
						return fmt.Errorf("invalid press number: %d (must be 0-5)", *pressNumber)
					}

					// Get all cycles and filter by press
					allCycles, err := registry.PressCycles.List()
					if err != nil {
						return fmt.Errorf("failed to retrieve cycles: %v", err)
					}

					// Filter cycles by press number
					var cycles []*models.Cycle
					for _, cycle := range allCycles {
						if cycle.PressNumber == press {
							cycles = append(cycles, cycle)
						}
					}

					if len(cycles) == 0 {
						fmt.Printf("No cycles found for press %d.\n", *pressNumber)
						return nil
					}

					// Create tabwriter for nice formatting
					w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

					// Print header
					fmt.Fprintln(w, "ID\tPRESS\tTOOL ID\tPOSITION\tTOTAL CYCLES\tDATE\tPERFORMED BY")
					fmt.Fprintln(w, "----\t-----\t-------\t--------\t------------\t----\t------------")

					// Print each cycle
					for _, cycle := range cycles {
						fmt.Fprintf(w, "%d\t%d\t%d\t%s\t%d\t%s\t%d\n",
							cycle.ID,
							cycle.PressNumber,
							cycle.ToolID,
							cycle.ToolPosition.GermanString(),
							cycle.TotalCycles,
							cycle.Date.Format(constants.DateTimeFormat),
							cycle.PerformedBy,
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
			customDBPath := createDBPathOption(cmd, "")
			cycleID := cli.Int64Arg(cmd, "cycle-id", cli.Required)

			return func(cmd *cli.Command) error {
				return withDBOperation(customDBPath, func(registry *services.Registry) error {
					// First check if there are any regenerations that reference this cycle
					hasRegenerations, err := registry.ToolRegenerations.HasRegenerationsForCycle(*cycleID)
					if err != nil {
						return fmt.Errorf("failed to check for regenerations: %v", err)
					}

					if hasRegenerations {
						return fmt.Errorf("cannot delete cycle %d: there are regenerations that reference this cycle. Delete the regenerations first", *cycleID)
					}

					fmt.Printf("Deleting cycle %d...\n", *cycleID)

					// Delete cycle
					err = registry.PressCycles.Delete(*cycleID)
					if err != nil {
						return fmt.Errorf("failed to delete cycle: %v", err)
					}

					fmt.Printf("Successfully deleted cycle %d.\n", *cycleID)
					return nil
				})
			}
		}),
	}
}
