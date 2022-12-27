package expense

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

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

func (ex Expense) checkEmptyField() bool {
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

func (ex *Expense) bindRequestBody(c echo.Context) (bool, error) {
	err := c.Bind(&ex)
	if err != nil || ex.checkEmptyField() {
		return true, c.JSON(http.StatusBadRequest, Err{Msg: "Invalid request body"})
	}
	return false, nil
}
