package task

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
	"sword-challenge/internal/user"
	"sword-challenge/internal/util"
	"testing"
)

func TestTaskHandlers(t *testing.T) {
	db, mock, _ := sqlmock.New()
	t.Cleanup(func() {
		db.Close()
	})

	logger, _ := zap.NewDevelopment()

	sqlxDb := sqlx.NewDb(db, "mysql")
	userService := &user.Service{
		DB:     sqlxDb,
		Logger: logger.Sugar(),
	}

	service := &Service{
		db:          sqlxDb,
		logger:      logger.Sugar(),
		userService: userService,
	}

	t.Run("shouldReceiveRequestedTask", shouldReceiveRequestedTask(service, mock))
	t.Run("shouldCreateRequestedTask", shouldCreateRequestedTask(service, mock))
	t.Run("shouldUpdateRequestedTask", shouldUpdateRequestedTask(service, mock))
	t.Run("shouldDeleteRequestedTask", shouldDeleteRequestedTask(service, mock))
}

func shouldReceiveRequestedTask(service *Service, mock sqlmock.Sqlmock) func(t *testing.T) {
	return func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Params = append(c.Params, gin.Param{
			Key:   "task-id",
			Value: "1",
		})
		c.Set(util.UserContextKey, &user.User{ID: 1, Role: &user.Role{Name: "manager"}})

		rows := sqlmock.NewRows([]string{"id", "completed_date", "user.id", "user.username"}).AddRow(1, nil, 2, "joel")

		mock.ExpectQuery(
			"SELECT t.id, t.completed_date, u.id as 'user.id', u.username as 'user.username' FROM tasks t INNER JOIN users u on t.user_id = u.id WHERE t.id = .+;").
			WithArgs(1).
			WillReturnRows(rows)

		service.getTasks(c)
		c.Writer.Flush()

		assert.Equal(t, 200, w.Code)
		var taskReceived Task
		if err := json.Unmarshal(w.Body.Bytes(), &taskReceived); err != nil {
			t.Fatalf("Failed to parse response body: error: %v", err)
		}
		assert.Equal(t, taskReceived.ID, 1)
	}
}

func shouldCreateRequestedTask(service *Service, mock sqlmock.Sqlmock) func(t *testing.T) {
	return func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonTask, _ := json.Marshal(Task{
			Summary:       "test",
			CompletedDate: nil,
			User: &user.User{
				ID:       1,
				Username: "ol",
			},
		})
		req, _ := http.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(jsonTask))

		c.Request = req
		c.Params = append(c.Params, gin.Param{
			Key:   "task-id",
			Value: "1",
		})
		c.Set(util.UserContextKey, &user.User{ID: 1, Role: &user.Role{Name: "manager"}})

		mock.ExpectExec("INSERT INTO tasks (.+) VALUES (.+);").WillReturnResult(sqlmock.NewResult(5, 1))

		service.createTask(c)
		c.Writer.Flush()

		assert.Equal(t, 201, w.Code)
		var taskReceived Task
		if err := json.Unmarshal(w.Body.Bytes(), &taskReceived); err != nil {
			t.Fatalf("Failed to parse response body: error: %v", err)
		}
		assert.Equal(t, taskReceived.ID, 5)
	}
}

func shouldUpdateRequestedTask(service *Service, mock sqlmock.Sqlmock) func(t *testing.T) {
	return func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonTask, _ := json.Marshal(Task{
			Summary:       "test",
			CompletedDate: nil,
			User:          &user.User{ID: 2, Username: "o"},
		})
		req, _ := http.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(jsonTask))

		c.Request = req
		c.Params = append(c.Params, gin.Param{
			Key:   "task-id",
			Value: "1",
		})
		c.Set(util.UserContextKey, &user.User{ID: 1, Role: &user.Role{Name: "manager"}})
		rows := sqlmock.NewRows([]string{"id", "completed_date", "user.id", "user.username"}).AddRow(1, nil, 1, "joel")

		mock.ExpectQuery("SELECT").WillReturnRows(rows)
		mock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(5, 1))
		updatedRows := sqlmock.NewRows([]string{"id", "completed_date", "user.id", "user.username"}).AddRow(1, nil, 5, "joel")
		mock.ExpectQuery("SELECT").WillReturnRows(updatedRows)

		service.updateTask(c)
		c.Writer.Flush()

		assert.Equal(t, 200, w.Code)
		var taskReceived Task
		if err := json.Unmarshal(w.Body.Bytes(), &taskReceived); err != nil {
			t.Fatalf("Failed to parse response body: error: %v", err)
		}
		assert.Equal(t, taskReceived.User.ID, 5)
	}
}

func shouldDeleteRequestedTask(service *Service, mock sqlmock.Sqlmock) func(t *testing.T) {
	return func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Params = append(c.Params, gin.Param{
			Key:   "task-id",
			Value: "1",
		})
		c.Set(util.UserContextKey, &user.User{ID: 1, Role: &user.Role{Name: "manager"}})

		mock.ExpectExec(
			"DELETE FROM tasks t WHERE t.id = .+;").
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		service.deleteTask(c)
		c.Writer.Flush()

		assert.Equal(t, 200, w.Code)
	}
}
