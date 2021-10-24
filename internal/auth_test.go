package internal

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sword-challenge/internal/util"
	"testing"
	"time"
)

func TestAuth2(t *testing.T) {
	w := httptest.NewRecorder()
	c, e := gin.CreateTestContext(w)
	db, _, _ := sqlmock.New()
	t.Cleanup(func() {
		db.Close()
	})

	u, _ := url.Parse("http://localhost:8080/v1/tasks/1")
	req := &http.Request{
		URL:    u,
		Header: make(http.Header),
	}
	c.Request = req
	//c.Params = append(c.Params, gin.Param{Key: "task-id", Value: "1"})
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal(err)
	}
	server, err := NewServer(sqlx.NewDb(db, "mysql"), logger.Sugar(), e)
	if err != nil {
		t.Fatal(err)
	}
	server.requireAuthentication(c)
	c.Writer.Flush()

	assert.Equal(t, 401, w.Code)
}

func makeRequestWithValidTokenAndCheckStatusCode2(method string, path string, token string, expectedStatus int, s *httptest.Server, mock sqlmock.Sqlmock) func(t *testing.T) {
	return func(t *testing.T) {
		client := &http.Client{}
		req, _ := http.NewRequest(
			method,
			s.URL+path,
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
			mockTokenInDatabase(mock, token, "manager")
		}

		get, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, expectedStatus, get.StatusCode)
	}
}

func mockTokenInDatabase(mock sqlmock.Sqlmock, token string, role string) {
	rows := sqlmock.NewRows([]string{"id", "username", "role.name", "role.id"}).AddRow(1, "joao", role, 2)

	mock.ExpectQuery(
		"SELECT user.id, user.username, role.name as 'role.name', role.id as 'role.id' FROM users user INNER JOIN tokens t on user.id = t.user_id LEFT JOIN roles role on user.role_id = role.id WHERE t.uuid = .;").
		WithArgs(token).
		WillReturnRows(rows)
}
