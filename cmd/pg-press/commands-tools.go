package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"

	"github.com/SuperPaintman/nice/cli"
)

func toolsCommand() cli.Command {
	return cli.Command{
		Name:  "tools",
		Usage: cli.Usage("Handle tools database table, list all tools"),
		Commands: []cli.Command{
			listToolsCommand(),
			deleteToolCommand(),

			markDeadCommand(),
			reviveDeadToolCommand(),

			listCyclesCommand(),

			listRegenerationsCommand(),
			deleteRegenerationCommand(),
		},
	}
}

// -----------------------------------------------------------------------------
// Tool Commands
// -----------------------------------------------------------------------------

func listToolsCommand() cli.Command {
	return cli.Command{
		Name:  "list",
		Usage: cli.Usage("List tools from the database with optional ID filtering"),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := createDBPathOption(cmd)
			idRange := cli.StringArg(cmd, "id-range",
				cli.Usage("ID range (e.g., '5..8' for range or '5,7,9' for specific IDs)"),
				cli.Optional)

			return func(cmd *cli.Command) error {
				return withDBOperation(*customDBPath, false, func() error {
					tools, merr := db.ListTools()
					if merr != nil {
						return errors.Wrap(merr, "list tools")
					}

					// Filter tools by ID if range/list is specified
					if *idRange != "" {
						var err error

						tools, err = filterToolsByIDs[*shared.Tool](tools, *idRange)
						if err != nil {
							return errors.Wrap(err, "filter tools by IDs")
						}
					}

					// Create tabwriter for nice formatting
					w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

					fmt.Printf("=== TOOLS ===\n\n")
					fmt.Printf("ID\tFORMAT\tPOSITION\tTYPE\tCODE\tCYCLES OFFSET\tCYCLES\tCASSETTE\tIS DEAD\n")
					fmt.Printf("--\t------\t--------\t----\t----\t-------------\t------\t--------\t-------\n")

					for _, t := range tools {
						if t.IsCassette() {
							continue
						}
						fmt.Printf("%d\t%dx%d\t%d\t%s\t%s\t%d\t%d\t%d\t%t\n",
							t.ID,
							t.Width, t.Height,
							t.Position,
							t.Type,
							t.Code,
							t.CyclesOffset,
							t.Cycles,
							t.Cassette,
							t.IsDead,
						)
					}

					fmt.Printf("\n=== CASSETTES ===\n\n")
					fmt.Printf("ID\tFORMAT\tPOSITION\tTYPE\tCODE\tCYCLES OFFSET\tCYCLES\tMIN THICKNESS\tMAX THICKNESS\tIS DEAD\n")
					fmt.Printf("--\t------\t--------\t----\t----\t-------------\t------\t-------------\t-------------\t-------\n")

					for _, t := range tools {
						if !t.IsCassette() {
							continue
						}
						fmt.Printf("%d\t%dx%d\t%d\t%s\t%s\t%d\t%d\t%d\t%d\t%t\n",
							t.ID,
							t.Width, t.Height,
							t.Position,
							t.Type,
							t.Code,
							t.CyclesOffset,
							t.Cycles,
							t.MinThickness,
							t.MaxThickness,
							t.IsDead,
						)
					}

					// Flush the tabwriter
					return w.Flush()
				})
			}
		}),
	}
}

func deleteToolCommand() cli.Command {
	return cli.Command{
		Name:  "delete",
		Usage: cli.Usage("Delete a tool by ID"),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := createDBPathOption(cmd)
			toolIDArg := cli.Int64Arg(cmd, "tool-id", cli.Required)

			return func(cmd *cli.Command) error {
				return withDBOperation(*customDBPath, false, func() error {
					merr := db.DeleteTool(shared.EntityID(*toolIDArg))
					if merr != nil {
						return errors.Wrap(merr, "delete tool")
					}
					return nil
				})
			}
		}),
	}
}

func markDeadCommand() cli.Command {
	return cli.Command{
		Name:  "mark-dead",
		Usage: cli.Usage("Mark a tool as dead by ID"),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := createDBPathOption(cmd)
			toolIDArg := cli.Int64Arg(cmd, "tool-id", cli.Required)

			return func(cmd *cli.Command) error {
				return withDBOperation(*customDBPath, false, func() error {
					merr := db.MarkToolAsDead(shared.EntityID(*toolIDArg))
					if merr != nil {
						return errors.Wrap(merr, "mark tool as dead")
					}
					return nil
				})
			}
		}),
	}
}

