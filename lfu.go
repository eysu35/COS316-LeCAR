
import (
	"container/list"
)

// An LFU is a fixed-size in-memory cache with least-frequently-used eviction
type LFU struct {
}

// NewLFU returns a pointer to a new LFU with a capacity to store limit bytes
func NewLfu(limit int) *LFU {

}

// MaxStorage returns the maximum number of bytes this LFU can store
func (lfu *LFU) MaxStorage() int {
	return lfu.cap
}

// RemainingStorage returns the number of unused bytes available in this LFU
func (lfu *LRU) RemainingStorage() int {
	return lfu.cap - lfu.size
}

// Get returns the value associated with the given key, if it exists.
// This operation counts as a "use" for that key-value pair
// ok is true if a value was found and false otherwise.
func (lfu *LRU) Get(key string) (value []byte, ok bool) {

}

// Remove removes and returns the value associated with the given key, if it exists.
// ok is true if a value was found and false otherwise
func (lfu *LFU) Remove(key string) (value []byte, ok bool) {
	
}

// Set associates the given value with the given key, possibly evicting values
// to make room. Returns true if the binding was added successfully, else false.
func (lfu *LFU) Set(key string, value []byte) bool {
	
}

// Len returns the number of bindings in the LFU.
func (lfu *LFU) Len() int {
	return lfu.keyList.Len()
}

// Stats returns statistics about how many search hits and misses have occurred.
func (lfu *LFU) Stats() *Stats {
	return &lfu.stats
}
