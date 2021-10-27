package task

import (
	"bytes"
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"sword-challenge/internal/user"
	"sword-challenge/internal/util"
	"testing"
)

var validTaskId = gin.Param{Key: "task-id", Value: "1"}
var validJsonTask, _ = json.Marshal(task{Summary: "test", CompletedDate: nil, User: &user.User{ID: 1, Username: "o"}})
var req, _ = http.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(validJsonTask))

const getTaskSQL = "SELECT t.id, t.summary, t.completed_date, u.id as 'user.id', u.username as 'user.username' FROM tasks t INNER JOIN users u on t.user_id = u.id WHERE t.id = .+;"
const getTasksSQL = "SELECT t.id, t.summary, t.completed_date, u.id as 'user.id', u.username as 'user.username' FROM tasks t INNER JOIN users u on t.user_id = u.id WHERE t.user_id = .+;"
const deleteTaskSQL = "DELETE FROM tasks t WHERE t.id = .+;"
const createTaskSQL = "INSERT INTO tasks (.+, .+) VALUES (.+, .+);"
const updateTaskSQL = "UPDATE tasks SET user_id = COALESCE(.+, .+), summary = COALESCE(.+, .+), completed_date = .+ WHERE id = .+;"

var taskColumns = []string{"id", "summary", "completed_date", "user.id", "user.username"}

type TaskAPITestSuite struct {
	suite.Suite
	db         *sqlx.DB
	sqlmock    sqlmock.Sqlmock
	service    *Service
	tEncryptor *taskCrypto

	// per test
	c *gin.Context
	w *httptest.ResponseRecorder
}

func (s *TaskAPITestSuite) SetupSuite() {
	db, mock, _ := sqlmock.New()
	logger, _ := zap.NewDevelopment()
	sqlxDb := sqlx.NewDb(db, "mysql")
	s.db = sqlxDb
	s.sqlmock = mock

	userService := &user.Service{
		DB:     sqlxDb,
		Logger: logger.Sugar(),
	}

	c, _ := NewCrypto("6368616e676520746869732070617373", logger.Sugar())
	s.tEncryptor = c
	s.service = &Service{
		db:            sqlxDb,
		logger:        logger.Sugar(),
		userService:   userService,
		taskPublisher: &LogPublisher{Logger: logger.Sugar()},
		taskEncryptor: c,
	}
}

func (s *TaskAPITestSuite) TearDownSuite() {
	s.db.Close()
}

func (s *TaskAPITestSuite) SetupTest() {
	s.w = httptest.NewRecorder()
	c, _ := gin.CreateTestContext(s.w)
	s.c = c
}

func (s *TaskAPITestSuite) TestInvalidTaskIDWhenGettingTask() {
	invalidTaskID := gin.Param{Key: "task-id", Value: "ola"}

	updateTaskRecorder := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(updateTaskRecorder)
	c2.Params = append(c2.Params, invalidTaskID)
	s.service.updateTask(c2)
	c2.Writer.Flush()

	deleteTaskRecorder := httptest.NewRecorder()
	c3, _ := gin.CreateTestContext(deleteTaskRecorder)
	s.service.deleteTask(c3)
	c3.Writer.Flush()

	assert.Equal(s.T(), 400, updateTaskRecorder.Code)
	assert.Equal(s.T(), 400, deleteTaskRecorder.Code)
}

func (s *TaskAPITestSuite) TestFailToParseRequestBodyWhenParsingTask() {
	req, _ := http.NewRequest(http.MethodPost, "/tasks", bytes.NewReader([]byte("asdasd")))

	createTaskRecorder := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(createTaskRecorder)
	c1.Params = append(c1.Params, validTaskId)
	c1.Request = req
	s.service.createTask(c1)
	c1.Writer.Flush()

	updateTaskRecorder := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(updateTaskRecorder)
	c2.Params = append(c2.Params, validTaskId)
	c2.Request = req
	s.service.updateTask(c2)
	c2.Writer.Flush()

	assert.Equal(s.T(), 400, createTaskRecorder.Code)
	assert.Equal(s.T(), 400, updateTaskRecorder.Code)
}

func (s *TaskAPITestSuite) TestDeleteRequestedTask() {
	s.c.Params = append(s.c.Params, validTaskId)
	s.c.Set(util.UserContextKey, &user.User{ID: 1, Role: &user.Role{Name: "manager"}})

	s.sqlmock.ExpectExec(deleteTaskSQL).WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
	s.service.deleteTask(s.c)
	s.c.Writer.Flush()

	assert.Equal(s.T(), 200, s.w.Code)
}

func (s *TaskAPITestSuite) TestDeleteRequestedTaskNotFound() {
	s.c.Params = append(s.c.Params, validTaskId)
	s.c.Set(util.UserContextKey, &user.User{ID: 1, Role: &user.Role{Name: "manager"}})

	s.sqlmock.ExpectExec(deleteTaskSQL).WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 0))

	s.service.deleteTask(s.c)
	s.c.Writer.Flush()

	assert.Equal(s.T(), 404, s.w.Code)
}

func TestTaskTestSuite(t *testing.T) {
	suite.Run(t, new(TaskAPITestSuite))
}
