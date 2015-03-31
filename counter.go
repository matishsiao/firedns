package main

import (
	"sync"
)

type Counter struct {
	mu *sync.Mutex
	Total_counter int64
	Cache_counter int64
	Misscache_counter int64
}

func(c *Counter) Aggregate() ([]int64) {
	c.mu.Lock()
	total := c.Total_counter
	cache := c.Cache_counter
	miss := c.Misscache_counter
	agg := []int64{total,cache,miss}
	c.Total_counter = 0
	c.Cache_counter = 0
	c.Misscache_counter = 0
	c.mu.Unlock()
	return agg
}