package expense

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/lib/pq"
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
	err = db.CreateTable()
	if err != nil {
		log.Fatal("Can't create table : ", err)
	}
	return db
}

func (d *DB) reConnectDB() {
	err := d.Database.Ping()
	if err != nil {
		fmt.Println("Reconnecting to DB")
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
			fmt.Println("Creating DB conenction")
			db = initDB()
		} else {
			db.reConnectDB()
		}
	} else {
		db.reConnectDB()
	}

	return db
}

func (d *DB) DiscDB() {
	d.Database.Close()
}

func (d *DB) CreateTable() error {
	createTb := `
	CREATE TABLE IF NOT EXISTS expenses (
		id SERIAL PRIMARY KEY,
		title TEXT,
		amount FLOAT,
		note TEXT,
		tags TEXT[]
	);`

	_, err := d.Database.Exec(createTb)
	return err
}

func (d *DB) InsertExpense(ex *Expense) error {
	row := d.Database.QueryRow("INSERT INTO expenses (title,amount,note,tags) values ($1,$2,$3,$4) RETURNING id", ex.Title, ex.Amount, ex.Note, pq.Array(&ex.Tags))
	return row.Scan(&ex.ID)
}
