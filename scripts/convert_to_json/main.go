// A minimal Go program to convert the SQLite database from an older version
// to JSON Format. This is just for private use and will be remove later.
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/knackwurstking/pg-press/internal/utils"
)

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: convert_to_json <path_to_sqlite_db>")
		return
	}

	dbPath := args[0]

	if err := createAttachments(dbPath); err != nil {
		panic("failed to creaete attachments: " + err.Error())
	}

	// TODO: Convert "metal_sheets" table entries to JSON files []*MetalSheet
}

// createAttachments reads the attachments SQL table and creates the images
// folder at "./images/*"
func createAttachments(dbPath string) error {
	imagesPath, err := filepath.Abs("./images")
	if err != nil {
		return err
	}
	os.Mkdir(imagesPath, 0700) // Ignore error if folder exists

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	const query = `SELECT id, mime_type, data FROM attachments;`
	r, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("query: %v", err)
	}
	defer r.Close()

	attachments := []Attachment{}
	for r.Next() {
		a := Attachment{}
		if err = r.Scan(&a.ID, &a.MimeType, &a.Data); err != nil {
			return fmt.Errorf("scan attachment: %v", err)
		}
		attachments = append(attachments, a)
	}

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
