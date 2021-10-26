package amqp

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestFailsToCreateService(t *testing.T) {
	assert.Panics(t, func() {
		_, _ = NewService(&amqp.Channel{}, zap.NewNop().Sugar(), " ")
	})
}
