package utils

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAtomicString_NewAndLoadStore(t *testing.T) {
	as := NewAtomicString("initial")
	assert.Equal(t, "initial", as.Load())

	as.Store("updated")
	assert.Equal(t, "updated", as.Load())
}

func TestAtomicString_NilColdStart(t *testing.T) {
	var as AtomicString
	assert.Equal(t, "", as.Load())

	as.Store("ignite")
	assert.Equal(t, "ignite", as.Load())
}

func TestAtomicString_Swap(t *testing.T) {
	as := NewAtomicString("old-value")
	old := as.Swap("new-value")

	assert.Equal(t, "old-value", old)
	assert.Equal(t, "new-value", as.Load())
}

func TestAtomicString_NilSwap(t *testing.T) {
	var as AtomicString
	old := as.Swap("ignite")

	assert.Equal(t, "", old)
	assert.Equal(t, "ignite", as.Load())
}

func TestAtomicString_CompareAndSwap(t *testing.T) {
	as := NewAtomicString("state-1")

	success := as.CompareAndSwap("wrong-state", "state-2")
	assert.False(t, success)
	assert.Equal(t, "state-1", as.Load())

	success = as.CompareAndSwap("state-1", "state-2")
	assert.True(t, success)
	assert.Equal(t, "state-2", as.Load())
}

func TestAtomicString_NilCompareAndSwap(t *testing.T) {
	var as AtomicString

	success := as.CompareAndSwap("wrong", "ignite")
	assert.False(t, success)
	assert.Equal(t, "", as.Load())

	success = as.CompareAndSwap("", "ignite")
	assert.True(t, success)
	assert.Equal(t, "ignite", as.Load())
}

func TestAtomicString_ConcurrentDataRace(t *testing.T) {
	as := NewAtomicString("baseline")
	var wg sync.WaitGroup

	workers := 50
	iterations := 100
	wg.Add(workers * 2)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				as.Store("writer-payload")
			}
		}()
	}

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = as.Load()
			}
		}()
	}

	wg.Wait()
	assert.Contains(t, []string{"baseline", "writer-payload"}, as.Load())
}
