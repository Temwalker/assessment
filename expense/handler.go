package expense

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func getIDParam(c echo.Context) (int, bool, error) {
	id := c.Param("id")
	intVar, err := strconv.Atoi(id)
	if err != nil {
		return 0, true, c.JSON(http.StatusBadRequest, Err{Msg: "ID is not numeric"})
	}
	return intVar, false, nil
}

func bindRequestBody(c echo.Context, ex *Expense) (bool, error) {
	err := c.Bind(&ex)
	if err != nil || ex.checkEmptyField() {
		return true, c.JSON(http.StatusBadRequest, Err{Msg: "Invalid request body"})
	}
	return false, nil
}

func returnExpenseByID(err error, c echo.Context, ex Expense) error {
	if err == nil {
		return c.JSON(http.StatusOK, ex)
	}
	if err.Error() == sql.ErrNoRows.Error() {
		return c.JSON(http.StatusBadRequest, Err{Msg: "Expense not found"})
	}
	return c.JSON(http.StatusInternalServerError, Err{Msg: "Internal error"})
}

func returnExpensesList(err error, c echo.Context, expenses []Expense) error {
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Msg: "Internal error"})
	}
	return c.JSON(http.StatusOK, expenses)
}

func (d *DB) CreateExpenseHandler(c echo.Context) error {
	ex := Expense{}
	ifErr, respErr := bindRequestBody(c, &ex)
	if ifErr {
		return respErr
	}
	err := d.InsertExpense(&ex)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Msg: "Internal error"})
	}
	return c.JSON(http.StatusCreated, ex)
}

func (d *DB) GetExpenseByIdHandler(c echo.Context) error {
	intVar, ifErr, respErr := getIDParam(c)
	if ifErr {
		return respErr
	}
	ex := Expense{}
	err := d.SelectExpenseByID(intVar, &ex)
	return returnExpenseByID(err, c, ex)
}

func (d *DB) UpdateExpenseByIDHandler(c echo.Context) error {
	intVar, ifErr, respErr := getIDParam(c)
	if ifErr {
		return respErr
	}
	ex := Expense{}
	ifErr, respErr = bindRequestBody(c, &ex)
	if ifErr {
		return respErr
	}
	err := d.UpdateExpenseByID(intVar, &ex)
	return returnExpenseByID(err, c, ex)
}

func (d *DB) GetAllExpensesHandler(c echo.Context) error {
	expenses := []Expense{}
	err := d.SelectAllExpenses(&expenses)
	return returnExpensesList(err, c, expenses)
}
