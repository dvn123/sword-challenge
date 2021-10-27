package task

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"sword-challenge/internal/user"
	"sword-challenge/internal/util"
	"time"
)

func (s *TaskAPITestSuite) TestUpdateTaskFailureWhenFetchingTaskFromTheDatabase() {
	jsonTask, _ := json.Marshal(task{Summary: "test", CompletedDate: nil, User: &user.User{ID: 2, Username: "o"}})
	req, _ := http.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(jsonTask))

	s.c.Request = req
	s.c.Params = append(s.c.Params, validTaskId)
	s.c.Set(util.UserContextKey, &user.User{ID: 1, Role: &user.Role{Name: "manager"}})

	s.sqlmock.ExpectQuery(getTaskSQL).WillReturnError(fmt.Errorf("e"))

	s.service.updateTask(s.c)
	s.c.Writer.Flush()

	assert.Equal(s.T(), 500, s.w.Code)
}

func (s *TaskAPITestSuite) TestUpdateTaskWhenTaskDoesntExist() {
	jsonTask, _ := json.Marshal(task{Summary: "test", CompletedDate: nil, User: &user.User{ID: 2, Username: "o"}})
	req, _ := http.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(jsonTask))

	s.c.Request = req
	s.c.Params = append(s.c.Params, validTaskId)
	s.c.Set(util.UserContextKey, &user.User{ID: 1, Role: &user.Role{Name: "manager"}})
	rows := sqlmock.NewRows(taskColumns)
	s.sqlmock.ExpectQuery(getTaskSQL).WillReturnRows(rows)

	s.service.updateTask(s.c)
	s.c.Writer.Flush()

	assert.Equal(s.T(), 404, s.w.Code)
}

func (s *TaskAPITestSuite) TestUpdateTaskWhenTaskExistsButDoesntBelongToNonAdminUser() {
	req, _ := http.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(validJsonTask))

	s.c.Request = req
	s.c.Params = append(s.c.Params, validTaskId)
	s.c.Set(util.UserContextKey, &user.User{ID: 1, Role: &user.Role{Name: "technician"}})
	rows := sqlmock.NewRows(taskColumns).AddRow(1, "1", nil, 2, "joel")
	s.sqlmock.ExpectQuery(getTaskSQL).WillReturnRows(rows)

	s.service.updateTask(s.c)
	s.c.Writer.Flush()

	assert.Equal(s.T(), 403, s.w.Code)
}

func (s *TaskAPITestSuite) TestUpdateTaskFailToUpdateTask() {
	t := time.Date(2011, 1, 1, 1, 1, 1, 1, time.UTC)
	jsonTask, _ := json.Marshal(task{Summary: "test", CompletedDate: &t, User: &user.User{ID: 2, Username: "o"}})
	req, _ := http.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(jsonTask))

	s.c.Request = req
	s.c.Params = append(s.c.Params, gin.Param{Key: "task-id", Value: "1"})
	s.c.Set(util.UserContextKey, &user.User{ID: 1, Role: &user.Role{Name: "manager"}})
	rows := sqlmock.NewRows(taskColumns).AddRow(1, "1", nil, 1, "joel")

	s.sqlmock.ExpectQuery(getTaskSQL).WillReturnRows(rows)
	s.sqlmock.ExpectExec(updateTaskSQL).WillReturnError(fmt.Errorf("a"))

	s.service.updateTask(s.c)
	s.c.Writer.Flush()

	assert.Equal(s.T(), 500, s.w.Code)
}

func (s *TaskAPITestSuite) TestUpdateTaskSuccess() {
	t := time.Date(2011, 1, 1, 1, 1, 1, 1, time.UTC)
	completedTask := task{Summary: "test", CompletedDate: &t, User: &user.User{ID: 1, Username: "o"}}
	et, _ := s.tEncryptor.encryptTask(&completedTask)
	jsonTask, _ := json.Marshal(completedTask)
	req, _ := http.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(jsonTask))

	s.c.Request = req
	s.c.Params = append(s.c.Params, gin.Param{Key: "task-id", Value: "1"})
	s.c.Set(util.UserContextKey, &user.User{ID: 1, Role: &user.Role{Name: "technician"}})
	rows := sqlmock.NewRows(taskColumns).AddRow(1, "1", nil, 1, "joel")

	s.sqlmock.ExpectQuery(getTaskSQL).WillReturnRows(rows)
	s.sqlmock.ExpectExec(updateTaskSQL).WillReturnResult(sqlmock.NewResult(5, 1))
	updatedRows := sqlmock.NewRows(taskColumns).AddRow(1, et.EncryptedSummary, &t, 5, "joel")
	s.sqlmock.ExpectQuery(getTaskSQL).WillReturnRows(updatedRows)

	userRows := sqlmock.NewRows([]string{"id", "username", "role.name", "role.id"}).AddRow(1, "joao", util.AdminRole, 2).AddRow(2, "j", util.AdminRole, 2)
	s.sqlmock.ExpectQuery("SELECT u.id, u.username FROM users u INNER JOIN roles r on u.role_id = r.id WHERE r.name = .+;").WillReturnRows(userRows)

	s.service.updateTask(s.c)
	s.c.Writer.Flush()

	assert.Equal(s.T(), 200, s.w.Code)
	var taskReceived task
	if err := json.Unmarshal(s.w.Body.Bytes(), &taskReceived); err != nil {
		s.T().Fatalf("Failed to parse response body: error: %v", err)
	}
	assert.Equal(s.T(), taskReceived.User.ID, 5)
}
