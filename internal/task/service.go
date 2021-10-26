package task

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"sword-challenge/internal/user"
)

type Service struct {
	db            *sqlx.DB
	logger        *zap.SugaredLogger
	userService   *user.Service
	taskPublisher Publisher
	taskEncryptor *taskCrypto
}

type Publisher interface {
	PublishTask(t Notification) error
}

func NewService(router *gin.RouterGroup, userService *user.Service, db *sqlx.DB, taskPublisher Publisher, logger *zap.SugaredLogger, key string) *Service {
	c, err := NewCrypto(key, logger)
	if err != nil {
		logger.Fatalw("Failed to create task encryptor", "error", err)
	}
	taskService := &Service{userService: userService, db: db, taskPublisher: taskPublisher, taskEncryptor: c, logger: logger}
	router.GET("/tasks/:task-id", taskService.getTasks)
	router.PUT("/tasks/:task-id", taskService.updateTask)
	router.DELETE("/tasks/:task-id", taskService.deleteTask)
	router.POST("/tasks", taskService.createTask)
	return taskService
}
