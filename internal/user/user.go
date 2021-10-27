package user

import (
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"net/http"
	"sword-challenge/internal/util"
	"time"
)

type Role struct {
	ID   string `json:"id,omitempty" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type User struct {
	ID       int    `json:"id,omitempty" binding:"required"`
	Role     *Role  `json:"role,omitempty"`
	Username string `json:"username"`
}

type Service struct {
	DB     *sqlx.DB
	Logger *zap.SugaredLogger
}

func NewService(db *sqlx.DB, logger *zap.SugaredLogger) *Service {
	service := &Service{DB: db, Logger: logger}
	return service
}

func (s *Service) SetupRoutes(publicAPI *gin.RouterGroup) {
	usersAPI := publicAPI.Group("")
	usersAPI.Use(gin.Logger())
	usersAPI.POST("/login", s.loginUser)
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

	c.SetCookie(util.AuthCookie, token, int(time.Hour), "/", "localhost", true, true)
	c.Status(http.StatusOK)
}
