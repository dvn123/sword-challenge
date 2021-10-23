package users

import (
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"log"
	"net/http"
)

type role struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type user struct {
	ID       int64 `json:"id"`
	Role     role
	Username string `json:"username"`
}

func Routes(router *gin.RouterGroup, db *sqlx.DB) {
	router.GET("/users", getUser(db))
	router.POST("/users", postUser(db))

}

func getUser(db *sqlx.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		users, err := getUserFromStore(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, "{}") //todo error object
			return
		}

		c.JSON(http.StatusOK, users)
	}

}

func postUser(db *sqlx.DB) func(c *gin.Context) {
	return func(c *gin.Context) {

		var user *user

		if err := c.BindJSON(user); err != nil {
			log.Println(err)
			return //todo
		}

		user, err := addUserToStore(db, user)
		if err != nil {
			log.Println(err)                             //todo
			c.JSON(http.StatusInternalServerError, "{}") //todo error object
		}
		c.JSON(http.StatusCreated, user)
	}
}
