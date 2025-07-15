package collector

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCountsIsThreadSafe(t *testing.T) {
	counter := NewStatusCodeCounter()
	codes := []int{100, 150, 200, 250, 300, 350, 400, 450, 500, 550}
	var wg sync.WaitGroup
	wg.Add(10)
	for i := range 10 {
		go func() {
			defer wg.Done()
			v := codes[i]
			for range 1000 {
				counter.Increment(v)
			}
		}()
	}
	wg.Wait()
	assert.Equal(t, counter.Count(), 10_000)
	s := counter.String()
	assert.Contains(t, s, "[100]: 1000")
	assert.Contains(t, s, "[150]: 1000")
	assert.Contains(t, s, "[200]: 1000")
	assert.Contains(t, s, "[250]: 1000")
	assert.Contains(t, s, "[300]: 1000")
	assert.Contains(t, s, "[350]: 1000")
	assert.Contains(t, s, "[400]: 1000")
	assert.Contains(t, s, "[450]: 1000")
	assert.Contains(t, s, "[500]: 1000")
	assert.Contains(t, s, "[550]: 1000")
}

func TestStringIsCorrect(t *testing.T) {
	counter := NewStatusCodeCounter()
	counter.Increment(100)
	counter.Increment(403)
	counter.Increment(500)
	s := counter.String()
	// go forcefully attempts to prevent consistency in map keys
	assert.Contains(t, s, "Breakdown\n\t")
	assert.Contains(t, s, "[500]: 1")
	assert.Contains(t, s, "[403]: 1")
	assert.Contains(t, s, "[100]: 1")
}
