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
	customMiddleware "github.com/Temwalker/assessment/middleware"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	setMiddleware(e)
	db := setRoute(e)
	go startServer(e)
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	<-shutdown
	db.DiscDB()
	shutDownServer(e)
}

func shutDownServer(e *echo.Echo) {
	fmt.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
func startServer(e *echo.Echo) {
	fmt.Println("start at port:", os.Getenv("PORT"))
	if err := e.Start(os.Getenv("PORT")); err != nil && err != http.ErrServerClosed {
		e.Logger.Fatal("shutting down the server")
	}
}

func setRoute(e *echo.Echo) *expense.DB {
	db := expense.GetDB()
	e.POST("/expenses", db.CreateExpenseHandler)
	e.GET("/expenses/:id", db.GetExpenseByIdHandler)
	e.PUT("/expenses/:id", db.UpdateExpenseByIDHandler)
	e.GET("/expenses", db.GetAllExpensesHandler)
	return db
}

func setMiddleware(e *echo.Echo) {
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(customMiddleware.Authorizer)
}