func reviveDeadToolCommand() cli.Command {
	return cli.Command{
		Name:  "revive",
		Usage: cli.Usage("Revive a dead tool by ID"),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := createDBPathOption(cmd)
			toolIDArg := cli.Int64Arg(cmd, "tool-id", cli.Required)

			return func(cmd *cli.Command) error {
				return withDBOperation(*customDBPath, false, func() error { // TODO: ...
					toolID := shared.EntityID(*toolIDArg)

					// Get tool first to check if it exists
					tool, err := r.Tool.Tools.GetByID(toolID)
					if err != nil {
						return fmt.Errorf("find tool with ID %d: %v", toolID, err)
					}

					if !tool.IsDead {
						fmt.Printf("Tool %d (%dx%d %s) is not dead and doesn't need to be revived.\n", tool.ID, tool.Width, tool.Height, tool.Code)
						return nil
					}

					// Revive tool (mark as alive)
					tool.IsDead = false
					err = r.Tool.Tools.Update(tool)
					if err != nil {
						return fmt.Errorf("revive tool: %v", err)
					}

					fmt.Printf("Successfully revived tool %d (%dx%d %s).\n", tool.ID, tool.Width, tool.Height, tool.Code)
					return nil
				})
			}
		}),
	}
}

// -----------------------------------------------------------------------------
// Tool Press Cycles Commands
// -----------------------------------------------------------------------------

// FIXME: tools/cassettes share the same ID space, so marking a cassette as dead should also be possible
func listCyclesCommand() cli.Command {
	return cli.Command{
		Name:  "list-cycles",
		Usage: cli.Usage("List press cycles for a tool by ID"),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := createDBPathOption(cmd)
			toolIDArg := cli.Int64Arg(cmd, "tool-id", cli.Required)

			return func(cmd *cli.Command) error {
				return withDBOperation(*customDBPath, func(r *common.DB) error {
					toolID := shared.EntityID(*toolIDArg)

					// Get tool first to check if it exists and show info
					tool, merr := r.Tool.Tools.GetByID(toolID)
					if merr != nil {
						return fmt.Errorf("find tool with ID %d: %v", toolID, merr)
					}

					fmt.Printf("Tool Information: ID %d (%dx%d %s) - %s\n\n",
						tool.ID, tool.Width, tool.Height, tool.Code, tool.Type) // NOTE: " - %s" removed, tool position

					// Get cycles for this tool
					cycles, merr := helper.ListCyclesForTool(r, toolID)
					if merr != nil {
						return fmt.Errorf("retrieve cycles: %v", merr)
					}

					// Display Cycles
					fmt.Printf("=== PRESS CYCLES ===\n")
					if len(cycles) == 0 {
						fmt.Println("No cycles found for this tool")
					} else {
						w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
						fmt.Fprintln(w, "ID\tTOOL ID\tPRESS\tPRESS CYCLES\tPARTIAL CYCLES\tSTART\tSTOP")
						fmt.Fprintln(w, "--\t-------\t-----\t------------\t--------------\t-----\t----")

						for _, cycle := range cycles {
							fmt.Fprintf(w, "%d\t%d\t%d\t%d\t%d\t%s\n%s",
								cycle.ID,
								cycle.ToolID,
								cycle.PressNumber,
								cycle.PressCycles,
								cycle.PartialCycles,
								cycle.Start.FormatDate(),
								cycle.Stop.FormatDate(),
							)
						}
						w.Flush()
						fmt.Printf("\nTotal cycles: %d\n", len(cycles))
					}

					return nil
				})
			}
		}),
	}
}

// -----------------------------------------------------------------------------
// Tool Regenerations Commands
// ---------------------------------------------------------------------------

