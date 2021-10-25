package task

import (
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"sword-challenge/internal/user"
	"sword-challenge/internal/util"
	"time"
)

type Task struct {
	ID            int        `json:"id,omitempty"`
	Summary       string     `json:"summary,omitempty"`
	CreatedDate   *time.Time `json:"createdDate" db:"created_date"`
	StartedDate   *time.Time `json:"startedDate" db:"started_date"`
	CompletedDate *time.Time `json:"completedDate" db:"completed_date"`
	User          *user.User `json:"user,omitempty"`
}

type Service struct {
	db          *sqlx.DB
	logger      *zap.SugaredLogger
	userService *user.Service
	//RabbitChannel *amqp.Channel
	taskPublisher Publisher
}

type Publisher interface {
	PublishTask(task Task)
}

//func NewService(router *gin.RouterGroup, userService *user.Service, db *sqlx.DB, RabbitChannel *amqp.Channel, logger *zap.SugaredLogger) *Service {
func NewService(router *gin.RouterGroup, userService *user.Service, db *sqlx.DB, taskPublisher Publisher, logger *zap.SugaredLogger) *Service {
	taskService := &Service{userService: userService, db: db, taskPublisher: taskPublisher, logger: logger}
	//taskService := &Service{userService: userService, db: db, RabbitChannel: RabbitChannel, logger: logger}
	router.GET("/tasks/:task-id", taskService.getTasks)
	router.PUT("/tasks/:task-id", taskService.updateTask)
	router.POST("/tasks", taskService.createTask)
	return taskService
}

func (s *Service) getTasks(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("task-id"))
	if err != nil {
		s.logger.Infow("Failed to parse task ID", "error", err)
		c.Status(http.StatusBadRequest) //todo error
		return
	}
	task, err := s.getTaskFromStore(id)
	if err != nil {
		s.logger.Warnw("Failed to get task from storage", "error", err)
		c.Status(http.StatusInternalServerError) //todo error object
		return
	}

	authUser, _ := c.Get(util.UserContextKey)
	_, err = user.CheckIdsMatchOrIsManager(authUser, id)
	if err != nil {
		c.Status(http.StatusForbidden)
		return
	}

	c.JSON(http.StatusOK, task)
}

func (s *Service) createTask(c *gin.Context) {
	task := &Task{}

	if err := c.BindJSON(task); err != nil {
		s.logger.Infow("Failed to parse task request body", "error", err)
		return
	}

	authUser, _ := c.Get(util.UserContextKey)
	_, err := user.CheckIdsMatchOrIsManager(authUser, task.User.ID)
	if err != nil {
		c.Status(http.StatusForbidden)
		return
	}

	task, err = s.addTaskToStore(task)
	if err != nil {
		s.logger.Warnw("Failed to add task to storage", "error", err)
		c.JSON(http.StatusInternalServerError, nil) //todo error object
		return
	}
	c.JSON(http.StatusCreated, task)
}

func (s *Service) updateTask(c *gin.Context) {
	receivedTask := &Task{}
	id, err := strconv.Atoi(c.Param("task-id"))
	if err != nil {
		s.logger.Infow("Failed to parse task ID", "error", err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	if err := c.BindJSON(receivedTask); err != nil {
		s.logger.Infow("Failed to parse task from body while updating", "error", err)
		return
	}
	receivedTask.ID = id

	authUser, _ := c.Get(util.UserContextKey)
	currentUser, err := user.CheckIdsMatchOrIsManager(authUser, user.GetIDOrZeroValue(receivedTask.User))
	if err != nil {
		c.Status(http.StatusForbidden)
		return
	}

	taskToUpdate, err := s.getTaskFromStore(receivedTask.ID)
	if err != nil {
		// Let's assume our server works very nicely and the only possible error is the task not being found, additional error handling would be here in a production here
		s.logger.Infow("Failed to find task", "taskId", id, "error", err)
		c.Status(http.StatusNotFound)
		return
	}

	// Check whether the task belongs to the user making the change or if the user is manager, we can only do this after we fetch the task from the database
	// We could also do this by checking
	if currentUser.Role.Name != util.AdminRole && taskToUpdate.User.ID != currentUser.ID {
		c.Status(http.StatusForbidden)
		return
	}

	updatedTask, err := s.updateTaskInStore(receivedTask)
	if err != nil {
		s.logger.Warnw("Failed to update task in storage", "error", err)
		c.JSON(http.StatusInternalServerError, nil) //todo error object
		return
	}

	if taskToUpdate.CompletedDate == nil && updatedTask.CompletedDate != nil {
		go s.taskPublisher.PublishTask(*updatedTask)
	}
	c.JSON(http.StatusOK, updatedTask)
}
