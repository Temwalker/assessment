package expense

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func (d *DB) CreateExpenseHandler(c echo.Context) error {
	ex := Expense{}
	err := c.Bind(&ex)
	if err != nil || checkEmptyField(ex) {
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
	err = d.SelectExpenseByID(intVar, &ex)
	if err == nil {
		return c.JSON(http.StatusOK, ex)
	}
	if err.Error() == sql.ErrNoRows.Error() {
		return c.JSON(http.StatusBadRequest, Err{Msg: "Expense not found"})
	}
	return c.JSON(http.StatusInternalServerError, Err{Msg: "Internal error"})

}

func (d *DB) UpdateExpenseByIDHandler(c echo.Context) error {
	id := c.Param("id")
	intVar, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Msg: "ID is not numeric"})
	}
	ex := Expense{}
	err = c.Bind(&ex)
	if err != nil || checkEmptyField(ex) {
		return c.JSON(http.StatusBadRequest, Err{Msg: "Invalid request body"})
	}
	err = d.UpdateExpenseByID(intVar, &ex)
	if err == nil {
		return c.JSON(http.StatusOK, ex)
	}
	if err.Error() == sql.ErrNoRows.Error() {
		return c.JSON(http.StatusBadRequest, Err{Msg: "Expense not found"})
	}
	return c.JSON(http.StatusInternalServerError, Err{Msg: "Internal error"})
}

func (d *DB) GetAllExpensesHandler(c echo.Context) error {
	expenses := []Expense{}
	err := d.SelectAllExpenses(&expenses)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Msg: "Internal error"})
	}
	return c.JSON(http.StatusOK, expenses)
}
