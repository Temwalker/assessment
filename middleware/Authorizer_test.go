package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestAuthorizer(t *testing.T) {
	t.Run("HeaderAuthorization Pass Return HTTP StatusOK", func(t *testing.T) {
		e := echo.New()
		e.Use(Authorizer)
		e.GET("/", func(c echo.Context) error {
			return c.String(http.StatusOK, "")
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add(echo.HeaderAuthorization, "November 10, 2009")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
	t.Run("HeaderAuthorization Pass But Error occurs in handler Return HTTP Internal Server Err", func(t *testing.T) {
		e := echo.New()
		e.Use(Authorizer)
		e.GET("/", func(c echo.Context) error {
			return assert.AnError
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add(echo.HeaderAuthorization, "November 10, 2009")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("HeaderAuthorization Fail Return HTTP StatusUnauthorized", func(t *testing.T) {
		e := echo.New()
		e.Use(Authorizer)
		e.GET("/", func(c echo.Context) error {
			return c.String(http.StatusOK, "")
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add(echo.HeaderAuthorization, "HELLO")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
