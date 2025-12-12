package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/services/common"
	"github.com/knackwurstking/pg-press/services/shared"

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
			reviveDeadToolCommand(),
			deleteToolCommand(),
			listCyclesCommand(),
			listRegenerationsCommand(),
		},
	}
}

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
				return withDBOperation(customDBPath, func(r *common.DB) error {
					// Get all tools from database
					tools, merr := r.Tool.Tool.List()
					if merr != nil {
						return errors.Wrap(merr, "retrieve tools")
					}

					// Filter tools by ID if range/list is specified
					if *idRange != "" {
						var err error
						tools, err = filterToolsByIDs(tools, *idRange)
						if err != nil {
							return errors.Wrap(err, "filter tools by IDs")
						}
					}

					if len(tools) == 0 {
						if *idRange != "" {
							fmt.Printf("No tools found matching ID criteria '%s'.\n", *idRange)
						} else {
							fmt.Println("No tools found in database.")
						}
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
						// TODO: Fetch actual press number for tool using the new helper function in helper
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

						fmt.Fprintf(w, "%d\t%sx%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
							tool.ID,
							tool.Width, tool.Height,
							tool.Code,
							tool.Type,
							tool.Slot.German(),
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
		}),
	}
}

func listDeadToolsCommand() cli.Command {
	return createSimpleCommand("list-dead", "List all dead tools from the database", func(r *common.DB) error {
		// Get all dead tools from database
		tools, err := r.Tool.Tool.ListDeadTools()
		if err != nil {
			return fmt.Errorf("retrieve dead tools: %v", err)
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
			customDBPath := createDBPathOption(cmd)
			toolIDArg := cli.Int64Arg(cmd, "tool-id", cli.Required)

			return func(cmd *cli.Command) error {
				return withDBOperation(customDBPath, func(r *common.DB) error {
					toolID := shared.EntityID(*toolIDArg)

					// Get tool first to check if it exists
					tool, err := r.Tool.Tool.GetByID(toolID)
					if err != nil {
						return fmt.Errorf("find tool with ID %d: %v", toolID, err)
					}

					if tool.IsDead {
						fmt.Printf("Tool %d (%sx%s %s) is already marked as dead.\n", tool.ID, tool.Width, tool.Height, tool.Code)
						return nil
					}

					// Mark tool as dead
					tool.IsDead = true
					merr := r.Tool.Tool.Update(tool)
					if merr != nil {
						return errors.Wrap(merr, "mark tool as dead")
					}

					fmt.Printf("Successfully marked tool %d (%sx%s %s) as dead.\n", tool.ID, tool.Width, tool.Height, tool.Code)
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
				return withDBOperation(customDBPath, func(r *common.DB) error {
					toolID := shared.EntityID(*toolIDArg)

					// Get tool first to check if it exists
					tool, err := r.Tool.Tool.GetByID(toolID)
					if err != nil {
						return fmt.Errorf("find tool with ID %d: %v", toolID, err)
					}

					if !tool.IsDead {
						fmt.Printf("Tool %d (%sx%s %s) is not dead and doesn't need to be revived.\n", tool.ID, tool.Width, tool.Height, tool.Code)
						return nil
					}

					// Revive tool (mark as alive)
					tool.IsDead = false
					err = r.Tool.Tool.Update(tool)
					if err != nil {
						return fmt.Errorf("revive tool: %v", err)
					}

					fmt.Printf("Successfully revived tool %d (%sx%s %s).\n", tool.ID, tool.Width, tool.Height, tool.Code)
					return nil
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
				return withDBOperation(customDBPath, func(r *common.DB) error {
					toolID := shared.EntityID(*toolIDArg)

					// Get tool first to check if it exists and show info
					tool, err := r.Tool.Tool.GetByID(toolID)
					if err != nil {
						return fmt.Errorf("find tool with ID %d: %v", toolID, err)
					}

					fmt.Printf("Deleting tool %d (%sx%s %s) and all related data...\n", tool.ID, tool.Width, tool.Height, tool.Code)

					// 1. Delete all regenerations for this tool first (they reference cycles)
					regenerations, err := r.Tool.Regeneration.GetRegenerationHistory(toolID)
					if err != nil {
						return fmt.Errorf("get regenerations for tool: %v", err)
					}

					if len(regenerations) > 0 {
						fmt.Printf("Deleting %d regeneration(s)...\n", len(regenerations))
						for _, regen := range regenerations {
							if err := r.Tool.Regeneration.Delete(regen.ID); err != nil {
								return fmt.Errorf("delete regeneration %d: %v", regen.ID, err)
							}
						}
					}

					// 2. Delete all cycles for this tool
					cycles, err := r.Press.Cycle.ListPressCyclesForTool(toolID)
					if err != nil {
						return fmt.Errorf("get cycles for tool: %v", err)
					}

					if len(cycles) > 0 {
						fmt.Printf("Deleting %d cycle(s)...\n", len(cycles))
						for _, cycle := range cycles {
							if err := r.Press.Cycle.Delete(cycle.ID); err != nil {
								return fmt.Errorf("delete cycle %d: %v", cycle.ID, err)
							}
						}
					}

					// 3. Finally, delete the tool itself
					err = r.Tool.Tool.Delete(toolID)
					if err != nil {
						return fmt.Errorf("delete tool: %v", err)
					}

					fmt.Printf("Successfully deleted tool %d (%sx%s %s) with %d cycle(s) and %d regeneration(s).\n",
						tool.ID, tool.Width, tool.Height, tool.Code, len(cycles), len(regenerations))
					return nil
				})
			}
		}),
	}
}

func listCyclesCommand() cli.Command {
	return cli.Command{
		Name:  "list-cycles",
		Usage: cli.Usage("List press cycles for a tool by ID"),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := createDBPathOption(cmd)
			toolIDArg := cli.Int64Arg(cmd, "tool-id", cli.Required)

			return func(cmd *cli.Command) error {
				return withDBOperation(customDBPath, func(r *common.DB) error {
					toolID := shared.EntityID(*toolIDArg)

					// Get tool first to check if it exists and show info
					tool, err := r.Tool.Tool.GetByID(toolID)
					if err != nil {
						return fmt.Errorf("find tool with ID %d: %v", toolID, err)
					}

					fmt.Printf("Tool Information: ID %d (%sx%s %s) - %s - %s\n\n",
						tool.ID, tool.Width, tool.Height, tool.Code, tool.Type, tool.Slot.German())

					// Get cycles for this tool
					cycles, err := r.Press.Cycle.ListPressCyclesForTool(toolID)
					if err != nil {
						return fmt.Errorf("retrieve cycles: %v", err)
					}

					// Display Cycles
					fmt.Printf("=== PRESS CYCLES ===\n")
					if len(cycles) == 0 {
						fmt.Println("No cycles found for this tool")
					} else {
						w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
						fmt.Fprintln(w, "ID\tPRESS\tPOSITION\tTOTAL CYCLES\tDATE\tPERFORMED BY")
						fmt.Fprintln(w, "----\t-----\t--------\t------------\t----\t------------")

						for _, cycle := range cycles {
							fmt.Fprintf(w, "%d\t%d\t%s\t%d\t%s\t%d\n",
								cycle.ID,
								cycle.PressNumber,
								cycle.ToolPosition.GermanString(),
								cycle.TotalCycles,
								cycle.Date.Format(env.DateTimeFormat),
								cycle.PerformedBy,
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

// filterToolsByIDs filters tools based on ID range or comma-separated list
func filterToolsByIDs(tools []*shared.Tool, idSpec string) ([]*shared.Tool, error) {
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
	var filteredTools []*shared.Tool
	for _, tool := range tools {
		if idSet[tool.ID] {
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

func listRegenerationsCommand() cli.Command {
	return cli.Command{
		Name:  "list-regenerations",
		Usage: cli.Usage("List regenerations for a tool by ID"),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := createDBPathOption(cmd)
			toolIDArg := cli.Int64Arg(cmd, "tool-id", cli.Required)

			return func(cmd *cli.Command) error {
				return withDBOperation(customDBPath, func(r *common.DB) error {
					toolID := shared.EntityID(*toolIDArg)

					// Get tool first to check if it exists and show info
					tool, err := r.Tool.Tool.GetByID(toolID)
					if err != nil {
						return fmt.Errorf("find tool with ID %d: %v", toolID, err)
					}

					fmt.Printf("Tool Information: ID %d (%sx%s %s) - %s - %s\n\n",
						tool.ID, tool.Width, tool.Height, tool.Code, tool.Type, tool.Slot.German())

					// Get regenerations for this tool
					regenerations, err := r.Tool.Regeneration.GetRegenerationHistory(toolID)
					if err != nil {
						return fmt.Errorf("retrieve regenerations: %v", err)
					}

					// Display Regenerations
					fmt.Printf("=== REGENERATIONS ===\n")
					if len(regenerations) == 0 {
						fmt.Println("No regenerations found for this tool")
					} else {
						w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
						fmt.Fprintln(w, "ID\tCYCLE ID\tREASON\tPERFORMED BY")
						fmt.Fprintln(w, "----\t--------\t------\t------------")

						for _, regen := range regenerations {
							performedByStr := "None"
							if regen.PerformedBy != nil {
								performedByStr = strconv.Itoa(int(*regen.PerformedBy))
							}

							fmt.Fprintf(w, "%d\t%d\t%s\t%s\n",
								regen.ID,
								regen.CycleID,
								regen.Reason,
								performedByStr,
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
