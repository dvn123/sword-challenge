package user

import (
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type Role struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
}

type User struct {
	ID       int    `json:"id,omitempty"`
	Role     *Role  `json:"role,omitempty"`
	Username string `json:"username"`
}

type Service struct {
	DB     *sqlx.DB
	Logger *zap.SugaredLogger
}

func NewService(auth *gin.RouterGroup, public *gin.RouterGroup, db *sqlx.DB, logger *zap.SugaredLogger) *Service {
	service := &Service{DB: db, Logger: logger}
	usersAPI := auth.Group("")
	usersAPI.Use(gin.Logger())
	public.POST("/login", service.loginUser)
	return service
}

func (s *Service) loginUser(c *gin.Context) {
	user := &User{}

	if err := c.BindJSON(user); err != nil {
		s.Logger.Infow("Failed to parse user request body", "error", err)
		return
	}

	token, err := s.authenticateUser(user.ID)
	// Here we would handle the error where the user doesn't exist differently, but since this is not a required endpoint for the API...
	if err != nil || token == "" {
		s.Logger.Warnw("Failed to add user to storage", "error", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.SetCookie("auth-token", token, int(time.Hour), "/", "localhost", true, true)
	c.Status(http.StatusOK)
}
