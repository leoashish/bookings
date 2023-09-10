package driver

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgconn"

	_ "github.com/jackc/pgx/v4/stdlib"

	_ "github.com/jackc/pgx/v4"
)

type DB struct {
	SQL *sql.DB
}

var dbConn = &DB{}

const maxOpenDBConn = 10
const maxIdleDBConn = 5
const maxDBLifeTime = time.Minute

func ConnectSQL(dsn string) (*DB, error) {
	d, err := NewDataBase(dsn)

	if err != nil {
		return nil, err
	}

	d.SetMaxOpenConns(maxOpenDBConn)
	d.SetMaxIdleConns(maxIdleDBConn)
	d.SetConnMaxLifetime(maxDBLifeTime)

	dbConn.SQL = d
	err = testDB(d)

	if err != nil {
		return nil, err
	}
	return dbConn, nil
}

// test the DB connection
func testDB(d *sql.DB) error {
	err := d.Ping()
	if err != nil {
		return err
	}
	return nil
}

// NewDatabase: Creates a new database for the application
func NewDataBase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)

	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
