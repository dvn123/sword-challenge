package task

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

func (s *Service) mustGetTaskID(c *gin.Context) (int, error) {
	id, err := strconv.Atoi(c.Param("task-id"))
	if err != nil || id == 0 {
		s.logger.Infow("Failed to parse task ID", "error", err)
		c.Status(http.StatusBadRequest)
		return 0, err
	}
	return id, nil
}

type LogPublisher struct {
	Logger *zap.SugaredLogger
}

func (r *LogPublisher) PublishTask(t Notification) error {
	r.Logger.Infow("Task published by LogPublisher", "taskId", t.ID)
	return nil
}
