package controller

import (
	"github.com/zncdata-labs/zncdata-stack-operator/internal/util"
	"os"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"github.com/stretchr/testify/assert"
)

const (
	DRIVER_POSTGRES = "postgres"
	DRIVER_MYSQL    = "mysql"
)

func getPgDSN() *DSN {
	getenv := os.Getenv("ENV")
	if getenv == "local" {
		return getLocalPgDSN()
	}
	return getEnvDSN(DRIVER_POSTGRES)
}
func getMysqlDSN() *DSN {
	getenv := os.Getenv("ENV")
	if getenv == "local" {
		return getLocalMysqlDSN()
	}
	return getEnvDSN(DRIVER_MYSQL)
}

func getLocalPgDSN() *DSN {
	return &DSN{
		Driver:   DRIVER_POSTGRES,
		Host:     "127.0.0.1",
		Port:     "5432",
		SSLMode:  false,
		Username: "root",
		Password: "123456",
	}
}
func getLocalMysqlDSN() *DSN {
	return &DSN{
		Driver:   DRIVER_MYSQL,
		Host:     "127.0.0.1",
		Port:     "3306",
		SSLMode:  false,
		Username: "root",
		Password: "123456",
		Database: "mysql",
	}
}
func getEnvDSN(driverName string) *DSN {
	switch driverName {
	case DRIVER_POSTGRES:
		return &DSN{
			Driver:   DRIVER_POSTGRES,
			Host:     os.Getenv("PG_DB_HOST"),
			Port:     os.Getenv("PG_DB_PORT"),
			SSLMode:  os.Getenv("PG_DB_SSLMODE") == "true",
			Username: os.Getenv("PG_DB_USERNAME"),
			Password: os.Getenv("PG_DB_PASSWORD"),
		}
	case DRIVER_MYSQL:
		return &DSN{
			Driver:   DRIVER_MYSQL,
			Host:     os.Getenv("MYSQL_DB_HOST"),
			Port:     os.Getenv("MYSQL_DB_PORT"),
			SSLMode:  os.Getenv("MYSQL_DB_SSLMODE") == "true",
			Username: os.Getenv("MYSQL_DB_USERNAME"),
			Password: os.Getenv("MYSQL_DB_PASSWORD"),
		}

	}
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
	initializer, err := NewDBInitializer(getPgDSN())
	if err != nil {
		t.Error("new db initializer", "err: ", err)
	}
	t.Log("new db initializer", "initializer: ", initializer)

	randomStr := strings.ToLower(util.GenerateRandomStr(4))
	username := "test_user" + randomStr
	password := "test_password"
	dbname := "test_db" + randomStr

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
	initializer, err := NewDBInitializer(getMysqlDSN())
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
