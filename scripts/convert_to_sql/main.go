package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/shared"
	m "github.com/knackwurstking/pg-press/scripts/models"
)

const (
	PathToImages = "./images"
	PathToJSON   = "./json"
)

func main() {
	// TODO: convert "./images" and "./json" to new SQL database format "..."
	// 	- "sql/tool.sqlite"	    : metal_sheets, tool_regenerations, tools
	//  - "sql/press.sqlite"    : cycles, presses, press_regenerations
	//  - "sql/note.sqlite"	    : notes
	//  - "sql/user.sqlite"	    : cookies, users
	//  - "sql/reports.sqlite"	: trouble_reports

	if err := db.Open("sql", true); err != nil {
		panic("failed to open database: " + err.Error())
	}
	defer db.Close()

	oldCycles := []m.Cycle{}
	if err := readJSON("cycles.json", oldCycles); err != nil {
		panic("failed to read cycles: " + err.Error())
	}

	oldTools := []m.Tool{}
	if err := readJSON("tools.json", oldTools); err != nil {
		panic("failed to read tools: " + err.Error())
	}

	if err := createToolData(oldCycles, oldTools); err != nil {
		panic("failed to convert tool data: " + err.Error())
	}

	if err := createPressData(oldCycles, oldTools); err != nil {
		panic("failed to convert press data: " + err.Error())
	}

	if err := createNoteData(); err != nil {
		panic("failed to convert note data: " + err.Error())
	}

	if err := createUserData(); err != nil {
		panic("failed to convert user data: " + err.Error())
	}

	if err := createReportsData(); err != nil {
		panic("failed to convert reports data: " + err.Error())
	}
}

func createToolData(oldCycles []m.Cycle, oldTools []m.Tool) error {
	{ // Metal Sheets
		oldMetalSheets := []m.MetalSheet{}
		if err := readJSON("metal-sheets.json", oldMetalSheets); err != nil {
			return err
		}

		for _, ms := range oldMetalSheets {
			base := shared.BaseMetalSheet{
				ID:         shared.EntityID(ms.ID),
				ToolID:     shared.EntityID(ms.ToolID),
				TileHeight: ms.TileHeight,
				Value:      ms.Value,
			}
			isUpper := ms.MarkeHeight > 0
			if isUpper {
				lms := &shared.LowerMetalSheet{
					BaseMetalSheet: base,
					MarkeHeight:    ms.MarkeHeight,
					STF:            ms.STF,
					STFMax:         ms.STFMax,
					Identifier:     shared.MachineType(ms.Identifier),
				}
				if err := db.AddLowerMetalSheet(lms); err != nil {
					return err
				}
			} else {
				ums := &shared.UpperMetalSheet{
					BaseMetalSheet: base,
				}
				if err := db.AddUpperMetalSheet(ums); err != nil {
					return err
				}
			}
		}
	}

	{ // Tool Regenerations
		oldToolRegenerations := []m.ToolRegeneration{}
		if err := readJSON("tool-regenerations.json", oldToolRegenerations); err != nil {
			return err
		}

		for _, r := range oldToolRegenerations {
			var start, stop shared.UnixMilli
			// Find cycle id inside the JSON data
			for _, c := range oldCycles {
				if c.ID == r.CycleID {
					stop = shared.NewUnixMilli(c.Date)
					start = stop // NOTE: Start time is not available in old data
					break
				}
			}
			err := db.AddToolRegeneration(&shared.ToolRegeneration{
				ID:     shared.EntityID(r.ID),
				ToolID: shared.EntityID(r.ToolID),
				Start:  start,
				Stop:   stop,
			})
			if err != nil {
				return err
			}
		}
	}

	{ // Tools
		for _, t := range oldTools {
			position := shared.SlotUnknown
			switch t.Position {
			case m.PositionTop:
				position = shared.SlotUpper
			case m.PositionTopCassette:
				position = shared.SlotUpperCassette
			case m.PositionBottom:
				position = shared.SlotLower
			}
			if position == shared.SlotUnknown {
				return fmt.Errorf("unknown tool position: %s", t.Position)
			}

			binding := shared.EntityID(0)
			if t.Binding == nil {
				binding = shared.EntityID(*t.Binding)
			}

			tool := &shared.Tool{
				ID:           shared.EntityID(t.ID),
				Width:        t.Format.Width,
				Height:       t.Format.Height,
				Position:     position,
				Type:         t.Type,
				Code:         t.Code,
				CyclesOffset: 0, // CyclesOffset does not exists in old data
				IsDead:       t.IsDead,
				Cassette:     binding,
				MinThickness: 0, // MinThickness does not exists in old data
				MaxThickness: 0, // MaxThickness does not exists in old data
			}

			if err := db.AddTool(tool); err != nil {
				return err
			}
		}
	}

	return nil
}

func createPressData(cycles []m.Cycle, tools []m.Tool) error {
	// TODO: Create presses from cycles and tools (`Press` & `PressNumber`)

	return errors.New("not implemented")
}

func createNoteData() error {
	return errors.New("not implemented")
}

func createUserData() error {
	return errors.New("not implemented")
}

func createReportsData() error {
	return errors.New("not implemented")
}

func readJSON(fileName string, t any) error {
	path := filepath.Join(PathToJSON, fileName)
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", path, err)
	}
	defer f.Close()

	d := json.NewDecoder(f)
	if err = d.Decode(t); err != nil {
		return fmt.Errorf("failed to decode %s: %w", path, err)
	}

	return nil
}
