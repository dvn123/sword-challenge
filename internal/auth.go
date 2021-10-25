package internal

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sword-challenge/internal/util"
)

func (s *SwordChallengeServer) requireAuthentication(c *gin.Context) {
	token, err := c.Cookie(util.AuthCookie)
	if err != nil {
		token = c.GetHeader(util.AuthHeader)
	}
	if token == "" {
		c.Status(http.StatusUnauthorized)
		c.Abort()
		return
	}
	user, err := s.userService.GetUserByToken(token)
	if err != nil {
		s.logger.Warnw("Failed to get user from token", "error", err)
		c.Status(http.StatusUnauthorized)
		c.Abort()
		return
	}

	c.Set(util.UserContextKey, user)
}
