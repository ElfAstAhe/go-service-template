package utils

import (
	"sync/atomic"
)

// AtomicString wraps a standard Go string to provide lock-free, thread-safe atomic operations.
// It fully encapsulates low-level pointer manipulations, exposing a standard sync/atomic compliant interface.
type AtomicString struct {
	ptr atomic.Pointer[string]
}

// NewAtomicString allocates and initializes an AtomicString primitive pre-populated with the provided value.
func NewAtomicString(val string) *AtomicString {
	as := &AtomicString{}
	as.Store(val)
	return as
}

// Load atomically retrieves the current underlying string value snapshot safely.
func (as *AtomicString) Load() string {
	p := as.ptr.Load()
	if p == nil {
		return ""
	}
	return *p
}

// Store atomically sets the underlying string value, replacing any previous payload.
func (as *AtomicString) Store(val string) {
	as.ptr.Store(&val)
}

// Swap atomically stores a new string value and returns the previous scalar data sequence snapshot.
func (as *AtomicString) Swap(newVal string) string {
	oldPtr := as.ptr.Swap(&newVal)
	if oldPtr == nil {
		return ""
	}
	return *oldPtr
}

// CompareAndSwap executes a lock-free Compare-And-Swap (CAS) operation for the underlying string.
// It automatically evaluates data equality, swaps pointers if a match occurs, and reports runtime success via bool.
func (as *AtomicString) CompareAndSwap(oldVal, newVal string) bool {
	// Loop guard ensures atomicity in case of concurrent micro-interferences with identical text values
	for {
		currentPtr := as.ptr.Load()
		currentVal := ""
		if currentPtr != nil {
			currentVal = *currentPtr
		}

		// Fail fast if the actual current underlying value doesn't match the expected old value
		if currentVal != oldVal {
			return false
		}

		// Attempt to atomically swap the pointer reference layout
		if as.ptr.CompareAndSwap(currentPtr, &newVal) {
			return true
		}
	}
}
