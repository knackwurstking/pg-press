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

	if err := createAttachments(dbPath); err != nil {
		panic("failed to create attachments: " + err.Error())
	}

	if err := createMetalSheets(dbPath); err != nil {
		panic("failed to create metal sheets: " + err.Error())
	}

	if err := createNotes(dbPath); err != nil {
		panic("failed to create notes: " + err.Error())
	}

	// TODO: Create: "press_cycles"
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

	// Write `notes` array to (JSON) "./json/notes.json"
	jsonPath, err := filepath.Abs(PathToJSON)
	if err != nil {
		return err
	}
	os.Mkdir(jsonPath, 0700) // Ignore error if folder exists

	jsonFilePath := filepath.Join(jsonPath, "notes.json")
	jsonFile, err := os.Create(jsonFilePath)
	if err != nil {
		return fmt.Errorf("create json file: %v", err)
	}
	defer jsonFile.Close()

	encoder := json.NewEncoder(jsonFile)
	encoder.SetIndent("", "\t")
	if err = encoder.Encode(notes); err != nil {
		return fmt.Errorf("encode notes to json: %v", err)
	}

	return nil
}
