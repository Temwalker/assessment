package expense

type Expense struct {
	ID     int      `json:"id"`
	Title  string   `json:"title"`
	Amount float64  `json:"amount"`
	Note   string   `json:"note"`
	Tags   []string `json:"tags"`
}

type Storage interface {
	CreateTable()
	InsertExpense(ex Expense) Expense
}

func CreateExpense(s Storage, ex Expense) Expense {
	return s.InsertExpense(ex)
}
