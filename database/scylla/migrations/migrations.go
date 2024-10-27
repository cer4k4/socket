package migrations

import (
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
		f, err := os.Open("database/scylla/migrations")
		if err != nil {
			return nil, err
		}
		list, err := f.Readdir(-1)
		f.Close()
		if err != nil {
			return nil, err
		}
		slices.SortFunc(list, func(a, b os.FileInfo) int {
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

		content, err := os.ReadFile("database/scylla/migrations/" + file.Name())
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
