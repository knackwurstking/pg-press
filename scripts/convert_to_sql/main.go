package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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
	databaseName := "tools.json"
	metalSheets := readJSON(databaseName, []m.MetalSheet{})
	toolRegenerations := readJSON(databaseName, []m.ToolRegeneration{})
	tools := readJSON(databaseName, []m.Tool{})

	// TODO: ...

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
