package expense

import (
	"database/sql"
	"net/http"
	"reflect"
	"strconv"

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

func (d *DB) GetExpenseByIdHandler(c echo.Context) error {
	id := c.Param("id")
	intVar, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Msg: "ID is not numeric"})
	}
	ex := Expense{}
	err = d.SelectExpenseById(intVar, &ex)
	if err == nil {
		return c.JSON(http.StatusOK, ex)
	}
	if err.Error() == sql.ErrNoRows.Error() {
		return c.JSON(http.StatusBadRequest, Err{Msg: "User not found"})
	}
	return c.JSON(http.StatusInternalServerError, Err{Msg: "Internal error"})

}
