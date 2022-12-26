//go:build unit

package expense

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestCreateExpense(t *testing.T) {
	want := Expense{
		ID:     1,
		Title:  "strawberry smoothie",
		Amount: 79,
		Note:   "night market promotion discount 10 bath",
		Tags:   []string{"food", "beverage"},
	}
	expected, _ := json.Marshal(want)
	e := echo.New()
	body := bytes.NewBufferString(`{
		"title": "strawberry smoothie",
		"amount": 79,
		"note": "night market promotion discount 10 bath", 
		"tags": ["food", "beverage"]
	}`)
	req := httptest.NewRequest(http.MethodPost, "/expenses", body)
	req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	mock.ExpectQuery("INSERT INTO expenses (.+) RETURNING id").
		WithArgs(want.Title, want.Amount, want.Note, pq.Array(&want.Tags)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mdb := DB{db}

	err = mdb.CreateExpenseHandler(c)

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, string(expected), strings.TrimSpace(rec.Body.String()))
	}

}

func TestCreateExpenseWithNoneJson(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/expenses", strings.NewReader("1234"))
	req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	mdb := DB{db}

	err = mdb.CreateExpenseHandler(c)

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestCreateExpenseWithEmptyJson(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/expenses", strings.NewReader(""))
	req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mdb := DB{}

	err := mdb.CreateExpenseHandler(c)

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestGetExpenseByID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/expenses", nil)
	req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
	t.Run("Get Expense By ID Return HTTP OK and Query Expense", func(t *testing.T) {
		want := Expense{
			ID:     1,
			Title:  "strawberry smoothie",
			Amount: 79,
			Note:   "night market promotion discount 10 bath",
			Tags:   []string{"food", "beverage"},
		}
		expected, _ := json.Marshal(want)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/:id")
		c.SetParamNames("id")
		c.SetParamValues(strconv.Itoa(want.ID))

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		mock.ExpectPrepare("SELECT id,title,amount,note,tags FROM expenses").
			ExpectQuery().WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "amount", "note", "tags"}).AddRow(want.ID, want.Title, want.Amount, want.Note, pq.Array(&want.Tags)))

		mdb := DB{db}

		err = mdb.GetExpenseByIdHandler(c)

		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, string(expected), strings.TrimSpace(rec.Body.String()))
		}
	})

	t.Run("Get Expense By ID(STRING) Return HTTP Status Bad Request", func(t *testing.T) {
		want := Err{"ID is not numeric"}
		expected, _ := json.Marshal(want)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/:id")
		c.SetParamNames("id")
		c.SetParamValues("NumberOne")

		mdb := &DB{}

		err := mdb.GetExpenseByIdHandler(c)

		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, string(expected), strings.TrimSpace(rec.Body.String()))
		}
	})

	t.Run("Get Expense By ID but not found Return HTTP Status Bad Request", func(t *testing.T) {
		want := Err{"Expense not found"}
		expected, _ := json.Marshal(want)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/:id")
		c.SetParamNames("id")
		c.SetParamValues("1")

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		mock.ExpectPrepare("SELECT id,title,amount,note,tags FROM expenses").
			ExpectQuery().WithArgs(1).WillReturnError(sql.ErrNoRows)

		mdb := DB{db}

		err = mdb.GetExpenseByIdHandler(c)

		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, string(expected), strings.TrimSpace(rec.Body.String()))
		}
	})

	t.Run("Get Expense By ID but DB close Return HTTP Internal Error", func(t *testing.T) {
		want := Err{"Internal error"}
		expected, _ := json.Marshal(want)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/:id")
		c.SetParamNames("id")
		c.SetParamValues("1")

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		mock.ExpectPrepare("SELECT id,title,amount,note,tags FROM expenses").WillReturnError(sql.ErrConnDone)

		mdb := DB{db}

		err = mdb.GetExpenseByIdHandler(c)

		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
			assert.Equal(t, string(expected), strings.TrimSpace(rec.Body.String()))
		}
	})

}

