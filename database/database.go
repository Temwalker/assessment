package database

import (
	"database/sql"
	"log"
	"os"
	"sync"
)

var lock = &sync.Mutex{}

type DB struct {
	Database *sql.DB
}

var db *DB

func initDB() *DB {
	database, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Connect to database error :", err)
	}
	db = &DB{
		Database: database,
	}
	if err != nil {
		log.Fatal("Can't create table : ", err)
	}
	return db
}

func (d *DB) reConnectDB() {
	err := d.Database.Ping()
	if err != nil {
		d.Database, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
		if err != nil {
			log.Fatal("Connect to database error :", err)
		}
	}
}

func GetDB() *DB {
	if db == nil {
		lock.Lock()
		defer lock.Unlock()
		if db == nil {
			db = initDB()
		} else {
			db.reConnectDB()
		}
	} else {
		db.reConnectDB()
	}

	return db
}

func (d *DB) CloseDB() {
	d.Database.Close()
}
