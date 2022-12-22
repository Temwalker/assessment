package expense

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (d *DB) CreateExpenseHandler(c echo.Context) error {
	ex := Expense{}
	err := c.Bind(&ex)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Msg: "Invalid request body"})
	}
	err = d.InsertExpense(&ex)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Msg: "Internal error"})
	}
	return c.JSON(http.StatusCreated, ex)
}
