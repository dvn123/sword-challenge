package task

import (
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"sword-challenge/internal/user"
	"sword-challenge/internal/util"
	"time"
)

type task struct {
	ID            int        `json:"id,omitempty"`
	Summary       string     `json:"summary,omitempty"`
	CreatedDate   *time.Time `json:"createdDate" db:"created_date"`
	StartedDate   *time.Time `json:"startedDate" db:"started_date"`
	CompletedDate *time.Time `json:"completedDate" db:"completed_date"`
	User          *user.User `json:"user,omitempty"`
}

type Service struct {
	db            *sqlx.DB
	logger        *zap.SugaredLogger
	userService   *user.Service
	RabbitChannel *amqp.Channel
}

func NewService(router *gin.RouterGroup, userService *user.Service, db *sqlx.DB, RabbitChannel *amqp.Channel, logger *zap.SugaredLogger) *Service {
	taskService := &Service{userService: userService, db: db, RabbitChannel: RabbitChannel, logger: logger}
	router.GET("/tasks/:task-id", taskService.getTasks)
	router.PUT("/tasks/:task-id", taskService.updateTask)
	router.POST("/tasks", taskService.createTask)
	return taskService
}

func (s *Service) getTasks(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("task-id"))
	if err != nil {
		s.logger.Errorf("Failed to parse task ID: %v", err)
		c.Status(http.StatusBadRequest) //todo error
		return
	}
	task, err := s.getTaskFromStore(id)
	if err != nil {
		s.logger.Errorf("Failed to get task from storage: %v", err)
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
	task := &task{}

	if err := c.BindJSON(task); err != nil {
		s.logger.Errorf("Failed to parse task request body: %v", err)
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
		s.logger.Errorf("Failed to add task to storage: %v", err)
		c.JSON(http.StatusInternalServerError, nil) //todo error object
		return
	}
	c.JSON(http.StatusCreated, task)
}

func (s *Service) updateTask(c *gin.Context) {
	task := &task{}
	id, err := strconv.Atoi(c.Param("task-id"))
	if err != nil {
		s.logger.Errorf("Failed to parse task ID: %v", err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	if err := c.BindJSON(task); err != nil {
		s.logger.Errorf("Failed to parse task from body while updating: %v", err)
		return
	}
	task.ID = id

	authUser, _ := c.Get(util.UserContextKey)
	currentUser, err := user.CheckIdsMatchOrIsManager(authUser, user.GetIDOrZeroValue(task.User))
	if err != nil {
		c.Status(http.StatusForbidden)
		return
	}

	taskToUpdate, err := s.getTaskFromStore(task.ID)
	if err != nil {
		// Let's assume our server works very nicely and the only possible error is the task not being found, additional error handling would be here in a production here
		s.logger.Errorf("task not found, id: %v: %v", id, err)
		c.Status(http.StatusNotFound)
		return
	}

	// Check whether the task belongs to the user making the change or if the user is manager, we can only do this after we fetch the task from the database
	// We could also do this by checking
	if currentUser.Role.Name != util.AdminRole && taskToUpdate.User.ID != currentUser.ID {
		c.Status(http.StatusForbidden)
		return
	}

	updatedTask, err := s.updateTaskInStore(task)
	if err != nil {
		s.logger.Errorf("Failed to update task in storage: %v", err)
		c.JSON(http.StatusInternalServerError, nil) //todo error object
		return
	}

	if taskToUpdate.CompletedDate == nil && updatedTask.CompletedDate != nil {
		go s.publishNotification(updatedTask)
	}
	c.JSON(http.StatusOK, updatedTask)
}

func (s *Service) publishNotification(updatedTask *task) {
	if err := s.RabbitChannel.Publish(
		"",      // publish to an exchange
		"tasks", // routing to 0 or more queues todo
		false,   // mandatory
		false,   // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            []byte(updatedTask.CompletedDate.String()),
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
		},
	); err != nil {
		s.logger.Warnw("Failed to publish task completion notification", "error", err)
	}
}
