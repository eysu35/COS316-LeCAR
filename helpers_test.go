package main

import (
	"testing"
)

/******************************************************************************/
/*                                 Helpers                                    */
/******************************************************************************/

// Returns true iff a and b represent equal slices of bytes.
func bytesEqual(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	if (a == nil && b != nil) || (a != nil && b == nil) {
		return false
	}

	for i, v := range a {
		if b[i] != v {
			return false
		}
	}

	return true
}

// Fails test t with an error message if fifo.MaxStorage() is not equal to capacity
func checkCapacity(t *testing.T, cache LeCaR, capacity int) {
	max := cache.MaxStorage()
	if max != capacity {
		t.Errorf("Expected cache to have %d MaxStorage, but it had %d", capacity, max)
	}
}
