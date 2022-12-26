package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Temwalker/assessment/expense"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func AuthVerify(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Request().Header.Get(echo.HeaderAuthorization) != "November 10, 2009" {
			return c.JSON(http.StatusUnauthorized, "")
		}

		if err := next(c); err != nil {
			c.Error(err)
		}
		return nil
	}
}

func main() {
	fmt.Println("start at port:", os.Getenv("PORT"))

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(AuthVerify)
	db := expense.GetDB()

	e.POST("/expenses", db.CreateExpenseHandler)
	e.GET("/expenses/:id", db.GetExpenseByIdHandler)
	e.PUT("/expenses/:id", db.UpdateExpenseByIDHandler)
	e.GET("/expenses", db.GetAllExpensesHandler)
	go func() {
		if err := e.Start(os.Getenv("PORT")); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	<-shutdown
	db.DiscDB()
	fmt.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
