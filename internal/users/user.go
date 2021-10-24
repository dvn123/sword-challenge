package users

import (
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"sword-challenge/internal/util"
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
	auth.GET("/users/:user-id", service.getUser)
	public.POST("/users", service.createUser)
	public.POST("/login", service.loginUser)
	return service
}

func (userService *Service) getUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("user-id"))
	if err != nil {
		userService.Logger.Errorf("Failed to parse user ID: %v", err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	authUser, _ := c.Get(util.UserContextKey)
	_, err = CheckIdsMatchOrIsManager(authUser, id)
	if err != nil {
		c.Status(http.StatusForbidden)
		return
	}

	users, err := userService.getUserFromStore(id)
	if err != nil {
		userService.Logger.Errorf("Failed to get user from storage: %v", err)
		c.JSON(http.StatusInternalServerError, nil) //todo error object
		return
	}

	c.JSON(http.StatusOK, users)
}

func (userService *Service) getUserByToken(c *gin.Context, token string) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		userService.Logger.Errorf("Failed to parse user ID: %v", err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}
	users, err := userService.getUserFromStore(id)
	if err != nil {
		userService.Logger.Errorf("Failed to get user from storage: %v", err)
		c.JSON(http.StatusInternalServerError, nil) //todo error object
		return
	}

	c.JSON(http.StatusOK, users)
}

func (userService *Service) createUser(c *gin.Context) {
	user := &User{}

	if err := c.BindJSON(user); err != nil {
		userService.Logger.Errorf("Failed to parse user request body: %v", err)
		return
	}

	user, err := userService.addUserToStore(user)
	if err != nil {
		userService.Logger.Errorf("Failed to add user to storage: %v", err)
		c.JSON(http.StatusInternalServerError, nil) //todo error object
		return
	}
	c.JSON(http.StatusCreated, user)
}