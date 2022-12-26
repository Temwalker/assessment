package expense

import (
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

func (d *DB) CreateExpenseHandler(c echo.Context) error {
	ex := Expense{}
	ifErr, respErr := ex.bindRequestBody(c)
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
	return ex.returnByIDHandler(c, d.SelectExpenseByID(intVar, &ex))
}

func (d *DB) UpdateExpenseByIDHandler(c echo.Context) error {
	intVar, ifErr, respErr := getIDParam(c)
	if ifErr {
		return respErr
	}
	ex := Expense{}
	ifErr, respErr = ex.bindRequestBody(c)
	if ifErr {
		return respErr
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
