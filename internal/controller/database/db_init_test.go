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
		Port:     "5432",
		SSLMode:  false,
		Username: "test",
		Password: "test",
	}
}
func getLocalPgDSN() *DSN {
	return &DSN{
		Driver:   "postgres",
		Host:     "127.0.0.1",
		Port:     "5432",
		SSLMode:  false,
		Username: "root",
		Password: "123456",
	}
}
func getLocalMysqlDSN() *DSN {
	return &DSN{
		Driver:   "mysql",
		Host:     "127.0.0.1",
		Port:     "3306",
		SSLMode:  false,
		Username: "root",
		Password: "123456",
		Database: "mysql",
	}
}
func getEnvDSN() *DSN {
	return &DSN{
		Driver:   os.Getenv("DB_DRIVER"),
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		SSLMode:  os.Getenv("DB_SSLMODE") == "true",
		Username: os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
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

func Test_Postgres(t *testing.T) {
	initializer, err := NewDBInitializer(getLocalPgDSN())
	if err != nil {
		t.Error("new db initializer", "err: ", err)
	}
	t.Log("new db initializer", "initializer: ", initializer)

	username := "test_user"
	password := "test_password"
	dbname := "test_db"

	err = initializer.initUser(username, password)
	if err != nil {
		t.Error("init user", "err: ", err)
	}
	t.Log("init user", "username: ", username, "password: ", password)

	err = initializer.initDatabase(username, dbname)
	if err != nil {
		t.Error("init database", "err: ", err)
	}
	t.Log("init database", "username: ", "test_user", "dbname: ", "test_db")

	err = initializer.dropDatabase(dbname)
	if err != nil {
		t.Error("drop database", "err: ", err)
	}
	t.Log("drop database", "dbname: ", dbname)

	err = initializer.dropUser("test_user")
	if err != nil {
		t.Error("drop user", "err: ", err)
	}
	t.Log("drop user", "username: ", "test_user")

}

func Test_Mysql(t *testing.T) {
	initializer, err := NewDBInitializer(getLocalMysqlDSN())
	if err != nil {
		t.Error("new db initializer", "err: ", err)
	}
	t.Log("new db initializer", "initializer: ", initializer)

	username := "test_user1111"
	password := "test_password1111"
	dbname := "test_db1111"

	err = initializer.initUser(username, password)
	if err != nil {
		t.Error("init user", "err: ", err)
	}
	t.Log("init user", "username: ", username, "password: ", password)

	err = initializer.initDatabase(username, dbname)
	if err != nil {
		t.Error("init database", "err: ", err)
	}
	t.Log("init database", "username: ", username, "dbname: ", dbname)

	err = initializer.dropDatabase(dbname)
	if err != nil {
		t.Error("drop database", "err: ", err)
	}
	t.Log("drop database", "dbname: ", dbname)

	err = initializer.dropUser(username)
	if err != nil {
		t.Error("drop user", "err: ", err)
	}
	t.Log("drop user", "username: ", username)

}
