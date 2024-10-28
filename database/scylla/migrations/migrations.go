package migrations

import (
	"fmt"
	"log"
	"sort"

	"os"

	"io/fs"
	"slices"
	"strings"

	"github.com/gocql/gocql"
)

func RunMigrations(session *gocql.Session) {
	// Create schema_migrations table if it doesn't exist
	err := session.Query(`
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version text PRIMARY KEY
        )
    `).Exec()
	if err != nil {
		log.Fatalf("Error creating schema_migrations table: %v", err)
	}

	// Get list of applied migrations
	appliedMigrations := make(map[string]bool)
	iter := session.Query("SELECT version FROM schema_migrations").Iter()
	var version string
	for iter.Scan(&version) {
		appliedMigrations[version] = true
	}
	if err := iter.Close(); err != nil {
		log.Fatalf("Error reading applied migrations: %v", err)
	}

	// Execute new migrations
	files, err := func() ([]fs.FileInfo, error) {
		// Use the directory path, not the file path
		dirPath := "/home/aka/Templates/simple_socket/database/scylla/migrations"

		f, err := os.Open(dirPath)
		if err != nil {
			return nil, fmt.Errorf("error opening directory: %v", err)
		}
		defer f.Close()

		list, err := f.Readdir(-1)
		if err != nil {
			return nil, fmt.Errorf("error reading directory: %v", err)
		}

		slices.SortFunc(list, func(a, b fs.FileInfo) int {
			return strings.Compare(a.Name(), b.Name())
		})

		return list, nil
	}()
	if err != nil {
		log.Fatalf("Error reading migrations directory: %v", err)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	for _, file := range files {
		if appliedMigrations[file.Name()] {
			continue // Migration already applied
		}

		content, err := os.ReadFile("/home/aka/Templates/simple_socket/database/scylla/migrations/001_create_messages_table.cql")
		if err != nil {
			log.Fatalf("Error reading migration file %s: %v", file.Name(), err)
		}

		query := string(content)
		if err := session.Query(query).Exec(); err != nil {
			log.Fatalf("Failed to execute migration %s: %v", file.Name(), err)
		}

		// Record applied migration
		err = session.Query("INSERT INTO schema_migrations (version) VALUES (?)", file.Name()).Exec()
		if err != nil {
			log.Fatalf("Failed to record migration %s: %v", file.Name(), err)
		}

		log.Printf("Migration %s applied successfully.", file.Name())
	}
}
