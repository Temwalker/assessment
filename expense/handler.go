package expense

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (d *DB) CreateExpenseHandler(c echo.Context) error {
	ex := Expense{}
	err := c.Bind(&ex)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	err = d.InsertExpense(&ex)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, ex)
}
