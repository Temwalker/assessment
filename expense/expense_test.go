package expense

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockDB struct {
	Storage
	Expenses []Expense
}

func (m mockDB) InsertExpense(ex Expense) Expense {
	m.Expenses = append(m.Expenses, ex)
	ex.ID = len(m.Expenses)
	return ex
}

func TestSingleCreate(t *testing.T) {
	ex := Expense{
		Title:  "strawberry smoothie",
		Amount: 79,
		Note:   "night market promotion discount 10 bath",
		Tags:   []string{"food", "beverage"},
	}
	want := Expense{
		ID:     1,
		Title:  "strawberry smoothie",
		Amount: 79,
		Note:   "night market promotion discount 10 bath",
		Tags:   []string{"food", "beverage"},
	}

	m := mockDB{
		Expenses: []Expense{},
	}

	got := CreateExpense(m, ex)

	assert.EqualValues(t, want, got)
}
