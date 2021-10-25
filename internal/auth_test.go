package internal

import (
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

const expectedFetchUserByTokenSQL = "SELECT u.id, u.username, r.name as 'role.name', r.id as 'role.id' FROM users u INNER JOIN tokens t on u.id = t.user_id LEFT JOIN roles r on u.role_id = r.id WHERE t.uuid = .+;"

func TestAuthMiddleware(t *testing.T) {
	db, mock, _ := sqlmock.New()
	t.Cleanup(func() {
		db.Close()
	})
	logger, _ := zap.NewDevelopment()
	sqlxDb := sqlx.NewDb(db, "mysql")
	server := SwordChallengeServer{db: sqlxDb, logger: logger.Sugar(), userService: &user.Service{DB: sqlxDb}}

	t.Run("shouldReturn401WhenNoTokenIsPresent", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/tasks/1", nil)

		c.Request = req

		server.requireAuthentication(c)
		c.Writer.Flush()

		assert.Equal(t, 401, w.Code)
	})

	t.Run("shouldReturn401WhenTokenExistsButNotFoundInStorage", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/tasks/1", nil)
		token := "123"
		req.AddCookie(&http.Cookie{Name: util.AuthCookie, Value: token})
		c.Request = req
		mock.ExpectQuery(
			expectedFetchUserByTokenSQL).
			WithArgs(token).
			WillReturnError(nil)

		server.requireAuthentication(c)
		c.Writer.Flush()

		assert.Equal(t, 401, w.Code)
	})

	t.Run("shouldSetUserInContextIfTokenIsValid", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/tasks/1", nil)
		token := "1234"
		role := "manager"
		id := 1
		req.Header.Add(util.AuthHeader, token)
		c.Request = req

		rows := sqlmock.NewRows([]string{"id", "username", "role.name", "role.id"}).AddRow(id, "joao", role, 2)

		mock.ExpectQuery(
			expectedFetchUserByTokenSQL).
			WithArgs(token).
			WillReturnRows(rows)

		server.requireAuthentication(c)
		c.Writer.Flush()

		actualUser := c.MustGet(util.UserContextKey).(*user.User)
		assert.Equal(t, actualUser.ID, 1)
		assert.Equal(t, actualUser.Role.Name, role)
	})
}
