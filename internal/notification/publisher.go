package notification

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"sword-challenge/internal/task"
)

type RabbitPublisher struct {
	RabbitChannel *amqp.Channel
	Logger        *zap.SugaredLogger
}

func (r *RabbitPublisher) PublishTask(t task.Task) {
	if err := r.RabbitChannel.Publish(
		"",      // publish to an exchange
		"tasks", // routing to 0 or more queues todo
		false,   // mandatory
		false,   // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            []byte(t.CompletedDate.String()),
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
		},
	); err != nil {
		r.Logger.Warnw("Failed to publish task completion notification", "error", err)
	}
}
