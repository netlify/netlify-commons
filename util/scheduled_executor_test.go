package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestScheduledExecutor(t *testing.T) {
	t.Run("default ticker starts with delay", func(t *testing.T) {
		var count int
		increaseCount := func() {
			count++
		}
		s := NewScheduledExecutor(3*time.Second, increaseCount)
		s.Start()
		assert.Equal(t, 0, count)
		time.Sleep(3 * time.Second)
		assert.Equal(t, 1, count)
		s.Stop()
	})
	t.Run("default ticker starts with delay", func(t *testing.T) {
		var count int
		increaseCount := func() {
			count++
		}
		s := NewScheduledExecutor(3*time.Second, increaseCount)
		s.Start()
		assert.Equal(t, 0, count)
		time.Sleep(3 * time.Second)
		assert.Equal(t, 1, count)
		s.Stop()
	})
}
