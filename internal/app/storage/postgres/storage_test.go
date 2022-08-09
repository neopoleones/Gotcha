package postgres_test

import (
	"os"
	"testing"
)

var (
	databaseConnectionString string
)

func TestMain(m *testing.M) {
	databaseConnectionString = os.Getenv("DB_CON")
	if databaseConnectionString == "" {
		databaseConnectionString = "postgres://postgres:test@server/gotcha-test?sslmode=disable"
	}
	os.Exit(m.Run())
}
