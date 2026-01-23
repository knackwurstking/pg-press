// A minimal Go program to convert the SQLite database from an older version
// to JSON Format. This is just for private use and will be remove later.
package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/knackwurstking/pg-press/internal/utils"
)

const (
	PathToImages = "./images"
	PathToJSON   = "./json"
)

var (
	FlagClean bool
)

func main() {
	{
		flag.BoolVar(&FlagClean, "clean", false, "Clean the output folders before writing new data")

		flag.Parse()
	}

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: convert_to_json <path_to_sqlite_db>")
		return
	}
	dbPath := args[0]

	if FlagClean {
		for _, path := range []string{PathToImages, PathToJSON} {
			absPath, err := filepath.Abs(path)
			if err != nil {
				panic("failed to get absolute path of " + path + " folder: " + err.Error())
			}
			if err = os.RemoveAll(absPath); err != nil {
				panic("failed to clean " + path + " folder: " + err.Error())
			}
		}
	}

	var err error
	if err = createAttachments(dbPath); err != nil {
		panic("failed to create attachments: " + err.Error())
	}

	if err = createMetalSheets(dbPath); err != nil {
		panic("failed to create metal sheets: " + err.Error())
	}

	if err = createNotes(dbPath); err != nil {
		panic("failed to create notes: " + err.Error())
	}

	if err = createCycles(dbPath); err != nil {
		panic("failed to create press cycles: " + err.Error())
	}

	if err = createToolRegenerations(dbPath); err != nil {
		panic("failed to create tool regenerations: " + err.Error())
	}

	if err = createTools(dbPath); err != nil {
		panic("failed to create tools: " + err.Error())
	}

	if err = createTroubleReports(dbPath); err != nil {
		panic("failed to create trouble reports: " + err.Error())
	}

	if err = createUsers(dbPath); err != nil {
		panic("failed to create users: " + err.Error())
	}
}

// createAttachments reads the attachments SQL table and creates the images
// folder at "./images/*"
func createAttachments(dbPath string) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	attachments := []Attachment{}
	{
		const query = `SELECT id, mime_type, data FROM attachments;`
		r, err := db.Query(query)
		if err != nil {
			return fmt.Errorf("query: %v", err)
		}
		defer r.Close()

		for r.Next() {
			a := Attachment{}
			if err = r.Scan(&a.ID, &a.MimeType, &a.Data); err != nil {
				return fmt.Errorf("scan attachment: %v", err)
			}
			attachments = append(attachments, a)
		}
	}

	imagesPath, err := filepath.Abs(PathToImages)
	if err != nil {
		return err
	}

	os.Mkdir(imagesPath, 0700) // Ignore error if folder exists

	for _, a := range attachments {
		fileName := utils.GetAttachmentFileName(a.ID + a.GetExtension())
		path := filepath.Join(imagesPath, fileName)
		if err = os.WriteFile(path, a.Data, 0644); err != nil {
			return fmt.Errorf("write attachment file: %v", err)
		}
		fmt.Fprintf(os.Stderr, "Wrote attachment with ID of %s to %s\n", a.ID, path)
	}

	return nil
}

func createMetalSheets(dbPath string) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("open db: %v", err)
	}

	metalSheets := []MetalSheet{}
	{
		const query = `SELECT id, tile_height, value, marke_height, stf, stf_max, identifier, tool_id FROM metal_sheets;`
		r, err := db.Query(query)
		if err != nil {
			return fmt.Errorf("query: %v", err)
		}
		defer r.Close()

		for r.Next() {
			ms := MetalSheet{}
			err := r.Scan(
				&ms.ID,
				&ms.TileHeight,
				&ms.Value,
				&ms.MarkeHeight,
				&ms.STF,
				&ms.STFMax,
				&ms.Identifier,
				&ms.ToolID,
			)
			if err != nil {
				return fmt.Errorf("scan metal sheet: %v", err)
			}
			fmt.Fprintf(os.Stderr, "Read metal sheet with ID of %d\n", ms.ID)
			metalSheets = append(metalSheets, ms)
		}
	}

	// Write `metalSheets` array to (JSON) "./json/metal_sheets.json"
	jsonPath, err := filepath.Abs(PathToImages)
	if err != nil {
		return err
	}
	os.Mkdir(jsonPath, 0700) // Ignore error if folder exists

	jsonFilePath := filepath.Join(jsonPath, "metal_sheets.json")
	jsonFile, err := os.Create(jsonFilePath)
	if err != nil {
		return fmt.Errorf("create json file: %v", err)
	}
	defer jsonFile.Close()

	encoder := json.NewEncoder(jsonFile)
	encoder.SetIndent("", "\t")
	if err = encoder.Encode(metalSheets); err != nil {
		return fmt.Errorf("encode metal sheets to json: %v", err)
	}

	return nil
}

func createNotes(dbPath string) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("open db: %v", err)
	}

	notes := []Note{}
	{
		const query = `SELECT id, level, content, created_at, linked FROM notes;`
		r, err := db.Query(query)
		if err != nil {
			return fmt.Errorf("query: %v", err)
		}
		defer r.Close()

		for r.Next() {
			n := Note{}
			err := r.Scan(
				&n.ID,
				&n.Level,
				&n.Content,
				&n.CreatedAt,
				&n.Linked,
			)
			if err != nil {
				return fmt.Errorf("scan note: %v", err)
			}
			fmt.Fprintf(os.Stderr, "Read note with ID of %d\n", n.ID)
			notes = append(notes, n)
		}
	}

	return writeJSON(notes)
}

