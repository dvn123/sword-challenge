package main

import (
	"context"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"log"
	"os"
	"sword-challenge/internal"
)

func main() {
	db, err := setupDatabase()
	defer func(db *sqlx.DB) {
		if err := db.Close(); err != nil {
			log.Printf("Failed to close DB connection. error: %v", err)
		}
	}(db)

	conn, ch := setupRabbit()
	defer func(conn *amqp.Connection) {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close RabbitMQ connection. error: %v", err)
		}
	}(conn)
	defer ch.Close()

	ginEngine := setupGin()

	logger := setupLogger()
	defer func(Logger *zap.SugaredLogger) {
		if err := Logger.Sync(); err != nil {
			log.Printf("Failed to sync zap logger. error: %v", err)
		}
	}(logger)

	s, err := internal.NewServer(db, logger, ginEngine, ch, os.Getenv("AES_KEY"))
	if err != nil {
		log.Fatalf("Failed to create server. error: %v", err)
	}

	// TODO This should be configurable
	err = s.RunMigrations()
	if err != nil {
		log.Fatalf("Failed to run migration on the database. err: %v", err)
	}
	// TODO This should be configurable
	err = s.StartWithGracefulShutdown(context.Background(), 8080)
	if err != nil {
		log.Fatalf("Failed to start server. error: %v", err)
	}
}

func setupGin() *gin.Engine {
	ginEngine := gin.New()
	ginEngine.Use(gin.Recovery())
	ginMode := os.Getenv("GIN_MODE")
	if ginMode != "" {
		gin.SetMode(ginMode)
	}
	return ginEngine
}

func setupRabbit() (*amqp.Connection, *amqp.Channel) {
	conn, err := amqp.Dial(os.Getenv("RABBIT_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ. error: %v", err)
	}

	c, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open RabbitMQ channel. error: %v", err)
	}

	return conn, c
}

func setupDatabase() (*sqlx.DB, error) {
	// multiStatements=true is bad (increases SQLi possibilities) but I need it here because we're migrating the database to the latest version from the server itself (server.go:36)
	// This is good for development but would be removed if this ever went to prod and the MySQL DB wasn't always running in Docker container
	db, err := sqlx.Open("mysql", os.Getenv("DB_USER")+":"+os.Getenv("DB_PASSWORD")+"@tcp("+os.Getenv("DB_HOST")+")/"+os.Getenv("DB_NAME")+"?multiStatements=true&parseTime=true")
	if err != nil {
		log.Fatalf("Failed to connect to the database. error: %v", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to validate DB connection. error: %v", err)
	}
	return db, nil
}

func setupLogger() *zap.SugaredLogger {
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.DisableStacktrace = true
	simpleLogger, _ := loggerConfig.Build()
	logger := simpleLogger.Sugar()
	return logger
}
