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
	"sync"
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
func NewServer(db *sqlx.DB, logger *zap.SugaredLogger, router *gin.Engine, rabbitCh *amqp.Channel, key string, queueName string) (*SwordChallengeServer, error) {
	s := &SwordChallengeServer{db: db, router: router, logger: logger}

	s.userService = user.NewService(db, logger)

	var not *serverAmqp.Service
	if rabbitCh != nil {
		notS, err := serverAmqp.NewService(rabbitCh, logger, queueName)
		if err != nil {
			return nil, err
		}
		not = notS
	}

	s.notificationService = not
	pub := &serverAmqp.Publisher{RabbitChannel: rabbitCh, Logger: logger, NotificationsQueue: queueName}
	s.tasksService = task.NewService(s.userService, db, pub, logger, key)

	return s, nil
}

func (s *SwordChallengeServer) SetupRoutes() {
	publicAPI := s.router.Group("api/v1")
	publicAPI.Use(gin.Logger())
	privateAPI := publicAPI.Group("")
	privateAPI.Use(s.requireAuthentication)

	publicAPI.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	s.userService.SetupRoutes(publicAPI)
	s.tasksService.SetupRoutes(privateAPI)
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

func (s *SwordChallengeServer) StartWithGracefulShutdown(ctx context.Context, port int) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	wg := &sync.WaitGroup{}
	if s.notificationService != nil {
		go s.notificationService.StartConsumer(ctx, wg)

	}
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

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Errorw("Failed to shut down server", "error", err)
	}

	s.logger.Infow("Server closed successfully")
	wg.Wait()
	s.logger.Infow("Server dependencies closed successfully")
	return nil
}
