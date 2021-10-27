package task

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestShouldInitializeService(t *testing.T) {
	service := NewService(nil, nil, nil, nil, "6368616e676520746869732070617373")
	assert.NotNil(t, service)
}

func TestShouldInitializeRoutes(t *testing.T) {
	c := gin.New()
	group := c.Group("")
	service := NewService(nil, nil, nil, nil, "6368616e676520746869732070617373")
	service.SetupRoutes(group)
	assert.NotNil(t, service)
	assert.Equal(t, 4, len(c.Routes()))
}