func createCycles(dbPath string) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("open db: %v", err)
	}

	cycles := []Cycle{}
	{
		const query = `SELECT id, press_number, tool_id, tool_position, total_cycles, date, performed_by FROM press_cycles;`
		r, err := db.Query(query)
		if err != nil {
			return fmt.Errorf("query: %v", err)
		}
		defer r.Close()

		for r.Next() {
			c := Cycle{}
			err := r.Scan(
				&c.ID,
				&c.PressNumber,
				&c.ToolID,
				&c.ToolPosition,
				&c.TotalCycles,
				&c.Date,
				&c.PerformedBy,
			)
			if err != nil {
				return fmt.Errorf("scan cycle: %v", err)
			}
			fmt.Fprintf(os.Stderr, "Read cycle with ID of %d\n", c.ID)
			cycles = append(cycles, c)
		}
	}

	return writeJSON(cycles)
}

func createToolRegenerations(dbPath string) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("open db: %v", err)
	}

	toolRegenerations := []ToolRegeneration{}
	{
		const query = `SELECT id, tool_id, cycle_id, reason, performed_by FROM tool_regenerations;`
		r, err := db.Query(query)
		if err != nil {
			return fmt.Errorf("query: %v", err)
		}
		defer r.Close()

		for r.Next() {
			tr := ToolRegeneration{}
			err := r.Scan(
				&tr.ID,
				&tr.ToolID,
				&tr.CycleID,
				&tr.Reason,
				&tr.PerformedBy,
			)
			if err != nil {
				return fmt.Errorf("scan tool regeneration: %v", err)
			}
			fmt.Fprintf(os.Stderr, "Read tool regeneration with ID of %d\n", tr.ID)
			toolRegenerations = append(toolRegenerations, tr)
		}
	}

	return writeJSON(toolRegenerations)
}

func writeJSON(data any) error {
	jsonPath, err := filepath.Abs(PathToJSON)
	if err != nil {
		return err
	}
	os.Mkdir(jsonPath, 0700) // Ignore error if folder exists

	jsonFilePath := filepath.Join(jsonPath, "tool_regenerations.json")
	jsonFile, err := os.Create(jsonFilePath)
	if err != nil {
		return fmt.Errorf("create json file: %v", err)
	}
	defer jsonFile.Close()

	encoder := json.NewEncoder(jsonFile)
	encoder.SetIndent("", "\t")
	if err = encoder.Encode(data); err != nil {
		return fmt.Errorf("encode tool regenerations to json: %v", err)
	}

	return nil
}

func createTools(dbPath string) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("open db: %v", err)
	}

	tools := []Tool{}
	{
		const query = `SELECT id, position, format, type, code, regenerating, is_dead, press, binding FROM tools;`
		r, err := db.Query(query)
		if err != nil {
			return fmt.Errorf("query: %v", err)
		}
		defer r.Close()

		for r.Next() {
			t := Tool{}
			var formatData []byte
			err := r.Scan(
				&t.ID,
				&t.Position,
				&formatData,
				&t.Type,
				&t.Code,
				&t.Regenerating,
				&t.IsDead,
				&t.Press,
				&t.Binding,
			)
			if err != nil {
				return fmt.Errorf("scan tool: %v", err)
			}
			if err = json.Unmarshal(formatData, &t.Format); err != nil {
				return fmt.Errorf("unmarshal tool format: %v", err)
			}
			fmt.Fprintf(os.Stderr, "Read tool with ID of %d\n", t.ID)
			tools = append(tools, t)
		}
	}

	return writeJSON(tools)
}

func createTroubleReports(dbPath string) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("open db: %v", err)
	}

	troubleReports := []TroubleReport{}
	{
		const query = `SELECT id, title, content, linked_attachments, use_markdown FROM trouble_reports;`
		r, err := db.Query(query)
		if err != nil {
			return fmt.Errorf("query: %v", err)
		}
		defer r.Close()

		for r.Next() {
			tr := TroubleReport{}
			var linkedAttachmentsData []byte
			err := r.Scan(
				&tr.ID,
				&tr.Title,
				&tr.Content,
				&linkedAttachmentsData,
				&tr.UseMarkdown,
			)
			if err != nil {
				return fmt.Errorf("scan trouble report: %v", err)
			}
			if err = json.Unmarshal(linkedAttachmentsData, &tr.LinkedAttachments); err != nil {
				return fmt.Errorf("unmarshal linked attachments: %v", err)
			}
			fmt.Fprintf(os.Stderr, "Read trouble report with ID of %d\n", tr.ID)
			troubleReports = append(troubleReports, tr)
		}
	}

	return writeJSON(troubleReports)
}

func createUsers(dbPath string) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("open db: %v", err)
	}

	users := []User{}
	{
		const query = `SELECT telegram_id, user_name, api_key, last_feed FROM users;`
		r, err := db.Query(query)
		if err != nil {
			return fmt.Errorf("query: %v", err)
		}
		defer r.Close()

		for r.Next() {
			u := User{}
			err := r.Scan(
				&u.TelegramID,
				&u.Name,
				&u.ApiKey,
				&u.LastFeed,
			)
			if err != nil {
				return fmt.Errorf("scan user: %v", err)
			}
			fmt.Fprintf(os.Stderr, "Read user with Telegram ID of %d\n", u.TelegramID)
			users = append(users, u)
		}
	}

	return writeJSON(users)
}
