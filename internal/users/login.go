package users

import (
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"time"
)

func (userService *Service) loginUser(c *gin.Context) {
	user := &User{}

	if err := c.BindJSON(user); err != nil {
		userService.Logger.Errorf("Failed to parse user request body: %v", err)
		return
	}

	token, err := userService.authenticateUser(user.ID)
	if err != nil || token == "" {
		userService.Logger.Errorf("Failed to add user to storage: %v", err)
		c.JSON(http.StatusInternalServerError, nil) //todo error object
		return
	}

	c.SetCookie("auth-token", token, int(time.Hour), "/", "localhost", true, true)

	c.Status(http.StatusOK)
}
