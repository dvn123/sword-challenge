package internal

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	serverAmqp "sword-challenge/internal/amqp"
	"sword-challenge/internal/task"
	"sword-challenge/internal/user"
	"syscall"
	"time"
)

type SwordChallengeServer struct {
	router              *gin.Engine
	server              *http.Server
	db                  *sqlx.DB
	logger              *zap.SugaredLogger
	userService         *user.Service
	tasksService        *task.Service
	notificationService *serverAmqp.Service
}

// NewServer setups the server routes and dependencies, everything is a bit too coupled so we have some funky logic to check whether we're using rabbit or not
func NewServer(db *sqlx.DB, logger *zap.SugaredLogger, router *gin.Engine, rabbitCh *amqp.Channel, key string) (*SwordChallengeServer, error) {
	s := &SwordChallengeServer{db: db, router: router, logger: logger}
	authorizedAPI := router.Group("api/v1")
	authorizedAPI.Use(s.requireAuthentication)

	publicAPI := router.Group("api/v1")

	publicAPI.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	s.userService = user.NewService(publicAPI, db, logger)

	var pub task.Publisher
	pub = &task.LogPublisher{Logger: logger}
	if rabbitCh != nil {
		not, err := serverAmqp.NewService(rabbitCh, logger, "tasks")
		if err != nil {
			return nil, err
		}
		s.notificationService = not
		pub = &serverAmqp.Publisher{RabbitChannel: rabbitCh, Logger: logger, NotificationsQueue: "tasks"}
	}
	s.tasksService = task.NewService(authorizedAPI, s.userService, db, pub, logger, key)

	return s, nil
}

func (s *SwordChallengeServer) RunMigrations() error {
	driver, err := mysql.WithInstance(s.db.DB, &mysql.Config{})
	if err != nil {
		s.logger.Errorw("Failed to start go-migrate driver", "error", err)
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		os.Getenv("DB_NAME"), driver)
	if err != nil {
		s.logger.Errorw("Failed to start go-migrate migration instance", "error", err)
		return err
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		s.logger.Errorw("Failed to migrate database to latest version", "error", err)
		return err
	}
	return nil
}

func (s *SwordChallengeServer) StartWithGracefulShutdown(port int) error {
	s.notificationService.StartConsumer()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	s.server = &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: s.router,
	}

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatalw("Failed to start server", "error", err)
		}
	}()

	<-ctx.Done()

	stop()
	s.logger.Infow("Shutting server down gracefully, press Ctrl+C again to force")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.notificationService.Shutdown()
	if err != nil {
		s.logger.Errorw("Failed to shut down RabbitMQ connection", "error", err)
	}

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Errorw("Failed to shut down server", "error", err)
	}

	s.logger.Infow("Shutting down server")
	return nil
}
