// A minimal Go program to convert the SQLite database from an older version
// to JSON Format. This is just for private use and will be remove later.
package main

import (
	"flag"
	"fmt"
)

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: convert_to_json <path_to_sqlite_db>")
		return
	}

	dbPath := args[0]

	// First, log the command line arguments
	fmt.Printf("dbPath=%#v\n", dbPath)
}
