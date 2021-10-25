package internal

import (
	"go.uber.org/zap"
	"sword-challenge/internal/task"
)

type LogPublisher struct {
	logger *zap.SugaredLogger
}

func (r *LogPublisher) PublishTask(t task.NotificationTask) {
	r.logger.Infow("Task published by LogPublisher", "taskId", t.ID)
}
