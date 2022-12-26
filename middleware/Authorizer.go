package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func Authorizer(next echo.HandlerFunc) echo.HandlerFunc {
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
