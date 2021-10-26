package task

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestFailWhenKeyIsntHex(t *testing.T) {
	_, err := NewCrypto("asd", zap.NewNop().Sugar())

	assert.NotNil(t, err)
}

func TestFailWhenKeyIsTooShort(t *testing.T) {
	_, err := NewCrypto("1a", zap.NewNop().Sugar())

	assert.NotNil(t, err)
}

func TestEncodeAndDecodeTaskMatch(t *testing.T) {
	key := "6368616e676520746869732070617373776f726420746f206120736563726574"

	l, _ := zap.NewDevelopment()
	text := "olaola"
	dt := &task{ID: 1, Summary: text, CompletedDate: nil, User: nil}
	c, _ := NewCrypto(key, l.Sugar())

	et, _ := c.encryptTask(dt)

	dt2, _ := c.decryptTask(et, 1)
	assert.Equal(t, text, dt2.Summary)
}
