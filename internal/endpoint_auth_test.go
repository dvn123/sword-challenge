package internal

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sword-challenge/internal/util"
	"testing"
	"time"
)

func TestAuth(t *testing.T) {
	db, mock, _ := sqlmock.New()
	t.Cleanup(func() {
		db.Close()
	})

	logger := zap.NewNop()
	router := gin.Default()

	server, err := NewServer(sqlx.NewDb(db, "mysql"), logger.Sugar(), router, nil)
	if err != nil {
		t.Fatal(err)
	}
	s := httptest.NewServer(server.router)
	t.Cleanup(s.Close)
	t.Parallel()

	testData := [][]interface{}{
		// tasks
		{http.MethodGet, "/tasks/1", "", 401},
		{http.MethodGet, "/tasks/1", "1", 500},
		{http.MethodPut, "/tasks/1", "", 401},
		{http.MethodPut, "/tasks/1", "1", 400},
		{http.MethodPost, "/tasks", "", 401},
		{http.MethodPost, "/tasks", "1", 400},

		// users
		{http.MethodGet, "/users/1", "", 401},
		{http.MethodGet, "/users/1", "1", 500},
		{http.MethodPost, "/users", "", 400},
		{http.MethodPost, "/login", "", 400},
	}
	for _, test := range testData {
		test := test
		t.Run("shouldReturn"+strconv.Itoa(test[3].(int))+"for"+test[0].(string)+test[1].(string), makeRequestWithValidTokenAndCheckStatusCode(test[0].(string), test[1].(string), test[2].(string), test[3].(int), s, mock))

	}
}

func makeRequestWithValidTokenAndCheckStatusCode(method string, path string, token string, expectedStatus int, s *httptest.Server, mock sqlmock.Sqlmock) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		client := &http.Client{}
		req, _ := http.NewRequest(
			method,
			s.URL+"/api/v1"+path,
			nil,
		)
		if token != "" {
			req.Header.Add(util.AuthHeader, token)
		}
		if token != "" {
			req.AddCookie(&http.Cookie{
				Name:       util.AuthCookie,
				Value:      token,
				Path:       "",
				Domain:     "",
				Expires:    time.Now().Add(time.Hour),
				RawExpires: "",
				MaxAge:     0,
				Secure:     false,
				HttpOnly:   false,
				SameSite:   0,
				Raw:        "",
				Unparsed:   nil,
			})
		}

		if token != "" {
			rows := sqlmock.NewRows([]string{"id", "username", "role.name", "role.id"}).AddRow(1, "joao", "admin", 2)

			mock.ExpectQuery(
				"SELECT user.id, user.username, role.name as 'role.name', role.id as 'role.id' FROM users user INNER JOIN tokens t on user.id = t.user_id LEFT JOIN roles role on user.role_id = role.id WHERE t.uuid = .;").
				WithArgs(token).
				WillReturnRows(rows)
		}

		get, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, expectedStatus, get.StatusCode)
	}
}
