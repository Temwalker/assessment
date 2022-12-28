package database

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestDatabase(t *testing.T) {
	t.Run("Get DB Success", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		mock.ExpectPing().WillReturnError(nil)
		d, _ := GetDB()
		d.Database = db
		_, err = GetDB()
		defer d.CloseDB()
		assert.NoError(t, err)
	})
	t.Run("Get DB Error", func(t *testing.T) {
		_, err := GetDB()
		assert.Error(t, err)
	})
}
