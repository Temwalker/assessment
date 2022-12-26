package expense

import (
	"database/sql"
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

func (d *DB) SelectExpenseByID(rowId int, ex *Expense) error {
	stmt, err := d.Database.Prepare("SELECT id,title,amount,note,tags FROM expenses where id=$1")
	if err != nil {
		return err
	}
	defer stmt.Close()
	row := stmt.QueryRow(rowId)
	return row.Scan(&ex.ID, &ex.Title, &ex.Amount, &ex.Note, pq.Array(&ex.Tags))
}

func (d *DB) UpdateExpenseByID(rowId int, ex *Expense) error {
	sqlStatement := `
	UPDATE expenses
	SET title=$2 , amount=$3 , note=$4 , tags=$5
	WHERE id=$1
	RETURNING id;`
	stmt, err := d.Database.Prepare(sqlStatement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	row := stmt.QueryRow(rowId, ex.Title, ex.Amount, ex.Note, pq.Array(&ex.Tags))
	return row.Scan(&ex.ID)
}

func (d *DB) SelectAllExpenses(expenses *[]Expense) error {
	sqlStatement := "SELECT * FROM expenses;"
	stmt, err := d.Database.Prepare(sqlStatement)
	if err != nil {
		log.Println("can't prepare ", err)
		return err
	}
	rows, err := stmt.Query()
	if err != nil {
		log.Println("can't query all expenses", err)
		return err
	}
	for rows.Next() {
		var ex Expense
		err := rows.Scan(&ex.ID, &ex.Title, &ex.Amount, &ex.Note, pq.Array(&ex.Tags))
		if err != nil {
			return err
		}
		*expenses = append(*expenses, ex)
	}

	return nil
}
