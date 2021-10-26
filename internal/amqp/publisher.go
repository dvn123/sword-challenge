package amqp

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"sword-challenge/internal/task"
)

type Publisher struct {
	RabbitChannel      *amqp.Channel
	Logger             *zap.SugaredLogger
	NotificationsQueue string
}

func (r *Publisher) PublishTask(t task.Notification) error {
	jsonTask, err := json.Marshal(t)
	if err != nil {
		r.Logger.Warnw("Failed to marshal task to JSON when sending notification", "error", err)
		return err
	}

	err = r.RabbitChannel.Publish("", r.NotificationsQueue, false, false, amqp.Publishing{ContentType: gin.MIMEJSON, Body: jsonTask})
	if err != nil {
		r.Logger.Warnw("Failed to publish task completion notification", "error", err)
		return err
	}
	return nil
}
