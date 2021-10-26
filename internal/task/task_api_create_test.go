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
)

func (s *TaskAPITestSuite) TestFailToCreateRequestedTaskInDatabase() {
	jsonTask, _ := json.Marshal(task{Summary: "test", CompletedDate: nil, User: &user.User{ID: 1, Username: "ol"}})
	req, _ := http.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(jsonTask))

	s.c.Request = req
	s.c.Params = append(s.c.Params, gin.Param{Key: "task-id", Value: "1"})
	s.c.Set(util.UserContextKey, &user.User{ID: 1, Role: &user.Role{Name: "manager"}})

	s.sqlmock.ExpectExec(createTaskSQL).WillReturnError(fmt.Errorf("as"))

	s.service.createTask(s.c)
	s.c.Writer.Flush()

	assert.Equal(s.T(), 500, s.w.Code)
}

func (s *TaskAPITestSuite) TestCreateRequestedTask() {
	jsonTask, _ := json.Marshal(task{Summary: "test", CompletedDate: nil, User: &user.User{ID: 1, Username: "ol"}})
	req, _ := http.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(jsonTask))

	s.c.Request = req
	s.c.Params = append(s.c.Params, gin.Param{Key: "task-id", Value: "1"})
	s.c.Set(util.UserContextKey, &user.User{ID: 1, Role: &user.Role{Name: "manager"}})

	s.sqlmock.ExpectExec(createTaskSQL).WillReturnResult(sqlmock.NewResult(5, 1))

	s.service.createTask(s.c)
	s.c.Writer.Flush()

	assert.Equal(s.T(), 201, s.w.Code)
	var taskReceived task
	if err := json.Unmarshal(s.w.Body.Bytes(), &taskReceived); err != nil {
		s.T().Fatalf("Failed to parse response body: error: %v", err)
	}
	assert.Equal(s.T(), taskReceived.ID, 5)
}
