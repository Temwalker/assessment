//go:build integration && db

package expense

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestDBCreateExpense(t *testing.T) {
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

	db := GetDB()
	defer db.DiscDB()

	err := db.CreateExpenseHandler(c)
	got := Expense{}
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		json.Unmarshal(rec.Body.Bytes(), &got)
		assert.Greater(t, got.ID, int(0))
	}

}

func TestDBCreateExpenseWithNoneJson(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/expenses", strings.NewReader("1234"))
	req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	db := GetDB()
	defer db.DiscDB()

	err := db.CreateExpenseHandler(c)

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestDBCreateExpenseWithEmptyJson(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/expenses", strings.NewReader(""))
	req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	db := GetDB()
	defer db.DiscDB()

	err := db.CreateExpenseHandler(c)

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestDBCreateExpenseWithNoConnection(t *testing.T) {
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

	db := GetDB()
	db.DiscDB()

	err := db.CreateExpenseHandler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}
}

func seedExpense() (Expense, error) {
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

	db := GetDB()
	defer db.DiscDB()

	err := db.CreateExpenseHandler(c)
	got := Expense{}
	if err != nil {
		return got, err
	}
	json.Unmarshal(rec.Body.Bytes(), &got)
	return got, nil
}

func TestDBGetExpenseByID(t *testing.T) {
	seed, err := seedExpense()
	if err != nil {
		t.Fatal("can't seed expense : ", err)
	}
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/expenses", nil)
	req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
	tests := []struct {
		testname     string
		id           string
		wantClosedDB bool
		httpStatus   int
		want         interface{}
	}{
		{"Get Expense By ID Return HTTP OK and Query Expense", strconv.Itoa(seed.ID), false, http.StatusOK, seed},
		{"Get Expense By ID but not found Return HTTP Status Bad Request", "0", false, http.StatusBadRequest, Err{"Expense not found"}},
		{"Get Expense By ID but DB close Return HTTP Internal Error", "1", true, http.StatusInternalServerError, Err{"Internal error"}},
	}
	for _, tt := range tests {
		t.Run(tt.testname, func(t *testing.T) {
			expected, _ := json.Marshal(tt.want)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/:id")
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			db := GetDB()
			if tt.wantClosedDB {
				db.DiscDB()
			} else {
				defer db.DiscDB()
			}

			err = db.GetExpenseByIdHandler(c)
			if assert.NoError(t, err) {
				assert.Equal(t, tt.httpStatus, rec.Code)
				assert.Equal(t, string(expected), strings.TrimSpace(rec.Body.String()))
			}
		})
	}
}
