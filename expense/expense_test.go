package expense

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

	got := CreateExpense(ex)

	assert.EqualValues(t, want, got)
}
