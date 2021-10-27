package internal

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"sword-challenge/internal/task"
	"sword-challenge/internal/user"
	"testing"
	"time"
)

func TestStartAndStopServer(t *testing.T) {
	db, _, _ := sqlmock.New()
	logger, _ := zap.NewDevelopment()

	router := gin.New()
	sqlxDB := sqlx.NewDb(db, "mysql")

	pub := &task.LogPublisher{Logger: logger.Sugar()}
	userService := user.NewService(sqlxDB, logger.Sugar())

	tasksService := task.NewService(userService, sqlxDB, pub, logger.Sugar(), "6368616e676520746869732070617373")

	server := &SwordChallengeServer{
		router:       router,
		server:       nil,
		db:           sqlxDB,
		logger:       logger.Sugar(),
		userService:  userService,
		tasksService: tasksService,
	}
	server.SetupRoutes()
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		err := server.StartWithGracefulShutdown(ctx, 9090)
		if err != nil {
			assert.Nil(t, err)
		}
	}()
	time.Sleep(100 * time.Millisecond)
	cancel()
}
