package amqp

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"sword-challenge/internal/task"
	"testing"
)

func TestFailsToPublishMessage(t *testing.T) {
	pu := Publisher{Logger: zap.NewNop().Sugar(), RabbitChannel: &amqp.Channel{}, NotificationsQueue: " "}
	assert.Panics(t, func() {
		_ = pu.PublishTask(task.Notification{})
	})
}
