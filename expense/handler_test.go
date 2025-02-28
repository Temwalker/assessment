//go:build unit

package expense

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Temwalker/assessment/database"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestCreateHandler(t *testing.T) {
	t.Run("Create Handler Success (DB Connnection OK , Create Table OK)", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		mock.ExpectPing().WillReturnError(nil)
		mock.ExpectExec("CREATE TABLE IF NOT EXISTS expenses (.+)").WillReturnResult(driver.ResultNoRows)
		d, _ := database.GetDB()
		d.Database = db
		assert.NotPanics(t, func() { NewHandler() })
	})

	t.Run("Create Handler but handler can not Create Table Should Panic", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		mock.ExpectPing().WillReturnError(nil)
		mock.ExpectExec("CREATE TABLE IF NOT EXISTS expenses (.+)").WillReturnError(sql.ErrConnDone)
		d, _ := database.GetDB()
		d.Database = db
		assert.Panics(t, func() { NewHandler() })
	})

	t.Run("Create Handler but handler can not get DB connection Should Panic", func(t *testing.T) {
		assert.Panics(t, func() { NewHandler() })
	})
}

func TestCloseHandler(t *testing.T) {
	t.Run("Close Handler Success", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		mock.ExpectPing().WillReturnError(nil)
		mock.ExpectExec("CREATE TABLE IF NOT EXISTS expenses (.+)").WillReturnResult(driver.ResultNoRows)
		mock.ExpectClose().WillReturnError(nil)
		d, _ := database.GetDB()
		d.Database = db
		h := NewHandler()
		err = h.Close()
		assert.NoError(t, err)
	})
	t.Run("Close Handler Fail", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		mock.ExpectPing().WillReturnError(nil)
		mock.ExpectExec("CREATE TABLE IF NOT EXISTS expenses (.+)").WillReturnResult(driver.ResultNoRows)
		mock.ExpectClose().WillReturnError(assert.AnError)
		d, _ := database.GetDB()
		d.Database = db
		h := NewHandler()
		err = h.Close()
		assert.Error(t, err)

	})
}

func TestCreateExpense(t *testing.T) {
	t.Run("Create Expense Return HTTP StatusCreated and Created Expense", func(t *testing.T) {
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
		h := Handler{
			Storage: &database.DB{Database: db},
		}

		err = h.CreateExpenseHandler(c)

		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Equal(t, string(expected), strings.TrimSpace(rec.Body.String()))
		}
	})

	t.Run("Create Expense with none JSON Return HTTP StatusBadRequest", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/expenses", strings.NewReader("1234"))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		db, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		h := Handler{
			Storage: &database.DB{Database: db},
		}

		err = h.CreateExpenseHandler(c)

		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("Create Expense with Empty JSON Return HTTP StatusBadRequest", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/expenses", strings.NewReader(""))
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		h := Handler{
			Storage: &database.DB{},
		}

		err := h.CreateExpenseHandler(c)

		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("Create Expense When DB Close Return HTTP Internal Error", func(t *testing.T) {
		want := Err{Msg: "Internal error"}
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
			WillReturnError(sql.ErrConnDone)
		h := Handler{
			Storage: &database.DB{Database: db},
		}

		err = h.CreateExpenseHandler(c)

		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
			assert.Equal(t, string(expected), strings.TrimSpace(rec.Body.String()))
		}
	})

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

		h := Handler{
			Storage: &database.DB{Database: db},
		}

		err = h.GetExpenseByIdHandler(c)

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

		h := Handler{
			Storage: &database.DB{},
		}

		err := h.GetExpenseByIdHandler(c)

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

		h := Handler{
			Storage: &database.DB{Database: db},
		}

		err = h.GetExpenseByIdHandler(c)

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

		h := Handler{
			Storage: &database.DB{Database: db},
		}

		err = h.GetExpenseByIdHandler(c)

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

		h := Handler{
			Storage: &database.DB{Database: db},
		}

		err = h.UpdateExpenseByIDHandler(c)

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

		h := Handler{
			Storage: &database.DB{},
		}

		err := h.UpdateExpenseByIDHandler(c)

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

			h := Handler{
				Storage: &database.DB{},
			}

			err := h.UpdateExpenseByIDHandler(c)

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

		h := Handler{
			Storage: &database.DB{Database: db},
		}

		err = h.UpdateExpenseByIDHandler(c)

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

		h := Handler{
			Storage: &database.DB{Database: db},
		}

		err = h.UpdateExpenseByIDHandler(c)

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

		h := Handler{
			Storage: &database.DB{Database: db},
		}

		err = h.GetAllExpensesHandler(c)

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

		h := Handler{
			Storage: &database.DB{Database: db},
		}

		err = h.GetAllExpensesHandler(c)
		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusOK, rec.Code)
			respEx := []Expense{}
			json.Unmarshal(rec.Body.Bytes(), &respEx)
			assert.Equal(t, expected, len(respEx))
		}
	})

	t.Run("Get All Expenses but can not scan query into variable Return HTTP Internal Error", func(t *testing.T) {
		want := Err{"Internal error"}
		expected, _ := json.Marshal(want)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockReturnRows := sqlmock.NewRows([]string{"id", "title", "amount", "note"}).
			AddRow(1, "strawberry smoothie", 79.00, "night market promotion discount 10 bath")
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		mock.ExpectPrepare("SELECT (.+) FROM expenses").
			ExpectQuery().
			WillReturnRows(mockReturnRows)

		h := Handler{
			Storage: &database.DB{Database: db},
		}

		err = h.GetAllExpensesHandler(c)
		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
			assert.Equal(t, string(expected), strings.TrimSpace(rec.Body.String()))
		}
	})

	t.Run("Get All Expenses but DB close when query Return HTTP Internal Error", func(t *testing.T) {
		want := Err{"Internal error"}
		expected, _ := json.Marshal(want)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		mock.ExpectPrepare("SELECT (.+) FROM expenses").ExpectQuery().
			WillReturnError(sql.ErrConnDone)

		h := Handler{
			Storage: &database.DB{Database: db},
		}

		err = h.GetAllExpensesHandler(c)
		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
			assert.Equal(t, string(expected), strings.TrimSpace(rec.Body.String()))
		}
	})

	t.Run("Get All Expenses but DB close when prepare Return HTTP Internal Error", func(t *testing.T) {
		want := Err{"Internal error"}
		expected, _ := json.Marshal(want)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		mock.ExpectPrepare("SELECT (.+) FROM expenses").WillReturnError(sql.ErrConnDone)

		h := Handler{
			Storage: &database.DB{Database: db},
		}

		err = h.GetAllExpensesHandler(c)
		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
			assert.Equal(t, string(expected), strings.TrimSpace(rec.Body.String()))
		}
	})

}
