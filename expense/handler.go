package expense

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/Temwalker/assessment/database"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	Storage *database.DB
}

func NewHandler() Handler {
	db, err := database.GetDB()
	if err != nil {
		log.Panic("Can't connect to DB : ", err)
	}
	err = CreateExpenseTable(db)
	if err != nil {
		log.Panic("Can't create table : ", err)
	}
	return Handler{
		Storage: db,
	}
}

func (h Handler) Close() error {
	err := h.Storage.CloseDB()
	if err != nil {
		log.Println("Can't close DB Connection  : ", err)
	}
	return err
}
func getIDParam(c echo.Context) (int, bool, error) {
	id := c.Param("id")
	intVar, err := strconv.Atoi(id)
	if err != nil {
		return 0, true, c.JSON(http.StatusBadRequest, Err{Msg: "ID is not numeric"})
	}
	return intVar, false, nil
}

func bindRequestBody(c echo.Context, ex *Expense) (bool, error) {
	err := c.Bind(ex)
	if err != nil || checkEmptyField(*ex) {
		return true, c.JSON(http.StatusBadRequest, Err{Msg: "Invalid request body"})
	}
	return false, nil
}

func checkEmptyField(ex Expense) bool {
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

func returnExpenseByID(err error, c echo.Context, ex Expense) error {
	if err == nil {
		return c.JSON(http.StatusOK, ex)
	}
	if err.Error() == sql.ErrNoRows.Error() {
		return c.JSON(http.StatusBadRequest, Err{Msg: "Expense not found"})
	}
	return c.JSON(http.StatusInternalServerError, Err{Msg: "Internal error"})
}

func returnExpenseCreated(err error, c echo.Context, ex Expense) error {
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Msg: "Internal error"})
	}
	return c.JSON(http.StatusCreated, ex)
}

func returnExpensesList(err error, c echo.Context, expenses []Expense) error {
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Msg: "Internal error"})
	}
	return c.JSON(http.StatusOK, expenses)
}

func (h Handler) CreateExpenseHandler(c echo.Context) error {
	ex := Expense{}
	ifErr, respErr := bindRequestBody(c, &ex)
	if ifErr {
		return respErr
	}
	err := InsertExpense(h.Storage, &ex)
	return returnExpenseCreated(err, c, ex)
}

func (h Handler) GetExpenseByIdHandler(c echo.Context) error {
	intVar, ifErr, respErr := getIDParam(c)
	if ifErr {
		return respErr
	}
	ex := Expense{}
	err := SelectExpenseByID(h.Storage, intVar, &ex)
	return returnExpenseByID(err, c, ex)
}

func (h Handler) UpdateExpenseByIDHandler(c echo.Context) error {
	intVar, ifErr, respErr := getIDParam(c)
	if ifErr {
		return respErr
	}
	ex := Expense{}
	ifErr, respErr = bindRequestBody(c, &ex)
	if ifErr {
		return respErr
	}
	err := UpdateExpenseByID(h.Storage, intVar, &ex)
	return returnExpenseByID(err, c, ex)
}

func (h Handler) GetAllExpensesHandler(c echo.Context) error {
	expenses := []Expense{}
	err := SelectAllExpenses(h.Storage, &expenses)
	return returnExpensesList(err, c, expenses)
}
