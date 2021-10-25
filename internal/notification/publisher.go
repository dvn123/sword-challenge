package notification

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"sword-challenge/internal/task"
)

type RabbitPublisher struct {
	RabbitChannel      *amqp.Channel
	Logger             *zap.SugaredLogger
	NotificationsQueue string
}

func (r *RabbitPublisher) PublishTask(t task.NotificationTask) {
	jsonTask, err := json.Marshal(t)
	if err != nil {
		r.Logger.Warnw("Failed to marshal task to JSON when sending notification", "error", err)
	}
	if err := r.RabbitChannel.Publish(
		"",
		r.NotificationsQueue,
		false,
		false,
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     gin.MIMEJSON,
			ContentEncoding: "",
			Body:            jsonTask,
			DeliveryMode:    amqp.Transient,
			Priority:        0,
		},
	); err != nil {
		r.Logger.Warnw("Failed to publish task completion notification", "error", err)
	}
}
