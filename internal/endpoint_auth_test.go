package internal

import (
	"bytes"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/http/httptest"
	serverAmqp "sword-challenge/internal/amqp"
	"sword-challenge/internal/task"
	"sword-challenge/internal/user"
	"sword-challenge/internal/util"
	"testing"
	"time"
)

// TestAuth is a test suite to check for the most common authentication cases:
// * Unauthorized when no token is passed
// * Forbidden when user ID does not match the one passed in the request
// * Forbidden when role isn't manager
// * Any status other than 401 and 403 when auth and authz conditions are met
func TestAuth(t *testing.T) {
	db, mock, _ := sqlmock.New()
	t.Cleanup(func() {
		db.Close()
	})

	logger := zap.NewNop()
	router := gin.New()
	sqlxDB := sqlx.NewDb(db, "mysql")

	pub := &serverAmqp.Publisher{Logger: logger.Sugar(), NotificationsQueue: "tasks"}
	userService := user.NewService(sqlxDB, logger.Sugar())

	tasksService := task.NewService(userService, sqlxDB, pub, logger.Sugar(), "6368616e676520746869732070617373")

	server := &SwordChallengeServer{
		router:              router,
		server:              nil,
		db:                  sqlxDB,
		logger:              logger.Sugar(),
		userService:         userService,
		tasksService:        tasksService,
		notificationService: nil,
	}
	server.SetupRoutes()

	s := httptest.NewServer(server.router)
	t.Cleanup(s.Close)

	taskWithUserID1 := []byte(`{"id":1, "summary": "a", "user": {"id": 1, "username": "a"}}`)

	testData := [][]interface{}{
		// tasks
		{http.MethodGet, "/tasks", 0, "", nil, 401},
		{http.MethodGet, "/tasks", 1, "manager", nil, 500},

		{http.MethodPut, "/tasks/1", 0, "", nil, 401},
		{http.MethodPut, "/tasks/1", 2, "technician", taskWithUserID1, 403},
		{http.MethodPut, "/tasks/1", 2, "manager", taskWithUserID1, 500},
		{http.MethodPut, "/tasks/1", 1, "manager", taskWithUserID1, 500},

		{http.MethodDelete, "/tasks/1", 0, "", nil, 401},
		{http.MethodDelete, "/tasks/1", 2, "technician", nil, 403},
		{http.MethodDelete, "/tasks/1", 2, "manager", nil, 500},

		{http.MethodPost, "/tasks", 0, "", nil, 401},
		{http.MethodPost, "/tasks", 2, "technician", taskWithUserID1, 403},
		{http.MethodPost, "/tasks", 2, "manager", taskWithUserID1, 500},
		{http.MethodPost, "/tasks", 1, "manager", taskWithUserID1, 500},

		// users
		{http.MethodPost, "/login", 0, "", nil, 400},
	}
	for _, test := range testData {
		test := test
		t.Run(
			fmt.Sprintf("shouldReturn%dForUrl%s%s%s", test[5].(int), test[0].(string), test[1].(string), test[3].(string)), // TODO: make this string a bit prettier
			makeRequestWithValidTokenAndCheckStatusCode(test[0].(string), test[1].(string), test[2].(int), test[3].(string), test[4], test[5].(int), s, mock),
		)
	}
}

// Generic auth test, tokenUserID = 0 means no token will be set
func makeRequestWithValidTokenAndCheckStatusCode(method string, path string, tokenUserID int, tokenUserRole string, reqBody interface{}, expectedStatus int, s *httptest.Server, mock sqlmock.Sqlmock) func(t *testing.T) {
	return func(t *testing.T) {
		client := &http.Client{}

		var body io.Reader
		if reqBody != nil {
			body = bytes.NewReader(reqBody.([]byte))
		}

		req, err := http.NewRequest(method, s.URL+"/api/v1"+path, body)
		if err != nil {
			t.Fatalf("Could not create request: %v", err)
		}

		token := "123"
		if tokenUserID != 0 {
			req.Header.Add(util.AuthHeader, token)
		}
		if tokenUserID != 0 {
			req.AddCookie(&http.Cookie{Name: util.AuthCookie, Value: token, Expires: time.Now().Add(time.Hour)})
		}

		if tokenUserID != 0 {
			rows := sqlmock.NewRows([]string{"id", "username", "role.name", "role.id"}).AddRow(tokenUserID, "joao", tokenUserRole, 2)
			mock.ExpectQuery(expectedFetchUserByTokenSQL).WithArgs(token).WillReturnRows(rows)
		}

		response, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, expectedStatus, response.StatusCode)
	}
}
