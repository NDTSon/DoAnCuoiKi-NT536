package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	// Locate the database file
	// Assuming running from project root, db is usually at data/dev.db
	dbPath := "data/dev.db"
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		// Try absolute path if needed, or check current dir
		cwd, _ := os.Getwd()
		fmt.Printf("Database file not found at %s. Current dir: %s\n", dbPath, cwd)
		return
	}

	fmt.Printf("Opening database at: %s\n", dbPath)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Query recordings
	rows, err := db.Query("SELECT id, room_name, status, video_path, created_at FROM recordings")
	if err != nil {
		fmt.Printf("Error querying recordings: %v\n", err)
		// Try checking if table exists
		_, err = db.Query("SELECT 1 FROM recordings")
		if err != nil {
			fmt.Println("Table 'recordings' likely does not exist!")
		}
		return
	}
	defer rows.Close()

	count := 0
	fmt.Println("--- Recordings in DB ---")
	for rows.Next() {
		var id, room, status, path, created string
		if err := rows.Scan(&id, &room, &status, &path, &created); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("[%d] ID: %s | Room: %s | Status: %s | Path: %s | Date: %s\n", count+1, id, room, status, path, created)
		count++
	}

	if count == 0 {
		fmt.Println("No recordings found in the database.")
	} else {
		fmt.Printf("Total found: %d\n", count)
	}
}
