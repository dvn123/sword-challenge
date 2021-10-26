package internal

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestStartAndStopServer(t *testing.T) {
	db, _, _ := sqlmock.New()
	logger, _ := zap.NewDevelopment()

	router := gin.New()

	ctx, cancel := context.WithCancel(context.Background())
	container, rabbitChan := startRabbitTestContainer(ctx)
	t.Cleanup(func() { container.Terminate(ctx) })

	server, _ := NewServer(sqlx.NewDb(db, "mysql"), logger.Sugar(), router, rabbitChan, "6368616e676520746869732070617373")

	go func() {
		err := server.StartWithGracefulShutdown(ctx, 9090)
		if err != nil {
			assert.Nil(t, err)
		}
	}()
	time.Sleep(time.Second)
	cancel()
	time.Sleep(time.Second)
}

func startRabbitTestContainer(ctx context.Context) (testcontainers.Container, *amqp.Channel) {
	req := testcontainers.ContainerRequest{
		Image:        "rabbitmq:3-management-alpine",
		ExposedPorts: []string{"5672/tcp"},
		WaitingFor:   wait.ForLog("Server startup complete"),
	}
	rabbitC, _ := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	endpoint, _ := rabbitC.PortEndpoint(ctx, "5672", "")
	conn, _ := amqp.Dial("amqp://guest:guest@" + endpoint)
	rabbitChan, _ := conn.Channel()
	return rabbitC, rabbitChan
}
