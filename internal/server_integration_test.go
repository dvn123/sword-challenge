package internal

import (
	"bytes"
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"sword-challenge/internal/util"
	"testing"
	"time"
)

type IntegrationTestSuite struct {
	suite.Suite
	db              *sqlx.DB
	sqlmock         sqlmock.Sqlmock
	s               *httptest.Server
	rabbitContainer *testcontainers.Container
}

func (s *IntegrationTestSuite) SetupSuite() {
	db, mock, _ := sqlmock.New()
	logger, _ := zap.NewDevelopment()
	sqlxDb := sqlx.NewDb(db, "mysql")
	s.db = sqlxDb
	s.sqlmock = mock

	router := gin.New()

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "rabbitmq:3-management-alpine",
		ExposedPorts: []string{"5672/tcp", "15672/tcp"},
		WaitingFor:   wait.ForLog("Server startup complete"),
	}
	rabbitC, _ := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	s.rabbitContainer = &rabbitC

	endpoint, _ := rabbitC.PortEndpoint(ctx, "5672", "")
	conn, _ := amqp.Dial("amqp://guest:guest@" + endpoint)
	rabbitChan, _ := conn.Channel()

	server, _ := NewServer(sqlx.NewDb(db, "mysql"), logger.Sugar(), router, rabbitChan, "6368616e676520746869732070617373")

	server.notificationService.StartConsumer()
	s.s = httptest.NewServer(server.router)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	(*s.rabbitContainer).Terminate(context.Background())
	s.db.Close()
}

func (s *IntegrationTestSuite) TestNoficationsAreConsumedWhenTaskIsCompleted() {
	client := &http.Client{}
	body := bytes.NewReader([]byte(`{"id":1, "completedDate": "2021-10-23T22:50:23Z", "summary": "a", "user": {"id": 1, "username": "a"}}`))

	req, _ := http.NewRequest(http.MethodPut, s.s.URL+"/api/v1/tasks/1", body)
	token := "123"
	req.Header.Add(util.AuthHeader, token)

	userRows := sqlmock.NewRows([]string{"id", "username", "role.name", "role.id"}).AddRow("1", "joao", "technician", 2)
	s.sqlmock.ExpectQuery(expectedFetchUserByTokenSQL).WithArgs(token).WillReturnRows(userRows)

	rows := sqlmock.NewRows([]string{"id", "summary", "completed_date", "user.id", "user.username"}).AddRow(1, "1", nil, 1, "joel")
	s.sqlmock.ExpectQuery("SELECT").WillReturnRows(rows)
	s.sqlmock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(5, 1))
	ti := time.Date(2011, 1, 1, 1, 1, 1, 1, time.UTC)
	updatedRows := sqlmock.NewRows([]string{"id", "summary", "completed_date", "user.id", "user.username"}).AddRow(1, "1", &ti, 5, "joel")
	s.sqlmock.ExpectQuery("SELECT").WillReturnRows(updatedRows)

	managerRows := sqlmock.NewRows([]string{"id", "username", "role.name", "role.id"}).AddRow("1", "joao", "manager", 2).AddRow("2", "joel", "manager", 2)
	s.sqlmock.ExpectQuery("SELECT").WithArgs("manager").WillReturnRows(managerRows)

	response, err := client.Do(req)
	if err != nil {
		s.T().Fatal(err)
	}

	assert.Equal(s.T(), 200, response.StatusCode)

	//Because the notification has no side effect, there's nothing to assert here so we just sleep to make sure the logs are printed
	// Maybe we could use a logger interface and mock it, but I don't think it's worth as this would have a side effect in a real life use case
	time.Sleep(time.Second)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
