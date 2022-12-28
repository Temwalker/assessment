package expense

import (
	"github.com/Temwalker/assessment/database"
	"github.com/lib/pq"
)

func CreateExpenseTable(d *database.DB) error {
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

func InsertExpense(d *database.DB, ex *Expense) error {
	row := d.Database.QueryRow("INSERT INTO expenses (title,amount,note,tags) values ($1,$2,$3,$4) RETURNING id",
		ex.Title, ex.Amount, ex.Note, pq.Array(&ex.Tags))
	return row.Scan(&ex.ID)
}

func UpdateExpenseByID(d *database.DB, rowId int, ex *Expense) error {
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

func SelectExpenseByID(d *database.DB, rowId int, ex *Expense) error {
	stmt, err := d.Database.Prepare("SELECT id,title,amount,note,tags FROM expenses where id=$1")
	if err != nil {
		return err
	}
	defer stmt.Close()
	row := stmt.QueryRow(rowId)
	return row.Scan(&ex.ID, &ex.Title, &ex.Amount, &ex.Note, pq.Array(&ex.Tags))
}

func SelectAllExpenses(d *database.DB, expenses *[]Expense) error {
	stmt, err := d.Database.Prepare("SELECT * FROM expenses;")
	if err != nil {
		return err
	}
	rows, err := stmt.Query()
	if err != nil {
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
