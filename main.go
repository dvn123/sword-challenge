package main

import (
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"log"
	"os"
	"sword-challenge/internal"
)

// In here we would read configurations (from file, env...) and set things to dev or prd mode. Ex: gin would use release mode in prod
func main() {
	ginEngine := gin.Default()

	// multiStatements=true is bad (increases SQLi possibilities) but I need it here because we're migrating the database to the latest version from the server itself (server.go:36)
	// This is good for development but would be removed if this ever went to prod and the MySQL DB wasn't always running in Docker container
	db, err := sqlx.Open("mysql", os.Getenv("DB_USER")+":"+os.Getenv("DB_PASSWORD")+"@tcp("+os.Getenv("DB_HOST")+")/"+os.Getenv("DB_NAME")+"?multiStatements=true&parseTime=true")
	if err != nil {
		log.Fatalf("Failed to connect to the database. err: %v", err)
	}
	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to validate DB connection. err: %v", err)
	}

	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.DisableStacktrace = true
	simpleLogger, _ := loggerConfig.Build()
	logger := simpleLogger.Sugar()
	defer func(Logger *zap.SugaredLogger) {
		err := Logger.Sync()
		if err != nil {
			log.Printf("Failed to sync zap logger. error: %v", err)
		}
	}(logger)

	s, err := internal.NewServer(db, logger, ginEngine)
	if err != nil {
		log.Fatalf("Failed to create server. err: %v", err)
	}
	err = s.RunMigrations()
	if err != nil {
		log.Fatalf("Failed to run migration on the database. err: %v", err)
	}
	err = s.StartWithGracefulShutdown(8080)
	if err != nil {
		log.Fatalf("Failed to start server. err: %v", err)
	}
}
