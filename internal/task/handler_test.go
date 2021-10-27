package task

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestShouldInitializeService(t *testing.T) {
	c := gin.New()
	group := c.Group("")
	NewService(group, nil, nil, nil, nil, "6368616e676520746869732070617373")
	assert.Equal(t, 4, len(c.Routes()))
}
