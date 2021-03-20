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
		s := NewScheduledExecutor(100*time.Millisecond, increaseCount)
		s.Start()
		defer s.Stop()
		time.Sleep(1 * time.Millisecond)
		assert.Equal(t, 0, count)
		time.Sleep(105 * time.Millisecond)
		assert.Equal(t, 1, count)
	})
	t.Run("start with 0 initial delay", func(t *testing.T) {
		var count int
		increaseCount := func() {
			count++
		}
		s := NewScheduledExecutorWithOpts(100*time.Millisecond, increaseCount, WithInitialDelay(0))
		s.Start()
		defer s.Stop()
		time.Sleep(1 * time.Millisecond)
		assert.Equal(t, 1, count)
		time.Sleep(105 * time.Millisecond)
		assert.Equal(t, 2, count)
	})
	t.Run("start with 50 milliseconds of initial delay", func(t *testing.T) {
		var count int
		increaseCount := func() {
			count++
		}
		s := NewScheduledExecutorWithOpts(100*time.Millisecond, increaseCount, WithInitialDelay(50*time.Millisecond))
		s.Start()
		defer s.Stop()
		time.Sleep(1 * time.Millisecond)
		assert.Equal(t, 0, count)
		time.Sleep(55 * time.Millisecond)
		assert.Equal(t, 1, count)
		time.Sleep(105 * time.Millisecond)
		assert.Equal(t, 2, count)
	})
}
