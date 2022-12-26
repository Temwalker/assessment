package expense

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func (d *DB) CreateExpenseHandler(c echo.Context) error {
	ex := Expense{}
	err := c.Bind(&ex)
	if err != nil || ex.checkEmptyField() {
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
	return ex.returnByIDHandler(c, d.SelectExpenseByID(intVar, &ex))

}

func (d *DB) UpdateExpenseByIDHandler(c echo.Context) error {
	id := c.Param("id")
	intVar, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Msg: "ID is not numeric"})
	}
	ex := Expense{}
	err = c.Bind(&ex)
	if err != nil || ex.checkEmptyField() {
		return c.JSON(http.StatusBadRequest, Err{Msg: "Invalid request body"})
	}
	return ex.returnByIDHandler(c, d.UpdateExpenseByID(intVar, &ex))
}

func (d *DB) GetAllExpensesHandler(c echo.Context) error {
	expenses := []Expense{}
	err := d.SelectAllExpenses(&expenses)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Msg: "Internal error"})
	}
	return c.JSON(http.StatusOK, expenses)
}
