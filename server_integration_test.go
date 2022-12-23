//go:build integration && server

package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
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

func request(method, url string, body io.Reader) *Response {
	req, _ := http.NewRequest(method, url, body)
	req.Header.Add("Authorization", "November 10, 2009")
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

func TestServerCreateUser(t *testing.T) {
	body := bytes.NewBufferString(`{
		"title": "strawberry smoothie",
		"amount": 79,
		"note": "night market promotion discount 10 bath", 
		"tags": ["food", "beverage"]
	}`)
	res := request(http.MethodPost, uri("expenses"), body)
	got := expense.Expense{}
	err := res.Decode(&got)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusCreated, res.StatusCode)
		assert.Equal(t, "strawberry smoothie", got.Title)
		assert.Greater(t, got.ID, int(0))
	}
}

func TestServerCreateUserWithNoneJson(t *testing.T) {
	body := bytes.NewBufferString("1234")
	res := request(http.MethodPost, uri("expenses"), body)
	got := expense.Expense{}
	err := res.Decode(&got)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	}
}
