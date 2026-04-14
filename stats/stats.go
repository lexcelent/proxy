package stats

import "sync/atomic"

type Counter struct {
	counter atomic.Int64
}

func New() *Counter {
	return &Counter{counter: atomic.Int64{}}
}

// Add just adds bytes to counter
func (c *Counter) Add(delta int64) {
	c.counter.Add(delta)
}

// Total returns counter value
func (c *Counter) Total() int64 {
	return c.counter.Load()
}
