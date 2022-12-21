package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

type DB struct {
	URL string
}

func New(url string) *DB {
	db := &DB{
		URL: url,
	}
	db.CreateTable()
	return db
}

func (db DB) CreateTable() {
	database, err := sql.Open("postgres", db.URL)
	if err != nil {
		log.Fatal("Connect to database error :", err)
	}
	defer database.Close()

	createTb := `
	CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY , name TEXT , age INT);
	`

	_, err = database.Exec(createTb)
	if err != nil {
		log.Fatal("can't create table :", err)
	}

	log.Println("Table Created")

}
