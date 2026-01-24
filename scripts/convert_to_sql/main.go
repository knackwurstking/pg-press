package main

import (
	"encoding/json"
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

	var err error

	if err = toolData(); err != nil {
		panic("failed to convert tool data: " + err.Error())
	}

	if err = pressData(); err != nil {
		panic("failed to convert press data: " + err.Error())
	}

	if err = noteData(); err != nil {
		panic("failed to convert note data: " + err.Error())
	}

	if err = userData(); err != nil {
		panic("failed to convert user data: " + err.Error())
	}

	if err = reportsData(); err != nil {
		panic("failed to convert reports data: " + err.Error())
	}
}

func toolData() error {
	if err := db.Open("sql", true); err != nil {
		return err
	}
	defer db.Close()

	{
		oldMetalSheets := []m.MetalSheet{}
		if err := readJSON("metal-sheets", oldMetalSheets); err != nil {
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

	{
		oldToolRegenerations := []m.ToolRegeneration{}
		err := readJSON("tool-regenerations", oldToolRegenerations)
		if err != nil {
			return err
		}

		newToolRegenerations := []*shared.ToolRegeneration{}
		for _, r := range oldToolRegenerations {
			// TODO: Create regeneration and add to database
		}
	}

	{
		oldTools := []m.Tool{}
		err = readJSON("tools", oldTools)
		if err != nil {
			return err
		}

		newTools := []*shared.Tool{}
		for _, t := range oldTools {
			// TODO: Convert old to new
		}

		// TODO: Write data to SQL database
	}

	return nil
}

func pressData() error {
	return readJSON("press.json")
}

func noteData() error {
	return readJSON("note.json")
}

func userData() error {
	return readJSON("user.json")
}

func reportsData() error {
	return readJSON("reports.json")
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