func TestUpdateExpenseByID(t *testing.T) {
	e := echo.New()
	t.Run("Update Expense By ID Return HTTP OK and Expense", func(t *testing.T) {
		want := Expense{
			ID:     1,
			Title:  "apple smoothie",
			Amount: 89,
			Note:   "no discount",
			Tags:   []string{"beverage"},
		}
		expected, _ := json.Marshal(want)
		body := bytes.NewBufferString(`{
			"id": 1,
			"title": "apple smoothie",
			"amount": 89,
			"note": "no discount", 
			"tags": ["beverage"]
		}`)
		req := httptest.NewRequest(http.MethodPut, "/expenses", body)
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/:id")
		c.SetParamNames("id")
		c.SetParamValues(strconv.Itoa(want.ID))

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		mock.ExpectPrepare("UPDATE expenses").
			ExpectQuery().WithArgs(want.ID, want.Title, want.Amount, want.Note, pq.Array(&want.Tags)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(want.ID))

		mdb := DB{db}

		err = mdb.UpdateExpenseByIDHandler(c)

		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, string(expected), strings.TrimSpace(rec.Body.String()))
		}
	})
	t.Run("Update Expense By ID(STRING) Return HTTP Status Bad Request", func(t *testing.T) {
		want := Err{"ID is not numeric"}
		expected, _ := json.Marshal(want)
		body := bytes.NewBufferString(`{
			"id": 1,
			"title": "apple smoothie",
			"amount": 89,
			"note": "no discount", 
			"tags": ["beverage"]
		}`)
		req := httptest.NewRequest(http.MethodPut, "/expenses", body)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/:id")
		c.SetParamNames("id")
		c.SetParamValues("NumberOne")

		mdb := &DB{}

		err := mdb.UpdateExpenseByIDHandler(c)

		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, string(expected), strings.TrimSpace(rec.Body.String()))
		}

	})

	validateTests := []struct {
		testname string
		testdata string
	}{
		{"Update Expense By ID but JSON Req have no title Return HTTP Status Bad Request",
			`{
			"id": 1,
			"amount": 89,
			"note": "no discount", 
			"tags": ["beverage"]
		}`},
		{"Update Expense By ID but JSON Req have no note Return HTTP Status Bad Request",
			`{
			"id": 1,
			"title": "apple smoothie",
			"amount": 89,
			"tags": ["beverage"]
		}`},
		{"Update Expense By ID but JSON Req have no tags Return HTTP Status Bad Request",
			`{
			"id": 1,
			"title": "apple smoothie",
			"amount": 89,
			"note": "no discount"
		}`},
		{"Update Expense By ID but JSON Req is empty Return HTTP Status Bad Request", ""},
	}
	for _, tt := range validateTests {
		t.Run(tt.testname, func(t *testing.T) {
			want := Err{Msg: "Invalid request body"}
			expected, _ := json.Marshal(want)
			req := httptest.NewRequest(http.MethodPut, "/expenses", bytes.NewBufferString(tt.testdata))
			req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/:id")
			c.SetParamNames("id")
			c.SetParamValues("1")

			mdb := DB{}

			err := mdb.UpdateExpenseByIDHandler(c)

			if assert.NoError(t, err) {
				assert.Equal(t, http.StatusBadRequest, rec.Code)
				assert.Equal(t, string(expected), strings.TrimSpace(rec.Body.String()))
			}
		})
	}

	t.Run("Update Expense By ID but ID not found Return HTTP Status Bad Request", func(t *testing.T) {
		want := Err{Msg: "Expense not found"}
		expected, _ := json.Marshal(want)
		body := bytes.NewBufferString(`{
			"id": 1,
			"title": "apple smoothie",
			"amount": 89,
			"note": "no discount", 
			"tags": ["beverage"]
		}`)
		req := httptest.NewRequest(http.MethodPut, "/expenses", body)
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/:id")
		c.SetParamNames("id")
		c.SetParamValues("1")

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		mock.ExpectPrepare("UPDATE expenses").
			ExpectQuery().WithArgs(1, "apple smoothie", 89.00, "no discount", pq.Array(&[]string{"beverage"})).
			WillReturnError(sql.ErrNoRows)

		mdb := DB{db}

		err = mdb.UpdateExpenseByIDHandler(c)

		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, string(expected), strings.TrimSpace(rec.Body.String()))
		}
	})

	t.Run("Update Expense By ID but DB close Return HTTP Internal Error", func(t *testing.T) {
		want := Err{"Internal error"}
		expected, _ := json.Marshal(want)
		body := bytes.NewBufferString(`{
			"id": 1,
			"title": "apple smoothie",
			"amount": 89,
			"note": "no discount", 
			"tags": ["beverage"]
		}`)
		req := httptest.NewRequest(http.MethodPut, "/expenses", body)
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/:id")
		c.SetParamNames("id")
		c.SetParamValues("1")

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		mock.ExpectPrepare("UPDATE expenses").WillReturnError(sql.ErrConnDone)

		mdb := DB{db}

		err = mdb.UpdateExpenseByIDHandler(c)

		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
			assert.Equal(t, string(expected), strings.TrimSpace(rec.Body.String()))
		}
	})
}

func TestGetAllExpenses(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/expenses", nil)
	req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
	t.Run("Get All Expenses Return HTTP OK and Query Expenses", func(t *testing.T) {
		expected := 2
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockReturnRows := sqlmock.NewRows([]string{"id", "title", "amount", "note", "tags"}).
			AddRow(1, "strawberry smoothie", 79.00, "night market promotion discount 10 bath", pq.Array([]string{"food", "beverage"})).
			AddRow(2, "apple smoothie", 89.00, "no discount", pq.Array([]string{"beverage"}))
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		mock.ExpectPrepare("SELECT (.+) FROM expenses").
			ExpectQuery().
			WillReturnRows(mockReturnRows)

		mdb := DB{db}

		err = mdb.GetAllExpensesHandler(c)

		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusOK, rec.Code)
			respEx := []Expense{}
			json.Unmarshal(rec.Body.Bytes(), &respEx)
			assert.Equal(t, expected, len(respEx))
		}
	})

	t.Run("Get All Expenses but no rows return Return HTTP OK and empty slice", func(t *testing.T) {
		expected := 0
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockReturnRows := sqlmock.NewRows([]string{"id", "title", "amount", "note", "tags"})
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		mock.ExpectPrepare("SELECT (.+) FROM expenses").
			ExpectQuery().
			WillReturnRows(mockReturnRows)

		mdb := DB{db}

		err = mdb.GetAllExpensesHandler(c)
		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusOK, rec.Code)
			respEx := []Expense{}
			json.Unmarshal(rec.Body.Bytes(), &respEx)
			assert.Equal(t, expected, len(respEx))
		}
	})

}
