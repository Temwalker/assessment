package database

import (
	"database/sql"
	"os"
	"sync"

	_ "github.com/lib/pq"
)

var once sync.Once

type DB struct {
	Database *sql.DB
}

var dbInstance *DB

func initDB() *DB {
	database, _ := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	dbInstance = &DB{
		Database: database,
	}
	return dbInstance
}

func (d *DB) checkConnectionDB() error {
	err := d.Database.Ping()
	if err != nil {
		d.Database, _ = sql.Open("postgres", os.Getenv("DATABASE_URL"))
		err = d.Database.Ping()
	}
	return err
}

func GetDB() (*DB, error) {
	var err error
	if dbInstance == nil {
		once.Do(
			func() {
				dbInstance = initDB()
				err = dbInstance.Database.Ping()
			})
	} else {
		err = dbInstance.checkConnectionDB()
	}

	return dbInstance, err
}

func (d *DB) CloseDB() error {
	return d.Database.Close()
}
