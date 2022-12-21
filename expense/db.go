package expense

import (
	"database/sql"
	"log"

	"github.com/lib/pq"
)

type DB struct {
	URL string
}

func NewDB(url string) *DB {
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
	CREATE TABLE IF NOT EXISTS expenses (
		id SERIAL PRIMARY KEY,
		title TEXT,
		amount FLOAT,
		note TEXT,
		tags TEXT[]
	);`

	_, err = database.Exec(createTb)
	if err != nil {
		log.Fatal("can't create table :", err)
	}
}

func (db DB) InsertExpense(ex Expense) Expense {
	database, err := sql.Open("postgres", db.URL)
	if err != nil {
		log.Fatal("Connect to database error :", err)
	}
	defer database.Close()

	row := database.QueryRow("INSERT INTO users (title,amount,note,tags) values ($1,$2,$3,$4) RETURNING id", ex.Title, ex.Amount, ex.Note, pq.Array(&ex.Tags))
	err = row.Scan(&ex.ID)
	if err != nil {
		log.Fatal("can't insert :", err)
	}
	return ex
}