func listRegenerationsCommand() cli.Command {
	return cli.Command{
		Name:  "list-regenerations",
		Usage: cli.Usage("List regenerations for a tool by ID"),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := createDBPathOption(cmd)
			toolIDArg := cli.Int64Arg(cmd, "tool-id", cli.Required)

			return func(cmd *cli.Command) error {
				return withDBOperation(*customDBPath, func(db *common.DB) error {
					toolID := shared.EntityID(*toolIDArg)

					// Get tool first to check if it exists and show info
					tool, merr := helper.GetToolByID(db, toolID)
					if merr != nil {
						return fmt.Errorf("find tool with ID %d: %v", toolID, merr)
					}

					baseTool := tool.GetBase()
					fmt.Printf("Tool Information: ID %d (%dx%d %s) - %s - %s\n\n",
						baseTool.ID, baseTool.Width, baseTool.Height, baseTool.Code, baseTool.Type, baseTool.Position.German())

					// Get regenerations for this tool
					regenerations, err := helper.GetRegenerationsForTool(db, toolID)
					if err != nil {
						return fmt.Errorf("retrieve regenerations: %v", err)
					}

					// Display Regenerations
					fmt.Printf("=== REGENERATIONS ===\n")
					if len(regenerations) == 0 {
						fmt.Println("No regenerations found for this tool")
					} else {
						w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
						fmt.Fprintln(w, "ID\tTOOL ID\tSTART\tSTOP\tCYCLES")
						fmt.Fprintln(w, "----\t-------\t-----\t----\t------")

						for _, regen := range regenerations {
							fmt.Fprintf(w, "%d\t%d\t%s\t%s\t%d\n",
								regen.ID,
								regen.ToolID,
								regen.Start.FormatDate(),
								regen.Stop.FormatDate(),
								regen.Cycles,
							)
						}
						w.Flush()
						fmt.Printf("\nTotal regenerations: %d\n", len(regenerations))
					}

					return nil
				})
			}
		}),
	}
}

func deleteRegenerationCommand() cli.Command {
	return cli.Command{
		Name:  "delete-regeneration",
		Usage: cli.Usage("Delete a tool regeneration by ID"),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := createDBPathOption(cmd)
			regenerationIDArg := cli.Int64Arg(cmd, "regeneration-id", cli.Required)

			return func(cmd *cli.Command) error {
				return withDBOperation(*customDBPath, func(r *common.DB) error {
					regenerationID := shared.EntityID(*regenerationIDArg)

					// Get regeneration first to check if it exists and show info
					regeneration, err := r.Tool.Regenerations.GetByID(regenerationID)
					if err != nil {
						return fmt.Errorf("find regeneration with ID %d: %v", regenerationID, err)
					}

					fmt.Printf("Deleting regeneration %d for tool %d...\n", regenerationID, regeneration.ToolID)

					// Delete the regeneration
					merr := r.Tool.Regenerations.Delete(regenerationID)
					if merr != nil {
						return fmt.Errorf("delete regeneration: %v", merr)
					}

					fmt.Printf("Successfully deleted regeneration %d.\n", regenerationID)
					return nil
				})
			}
		}),
	}
}

// -----------------------------------------------------------------------------
// Helper Functions
// -----------------------------------------------------------------------------

// filterToolsByIDs filters tools based on ID range or comma-separated list
func filterToolsByIDs[T shared.ModelTool](tools []T, idSpec string) ([]T, error) {
	var targetIDs []shared.EntityID
	var err error

	// Check if it's a range (contains "..")
	if strings.Contains(idSpec, "..") {
		targetIDs, err = parseIDRange(idSpec)
		if err != nil {
			return nil, err
		}
	} else {
		// Parse as comma-separated list
		targetIDs, err = parseIDList(idSpec)
		if err != nil {
			return nil, err
		}
	}

	// Create a set for efficient lookup
	idSet := make(map[shared.EntityID]bool)
	for _, id := range targetIDs {
		idSet[id] = true
	}

	// Filter tools
	var filteredTools []shared.ModelTool
	for _, tool := range tools {
		if idSet[tool.GetID()] {
			filteredTools = append(filteredTools, tool)
		}
	}

	return filteredTools, nil
}

// parseIDRange parses range like "5..8" into slice of IDs
func parseIDRange(rangeSpec string) ([]shared.EntityID, error) {
	parts := strings.Split(rangeSpec, "..")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid range format '%s', expected format: 'start..end'", rangeSpec)
	}

	start, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid start ID '%s': %v", parts[0], err)
	}

	end, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid end ID '%s': %v", parts[1], err)
	}

	if start > end {
		return nil, fmt.Errorf("start ID %d cannot be greater than end ID %d", start, end)
	}

	var ids []shared.EntityID
	for i := start; i <= end; i++ {
		ids = append(ids, shared.EntityID(i))
	}

	return ids, nil
}

// parseIDList parses comma-separated list like "5,7,9" into slice of IDs
func parseIDList(listSpec string) ([]shared.EntityID, error) {
	parts := strings.Split(listSpec, ",")
	var ids []shared.EntityID

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}

		id, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid ID '%s': %v", trimmed, err)
		}
		ids = append(ids, shared.EntityID(id))
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("no valid IDs found in list '%s'", listSpec)
	}

	return ids, nil
}
