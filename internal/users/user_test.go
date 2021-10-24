package users

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
	"sword-challenge/internal/util"
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

	t.Run("shouldReceiveRequestedUser", shouldReceiveRequestedUser(service, mock))
	t.Run("shouldCreateRequestedUser", shouldCreateRequestedUser(service, mock))
}

func shouldReceiveRequestedUser(service *Service, mock sqlmock.Sqlmock) func(t *testing.T) {
	return func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Params = append(c.Params, gin.Param{
			Key:   "user-id",
			Value: "1",
		})
		c.Set(util.UserContextKey, &User{ID: 1, Role: &Role{Name: "manager"}})

		rows := sqlmock.NewRows([]string{"id", "username", "role.name", "role.id"}).AddRow(1, "joao", "manager", 2)

		mock.ExpectQuery(
			"SELECT user.id, user.username, role.name as 'role.name', role.id as 'role.id' FROM users user LEFT JOIN roles role on user.role_id = role.id WHERE user.id = .+").
			WithArgs(1).
			WillReturnRows(rows)

		service.getUser(c)
		c.Writer.Flush()

		assert.Equal(t, 200, w.Code)
		var userReceived User
		if err := json.Unmarshal(w.Body.Bytes(), &userReceived); err != nil {
			t.Fatalf("Failed to parse response body: error: %v", err)
		}
		assert.Equal(t, userReceived.ID, 1)
	}
}

func shouldCreateRequestedUser(service *Service, mock sqlmock.Sqlmock) func(t *testing.T) {
	return func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonUser, _ := json.Marshal(User{
			Username: "ol",
			Role: &Role{
				ID: "1",
			},
		})

		req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewReader(jsonUser))

		c.Request = req
		c.Params = append(c.Params, gin.Param{
			Key:   "user-id",
			Value: "1",
		})
		c.Set(util.UserContextKey, &User{ID: 1, Role: &Role{Name: "manager"}})

		mock.ExpectExec("INSERT INTO users (.+) VALUES (.+, .+);").WillReturnResult(sqlmock.NewResult(5, 1))

		service.createUser(c)
		c.Writer.Flush()

		assert.Equal(t, 201, w.Code)
		var userReceived User
		if err := json.Unmarshal(w.Body.Bytes(), &userReceived); err != nil {
			t.Fatalf("Failed to parse response body: error: %v", err)
		}
		assert.Equal(t, userReceived.ID, 5)
	}
}
