package user

import (
	"bytes"
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserHandlers(t *testing.T) {
	db, mock, _ := sqlmock.New()
	t.Cleanup(func() {
		db.Close()
	})

	logger, _ := zap.NewDevelopment()

	sqlxDb := sqlx.NewDb(db, "mysql")

	service := &Service{
		DB:     sqlxDb,
		Logger: logger.Sugar(),
	}

	t.Run("shouldLoginUser", shouldLoginUser(service, mock))
}

func shouldLoginUser(service *Service, mock sqlmock.Sqlmock) func(t *testing.T) {
	return func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonUser, _ := json.Marshal(User{ID: 2})
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewReader(jsonUser))
		c.Request = req

		mock.ExpectExec("INSERT INTO tokens").WillReturnResult(sqlmock.NewResult(0, 1))

		service.loginUser(c)
		c.Writer.Flush()

		assert.Equal(t, 200, w.Code)
	}
}
