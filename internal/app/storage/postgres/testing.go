package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"testing"

	_ "github.com/lib/pq"
)

// TestDB returns connection to the test database and clear function
func TestDB(t *testing.T, source string) (*sql.DB, func(...string)) {
	db, err := sql.Open("postgres", source)

	if err != nil {
		t.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	return db, func(tables ...string) {
		for _, t := range tables {
			q := fmt.Sprintf("truncate \"%s\" cascade;", t)
			_, err := db.Exec(q)
			if err != nil {
				log.Fatalf("Failed to clean the test database: %v", err)
			}
		}
		_ = db.Close()
	}
}
