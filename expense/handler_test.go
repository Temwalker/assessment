package expense

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	e := echo.New()
	reqEx := &Expense{
		Title:  "strawberry smoothie",
		Amount: 79,
		Note:   "night market promotion discount 10 bath",
		Tags:   []string{"food", "beverage"},
	}
	ex, _ := json.Marshal(reqEx)
	req := httptest.NewRequest(http.MethodPost, "/expenses", strings.NewReader(string(ex)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	mock.ExpectQuery("INSERT INTO expenses (.+) RETURNING id").
		WithArgs(reqEx.Title, reqEx.Amount, reqEx.Note, pq.Array(&reqEx.Tags)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	want := &Expense{
		ID:     1,
		Title:  "strawberry smoothie",
		Amount: 79,
		Note:   "night market promotion discount 10 bath",
		Tags:   []string{"food", "beverage"},
	}

	expected, _ := json.Marshal(want)

	mdb := &DB{db}

	err = mdb.CreateExpenseHandler(c)

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, string(expected), strings.TrimSpace(rec.Body.String()))
	}

}
