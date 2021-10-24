package internal

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sword-challenge/internal/tasks"
	"sword-challenge/internal/users"
	"syscall"
	"time"
)

type SwordChallengeServer struct {
	router       *gin.Engine
	server       *http.Server
	db           *sqlx.DB
	logger       *zap.SugaredLogger
	userService  *users.Service
	tasksService *tasks.Service
}

func NewServer(db *sqlx.DB, logger *zap.SugaredLogger, router *gin.Engine) (*SwordChallengeServer, error) {
	s := &SwordChallengeServer{db: db, router: router, logger: logger}
	authorizedAPI := router.Group("api/v1")
	authorizedAPI.Use(s.requireAuthentication)

	publicAPI := router.Group("api/v1")

	publicAPI.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	s.userService = users.NewService(authorizedAPI, publicAPI, db, logger)
	s.tasksService = tasks.NewService(authorizedAPI, s.userService, db, logger)

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
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Errorw("Failed to shut down server", "error", err)
		return err
	}

	s.logger.Infow("Shutting down server")
	return nil
}
