package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/shared"
	m "github.com/knackwurstking/pg-press/scripts/models"

	_ "github.com/mattn/go-sqlite3"
)

var (
	PathToImages, _ = filepath.Abs("./images")
	PathToJSON, _   = filepath.Abs("./json")
	PathToSQL, _    = filepath.Abs("./sql")

	FlagClean bool
)

func main() {
	{ // Flags
		flag.BoolVar(&FlagClean, "clean", false,
			"clean existing SQL database files before conversion")

		flag.Parse()
	}

	sqlPath := PathToSQL
	if FlagClean {
		os.RemoveAll(sqlPath)
	}

	if err := db.Open(sqlPath, true); err != nil {
		panic("failed to open database: " + err.Error())
	}
	defer db.Close()

	images := []string{}
	{ // Load Attachments (images)
		imagesDir, err := os.ReadDir(PathToImages)
		if err != nil {
			panic("failed to read images directory: " + err.Error())
		}

		for _, img := range imagesDir {
			images = append(images, img.Name())
		}
	}

	oldCycles := []m.Cycle{}
	if err := readJSON("cycles.json", &oldCycles); err != nil {
		panic("failed to read cycles: " + err.Error())
	}

	oldTools := []m.Tool{}
	if err := readJSON("tools.json", &oldTools); err != nil {
		panic("failed to read tools: " + err.Error())
	}

	if err := CreateToolData(oldCycles, oldTools); err != nil {
		panic("failed to convert tool data: " + err.Error())
	}

	if err := CreatePressData(oldCycles, oldTools); err != nil {
		panic("failed to convert press data: " + err.Error())
	}

	if err := CreateNoteData(); err != nil {
		panic("failed to convert note data: " + err.Error())
	}

	if err := CreateUserData(); err != nil {
		panic("failed to convert user data: " + err.Error())
	}

	if err := CreateReportsData(images); err != nil {
		panic("failed to convert reports data: " + err.Error())
	}
}

func CreateToolData(oldCycles []m.Cycle, oldTools []m.Tool) error {
	{ // Metal Sheets
		oldMetalSheets := []m.MetalSheet{}
		if err := readJSON("metal-sheets.json", &oldMetalSheets); err != nil {
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
		if err := readJSON("tool-regenerations.json", &oldToolRegenerations); err != nil {
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
			if t.Binding != nil {
				binding = shared.EntityID(*t.Binding)
			}

			if t.Type == "" {
				t.Type = "?"
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
				MinThickness: 1, // MinThickness does not exists in old data
				MaxThickness: 2, // MaxThickness does not exists in old data
			}

			if err := db.AddTool(tool); err != nil {
				fmt.Printf("Failed to add tool: %#v\n", tool)
				return err
			}
		}
	}

	return nil
}

func CreatePressData(cycles []m.Cycle, tools []m.Tool) error {
	{ // Create presses from cycles and tools (`Press` & `PressNumber`)
		pressNumbers := []shared.PressNumber{}
		for _, c := range cycles {
			if slices.Contains(pressNumbers, shared.PressNumber(c.PressNumber)) {
				continue
			}
			pressNumbers = append(pressNumbers, shared.PressNumber(c.PressNumber))
		}
		for _, t := range tools {
			if t.Press == nil {
				continue
			}
			if slices.Contains(pressNumbers, shared.PressNumber(*t.Press)) {
				continue
			}
			pressNumbers = append(pressNumbers, shared.PressNumber(*t.Press))
		}

		// Sort press numbers from low to high
		slices.Sort(pressNumbers)
		for _, p := range pressNumbers {
			// Find tools for slots "up" and "down". Set machine type
			// (0 & 5 == SACMI, 2,3,4 == SITI).
			slotUp := shared.EntityID(0)
			slotDown := shared.EntityID(0)
			for _, t := range tools {
				if t.Press != nil && shared.PressNumber(*t.Press) == p {
					switch t.Position {
					case m.PositionTop:
						slotUp = shared.EntityID(t.ID)
					case m.PositionBottom:
						slotDown = shared.EntityID(t.ID)
					}
				}
			}

			pressType := shared.MachineType("")
			switch p {
			case 0, 5:
				pressType = shared.MachineTypeSACMI
			case 2, 3, 4:
				pressType = shared.MachineTypeSITI
			default:
				return fmt.Errorf("unknown press number: %d", p)
			}

			press := &shared.Press{
				Number:       p,
				Type:         pressType,
				Code:         "",
				SlotUp:       slotUp,
				SlotDown:     slotDown,
				CyclesOffset: 0,
			}
			if err := db.AddPress(press); err != nil {
				return err
			}
		}
	}

	{ // Cycles
		presses, herr := db.ListPress()
		if herr != nil {
			return herr
		}

		for _, c := range cycles {
			var pressID shared.EntityID
			for _, p := range presses {
				if p.Number == shared.PressNumber(c.PressNumber) {
					pressID = p.ID
				}
			}

			if pressID == 0 {
				return fmt.Errorf("press not found for cycle: %#v", c)
			}

			cycle := &shared.Cycle{
				ID:          shared.EntityID(c.ID),
				ToolID:      shared.EntityID(c.ToolID),
				PressID:     pressID,
				PressCycles: c.TotalCycles,
				Stop:        shared.NewUnixMilli(c.Date),
			}
			if err := db.AddCycle(cycle); err != nil {
				return err
			}
		}
	}

	return nil
}

func CreateNoteData() error {
	oldNotes := []m.Note{}
	if err := readJSON("notes.json", &oldNotes); err != nil {
		panic("failed to read notes: " + err.Error())
	}

	for _, n := range oldNotes {
		level := shared.LevelNormal
		switch n.Level {
		case m.LevelInfo:
			level = shared.LevelInfo
		case m.LevelAttention:
			level = shared.LevelAttention
		case m.LevelBroken:
			level = shared.LevelBroken
		default:
			return fmt.Errorf("unknown note level: %d", n.Level)
		}

		note := &shared.Note{
			ID:        shared.EntityID(n.ID),
			Level:     level,
			Content:   n.Content,
			CreatedAt: shared.NewUnixMilli(n.CreatedAt),
			Linked:    n.Linked,
		}
		if err := db.AddNote(note); err != nil {
			return err
		}
	}

	return nil
}

func CreateUserData() error {
	oldUsers := []m.User{}
	if err := readJSON("users.json", &oldUsers); err != nil {
		panic("failed to read users: " + err.Error())
	}

	for _, u := range oldUsers {
		user := &shared.User{
			ID:     shared.TelegramID(u.TelegramID),
			Name:   u.Name,
			ApiKey: u.ApiKey,
		}
		if err := db.AddUser(user); err != nil {
			return err
		}
	}

	return nil
}

// CreateReportsData migrates trouble reports from old JSON data to the new SQL database.
func CreateReportsData(images []string) error {
	troubleReports := []m.TroubleReport{}
	if err := readJSON("trouble-reports.json", &troubleReports); err != nil {
		return err
	}

	for _, tr := range troubleReports {
		report := &shared.TroubleReport{
			ID:                shared.EntityID(tr.ID),
			Title:             tr.Title,
			Content:           tr.Content,
			LinkedAttachments: tr.NewLinkedAttachments,
			UseMarkdown:       tr.UseMarkdown,
		}
		if err := db.AddTroubleReport(report); err != nil {
			return err
		}
	}

	return nil
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
