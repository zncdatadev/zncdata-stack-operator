package controller

import (
	"database/sql"
	"strconv"
)

type DSN struct {
	Driver   string
	Host     string
	Port     int
	SSLMode  bool
	Username string
	Password string
}

func (d *DSN) String() string {
	var sslMode string
	if d.SSLMode {
		sslMode = "require"
	} else {
		sslMode = "disable"
	}
	return "host=" + d.Host + " port=" + strconv.Itoa(d.Port) + " user=" + d.Username + " password=" + d.Password + " sslmode=" + sslMode
}

type IDBInitializer interface {
	initDatabase(username string, dbname string) error
	initUser(username string, password string) error
	dropDatabase(dbname string) error
	dropUser(username string) error
	setConnection(conn *sql.DB)
}

type DBInitializer struct {
	conn *sql.DB
}

func (d *DBInitializer) initDatabase(username string, dbname string) error {
	panic("implement me")
}

func (d *DBInitializer) initUser(username string, password string) error {
	panic("implement me")
}

func (d *DBInitializer) dropDatabase(dbname string) error {
	_, err := d.conn.Exec("DROP DATABASE " + dbname)
	return err
}

func (d *DBInitializer) dropUser(username string) error {
	_, err := d.conn.Exec("DROP USER " + username)
	return err
}

func (d *DBInitializer) setConnection(conn *sql.DB) {
	d.conn = conn
}

func NewDBInitializer(dsn *DSN) (IDBInitializer, error) {

	var initializer IDBInitializer
	switch dsn.Driver {
	case "postgres":
		initializer = &PostgresInitializer{}
	case "mysql":
		initializer = &MySQLInitializer{}
	default:
		panic("Unsupported driver")
	}
	dsnString := dsn.String()
	db, err := OpenDB(dsn.Driver, &dsnString)
	db.Ping()
	if err != nil {
		return nil, err
	}

	initializer.setConnection(db)

	return initializer, nil
}

func OpenDB(driver string, dsn *string) (*sql.DB, error) {
	db, err := sql.Open(driver, *dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

type PostgresInitializer struct {
	DBInitializer
}

func (d *PostgresInitializer) initDatabase(username string, dbname string) error {
	_, err := d.conn.Exec("CREATE DATABASE " + dbname + " OWNER " + username)
	if err != nil {
		return err
	}
	_, err = d.conn.Exec("GRANT ALL PRIVILEGES ON DATABASE " + dbname + " TO " + username)
	return err
}

func (d *PostgresInitializer) initUser(username string, password string) error {
	_, err := d.conn.Exec("CREATE USER " + username + " WITH PASSWORD '" + password + "'")
	return err
}

type MySQLInitializer struct {
	DBInitializer
}

func (d *MySQLInitializer) initDatabase(username string, dbname string) error {
	_, err := d.conn.Exec("CREATE DATABASE " + dbname)
	if err != nil {
		return err
	}
	_, err = d.conn.Exec("GRANT ALL PRIVILEGES ON " + dbname + ".* TO " + username)
	return err
}

func (d *MySQLInitializer) initUser(username string, password string) error {
	_, err := d.conn.Exec("CREATE USER " + username + " IDENTIFIED BY " + password)
	return err
}
