package expense

type Expense struct {
	ID     int      `json:"id"`
	Title  string   `json:"title"`
	Amount float64  `json:"amount"`
	Note   string   `json:"note"`
	Tags   []string `json:"tags"`
}

type Err struct {
	Msg string `json:"message"`
}

func checkEmptyField(ex Expense) bool {
	//check only string field
	if len(ex.Title) == 0 {
		return true
	}
	if len(ex.Note) == 0 {
		return true
	}
	if len(ex.Tags) == 0 {
		return true
	}
	return false
}
