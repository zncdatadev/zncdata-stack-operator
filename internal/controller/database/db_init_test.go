package controller

import (
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"github.com/stretchr/testify/assert"
)

var ()

func getPgDSN() *DSN {
	return &DSN{
		Driver:   "postgres",
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     5432,
		SSLMode:  false,
		Username: "test",
		Password: "test",
	}
}

func TestNewDBInitializer(t *testing.T) {
	// Test case 1: Postgres driver
	t.Run("Postgres driver", func(t *testing.T) {

		dbInitializer, _ := NewDBInitializer(getPgDSN())

		assert.NotNil(t, dbInitializer, "Expected non-nil dbInitializer")
		assert.IsType(t, &PostgresInitializer{}, dbInitializer, "Expected dbInitializer to be of type *PostgresInitializer")
	})

}

func tearDownUser(dbInitializer IDBInitializer, username string) {
	// Drop the user
	dbInitializer.dropDatabase(username)
}

func TestPostgresInitializer_initUser(t *testing.T) {

	// Test case 1: Valid input
	t.Run("Valid pg init user", func(t *testing.T) {
		dbInitializer, _ := NewDBInitializer(getPgDSN())
		defer tearDownUser(dbInitializer, "test_user")
		// Initialize the user
		err := dbInitializer.initUser("test_user", "test_password")

		// Assert the result
		assert.NoError(t, err, "Expected no error")
	})
}
