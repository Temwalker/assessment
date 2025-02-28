//go:build integration && server

package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/Temwalker/assessment/expense"
	"github.com/stretchr/testify/assert"
)

func uri(paths ...string) string {
	host := "http://" + os.Getenv("ASSESSMENT_SERVER")
	if paths == nil {
		return host
	}

	url := append([]string{host}, paths...)
	return strings.Join(url, "/")
}

func request(method, url string, auth string, body io.Reader) *Response {
	req, _ := http.NewRequest(method, url, body)
	req.Header.Add("Authorization", auth)
	req.Header.Add("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	return &Response{res, err}
}

type Response struct {
	*http.Response
	err error
}

func (r *Response) Decode(v interface{}) error {
	if r.err != nil {
		return r.err
	}
	return json.NewDecoder((r.Body)).Decode(v)
}

func (r *Response) DecodeString() (string, error) {
	if r.err != nil {
		return "", r.err
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func seedExpense() (expense.Expense, error) {
	body := bytes.NewBufferString(`{
		"title": "strawberry smoothie",
		"amount": 79,
		"note": "night market promotion discount 10 bath", 
		"tags": ["food", "beverage"]
	}`)
	res := request(http.MethodPost, uri("expenses"), "November 10, 2009", body)
	got := expense.Expense{}
	err := res.Decode(&got)
	if err != nil {
		return got, err
	}
	return got, nil
}

func TestServerCreateExpense(t *testing.T) {
	t.Run("Create Expense Return HTTP StatusCreated and Created Expense", func(t *testing.T) {
		body := bytes.NewBufferString(`{
			"title": "strawberry smoothie",
			"amount": 79,
			"note": "night market promotion discount 10 bath", 
			"tags": ["food", "beverage"]
		}`)
		res := request(http.MethodPost, uri("expenses"), "November 10, 2009", body)
		got := expense.Expense{}
		err := res.Decode(&got)
		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusCreated, res.StatusCode)
			assert.Equal(t, "strawberry smoothie", got.Title)
			assert.Greater(t, got.ID, int(0))
		}
	})

	t.Run("Create Expense with none JSON Return HTTP StatusBadRequest", func(t *testing.T) {
		body := bytes.NewBufferString("1234")
		res := request(http.MethodPost, uri("expenses"), "November 10, 2009", body)
		got := expense.Expense{}
		err := res.Decode(&got)
		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		}
	})
}

func TestServerGetExpenseByID(t *testing.T) {
	seed, err := seedExpense()
	if err != nil {
		t.Fatal("can't seed expense : ", err)
	}
	tests := []struct {
		testname   string
		auth       string
		id         string
		httpStatus int
		want       interface{}
	}{
		{"Get Expense By ID Return HTTP OK and Query Expense", "November 10, 2009", strconv.Itoa(seed.ID), http.StatusOK, seed},
		{"Get Expense By ID but not found Return HTTP Status Bad Request", "November 10, 2009", "0", http.StatusBadRequest, expense.Err{Msg: "Expense not found"}},
		{"Get Expense By ID but Authorization failed Return HTTP Status Unauthorized", "HELLO", strconv.Itoa(seed.ID), http.StatusUnauthorized, ""},
	}
	for _, tt := range tests {
		t.Run(tt.testname, func(t *testing.T) {
			expected, _ := json.Marshal(tt.want)
			res := request(http.MethodGet, uri("expenses", tt.id), tt.auth, nil)
			got, err := res.DecodeString()
			if assert.NoError(t, err) {
				assert.Equal(t, tt.httpStatus, res.StatusCode)
				assert.Equal(t, string(expected), strings.TrimSpace(got))
			}
		})
	}
}

func TestServerUpdateExpenseByID(t *testing.T) {
	seed, err := seedExpense()
	if err != nil {
		t.Fatal("can't seed expense : ", err)
	}
	wantOK := expense.Expense{
		ID:     seed.ID,
		Title:  "apple smoothie",
		Amount: 89,
		Note:   "no discount",
		Tags:   []string{"beverage"},
	}
	tests := []struct {
		testname   string
		auth       string
		id         string
		testdata   string
		httpStatus int
		want       interface{}
	}{
		{"Update Expense By ID Return HTTP OK and Query Expense", "November 10, 2009", strconv.Itoa(seed.ID),
			`{
			"id": ` + strconv.Itoa(seed.ID) + `,
			"title": "apple smoothie",
			"amount": 89,
			"note": "no discount", 
			"tags": ["beverage"]}`, http.StatusOK, wantOK},
		{"Update Expense By ID but not found Return HTTP Status Bad Request", "November 10, 2009", "0",
			`{
			"id": ` + strconv.Itoa(0) + `,
			"title": "apple smoothie",
			"amount": 89,
			"note": "no discount", 
			"tags": ["beverage"]}`, http.StatusBadRequest, expense.Err{Msg: "Expense not found"}},
		{"Update Expense By ID but Authorization failed Return HTTP Status Unauthorized", "HELLO", strconv.Itoa(seed.ID),
			`{
			"id": ` + strconv.Itoa(seed.ID) + `,
			"title": "apple smoothie",
			"amount": 89,
			"note": "no discount", 
			"tags": ["beverage"]}`, http.StatusUnauthorized, ""},
	}
	for _, tt := range tests {
		t.Run(tt.testname, func(t *testing.T) {
			expected, _ := json.Marshal(tt.want)
			res := request(http.MethodPut, uri("expenses", tt.id), tt.auth, bytes.NewBufferString(tt.testdata))
			got, err := res.DecodeString()
			if assert.NoError(t, err) {
				assert.Equal(t, tt.httpStatus, res.StatusCode)
				assert.Equal(t, string(expected), strings.TrimSpace(got))
			}
		})
	}
}

func TestServerGetAllExpenses(t *testing.T) {
	_, err := seedExpense()
	if err != nil {
		t.Fatal("can't seed expense : ", err)
	}
	t.Run("Get All Expenses Return HTTP OK and Query Expenses", func(t *testing.T) {
		res := request(http.MethodGet, uri("expenses"), "November 10, 2009", nil)
		got := []expense.Expense{}
		err := res.Decode(&got)
		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Less(t, 0, len(got))
		}
	})

	t.Run("Get All Expenses but Authorization failed Return HTTP Status Unauthorized", func(t *testing.T) {
		res := request(http.MethodGet, uri("expenses"), "HELLO", nil)
		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
		}
	})

}
