//go:build integration && db

package expense

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestDBCreateUser(t *testing.T) {
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

func TestDBCreateUserWithNoneJson(t *testing.T) {
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

func TestDBCreateUserWithEmptyJson(t *testing.T) {
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

func TestDBCreateUserWithNoConnection(t *testing.T) {
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
