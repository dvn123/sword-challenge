package tasks

import (
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"sword-challenge/internal/users"
	"sword-challenge/internal/util"
	"time"
)

type task struct {
	ID            int         `json:"id,omitempty"`
	Summary       string      `json:"summary,omitempty"`
	CreatedDate   *time.Time  `json:"createdDate" db:"created_date"`
	StartedDate   *time.Time  `json:"startedDate" db:"started_date"`
	CompletedDate *time.Time  `json:"completedDate" db:"completed_date"`
	User          *users.User `json:"user,omitempty"`
}

type Service struct {
	DB          *sqlx.DB
	Logger      *zap.SugaredLogger
	UserService *users.Service
}

func NewService(router *gin.RouterGroup, userService *users.Service, db *sqlx.DB, logger *zap.SugaredLogger) *Service {
	taskService := &Service{UserService: userService, DB: db, Logger: logger}
	router.GET("/tasks/:task-id", taskService.getTasks)
	router.PUT("/tasks/:task-id", taskService.updateTask)
	router.POST("/tasks", taskService.createTask)
	return taskService
}

func (taskService *Service) getTasks(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("task-id"))
	if err != nil {
		taskService.Logger.Errorf("Failed to parse task ID: %v", err)
		c.Status(http.StatusBadRequest) //todo error
		return
	}
	task, err := taskService.getTaskFromStore(id)
	if err != nil {
		taskService.Logger.Errorf("Failed to get task from storage: %v", err)
		c.Status(http.StatusInternalServerError) //todo error object
		return
	}

	authUser, _ := c.Get(util.UserContextKey)
	_, err = users.CheckIdsMatchOrIsManager(authUser, id)
	if err != nil {
		c.Status(http.StatusForbidden)
		return
	}

	c.JSON(http.StatusOK, task)
}

func (taskService *Service) createTask(c *gin.Context) {
	task := &task{}

	if err := c.BindJSON(task); err != nil {
		taskService.Logger.Errorf("Failed to parse task request body: %v", err)
		return
	}

	authUser, _ := c.Get(util.UserContextKey)
	_, err := users.CheckIdsMatchOrIsManager(authUser, task.User.ID)
	if err != nil {
		c.Status(http.StatusForbidden)
		return
	}

	task, err = taskService.addTaskToStore(task)
	if err != nil {
		taskService.Logger.Errorf("Failed to add task to storage: %v", err)
		c.JSON(http.StatusInternalServerError, nil) //todo error object
		return
	}
	c.JSON(http.StatusCreated, task)
}

func (taskService *Service) updateTask(c *gin.Context) {
	task := &task{}
	id, err := strconv.Atoi(c.Param("task-id"))
	if err != nil {
		taskService.Logger.Errorf("Failed to parse task ID: %v", err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	if err := c.BindJSON(task); err != nil {
		taskService.Logger.Errorf("Failed to parse task from body while updating: %v", err)
		return
	}
	task.ID = id

	authUser, _ := c.Get(util.UserContextKey)
	currentUser, err := users.CheckIdsMatchOrIsManager(authUser, task.User.ID)
	if err != nil {
		c.Status(http.StatusForbidden)
		return
	}

	taskToUpdate, err := taskService.getTaskFromStore(task.ID)
	if err != nil {
		// Let's assume our server works very nicely and the only possible error is the task not being found, additional error handling would be here in a production here
		taskService.Logger.Errorf("task not found, id: %v: %v", id, err)
		c.Status(http.StatusNotFound)
		return
	}

	// Check whether the task belongs to the user making the change or if the user is manager, we can only do this after we fetch the task from the database
	// We could also do this by checking
	if currentUser.Role.Name != util.AdminRole && taskToUpdate.User.ID != currentUser.ID {
		c.Status(http.StatusForbidden)
		return
	}

	task, err = taskService.updateTaskInStore(task)
	if err != nil {
		taskService.Logger.Errorf("Failed to update task in storage: %v", err)
		c.JSON(http.StatusInternalServerError, nil) //todo error object
		return
	}
	c.JSON(http.StatusOK, task)
}
