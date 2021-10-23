package internal

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/jmoiron/sqlx"
	"log"
	"os"
)

type server struct {
	router *gin.Engine
	db *sqlx.DB
}

func StartServer() {
	r := gin.Default()

	db, err := sqlx.Open("mysql", os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASSWORD") + "@tcp(" + os.Getenv("DB_HOST") + ")/" + os.Getenv("DB_NAME") + "?multiStatements=true")
	if err != nil {
		log.Fatalln(err) //todo
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		log.Fatalln(err) //todo
	}

	driver, err := mysql.WithInstance(db.DB, &mysql.Config{})
	if err != nil {
		log.Fatalln(err)  //todo
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		os.Getenv("DB_NAME"), driver)
	if err != nil {
		log.Fatalln(err)  //todo
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange  {
		log.Fatalln(err)  //todo
	}

	s := server{db: db, router: r}
	s.routes()

	err = s.router.Run()
	if err != nil {
		return //todo
	}
}
