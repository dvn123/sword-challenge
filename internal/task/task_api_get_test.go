package task

import (
	"encoding/json"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"sword-challenge/internal/user"
	"sword-challenge/internal/util"
)

func (s *TaskAPITestSuite) TestGetRequestedTaskDatabaseFailure() {
	s.c.Params = append(s.c.Params, validTaskId)
	s.c.Set(util.UserContextKey, &user.User{ID: 1, Role: &user.Role{Name: "manager"}})

	s.sqlmock.ExpectQuery(getTasksSQL).WithArgs(1).WillReturnError(fmt.Errorf("error"))
	s.service.getTasks(s.c)
	s.c.Writer.Flush()

	assert.Equal(s.T(), 500, s.w.Code)
}

func (s *TaskAPITestSuite) TestGetRequestedTask() {
	s.c.Params = append(s.c.Params, validTaskId)
	s.c.Set(util.UserContextKey, &user.User{ID: 1, Role: &user.Role{Name: "manager"}})

	expectedTask := task{ID: 1, CompletedDate: nil, User: &user.User{ID: 2, Username: "joel"}}
	et, _ := s.tEncryptor.encryptTask(&expectedTask)
	rows := sqlmock.NewRows(taskColumns).AddRow(1, et.EncryptedSummary, nil, 2, "joel")

	s.sqlmock.ExpectQuery(getTasksSQL).WithArgs(1).WillReturnRows(rows)
	s.service.getTasks(s.c)

	s.c.Writer.Flush()

	assert.Equal(s.T(), 200, s.w.Code)
	var taskReceived []task
	if err := json.Unmarshal(s.w.Body.Bytes(), &taskReceived); err != nil {
		s.T().Fatalf("Failed to parse response body: error: %v", err)
	}
	assert.Equal(s.T(), taskReceived[0], expectedTask)
}
