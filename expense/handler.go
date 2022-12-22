package expense

import (
	"net/http"
	"reflect"

	"github.com/labstack/echo/v4"
)

func (d *DB) CreateExpenseHandler(c echo.Context) error {
	ex := Expense{}
	err := c.Bind(&ex)
	emtpyEx := Expense{}
	if err != nil || reflect.DeepEqual(ex, emtpyEx) {
		return c.JSON(http.StatusBadRequest, Err{Msg: "Invalid request body"})
	}
	err = d.InsertExpense(&ex)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Msg: "Internal error"})
	}
	return c.JSON(http.StatusCreated, ex)
}
