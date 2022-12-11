package main

import (
	"container/heap"
	"container/list"
	"math"
)

// LeCaR is a fixed-size in-memory cache that uses Learning Cache Replacement
// to determine the optimal eviction policy

type Stats struct {
	Hits   int
	Misses int
}

// type IntHeap []int

type LeCaR struct {
	cache map[string][]byte // map of string keys to slice values
	cap   int               // the max capacity of the cache
	size  int               // number of bytes currently in the cache

	// LFU
	LFU      map[int]*list.List // maps freqs to list of keys
	LFUFreqs IntHeap            // keep a min-ordered list of freqs

	// LRU
	LRU         *list.List               // queue of least recently accessed keys
	LRUPointers map[string]*list.Element // keeps track of location of each key in list

	historyLRU map[string]int // keeps track of history of evictions by LRU
	historyLFU map[string]int // keeps track of history of evictions by LFU

	wLRU         float64 // weight of LRU policy
	wLFU         float64 // weight of LFU policy
	learningRate float64 // hyperparameter used for how quickly we update weights
	discountRate float64 // hyperparameter used for determine regret factor

	stats Stats // stats struct
}

// NewLeCaR returns a pointer to a new LeCaR with a capacity to store limit bytes
func NewLeCaR(limit int, learningRate float64, discountRate float64) *LeCaR {
	r := LeCaR{cache: map[string][]byte{},
		cap:          limit,
		size:         0,
		LFU:          map[int]*list.List{},
		LFUFreqs:     IntHeap{},
		LRU:          list.New(),
		LRUPointers:  map[string]*list.Element{},
		wLRU:         0.5,
		wLFU:         0.5,
		learningRate: learningRate,                             // initialized via LeCar paper
		discountRate: math.Pow(discountRate, 1/float64(limit))} // initialized via LeCaR paper

	heap.Init(&r.LFUFreqs)
	return &r
}

// MaxStorage returns the maximum number of bytes this LeCaR can store
func (lecar *LeCaR) MaxStorage() int {
	return lecar.cap
}

// RemainingStorage returns the number of unused bytes available in this LeCaR
func (lecar *LeCaR) RemainingStorage() int {
	return lecar.cap - lecar.size
}

// Get returns the value associated with the given key, if it exists.
// This operation counts as a "use" for that key-value pair
// ok is true if a value was found and false otherwise.
func (lecar *LeCaR) Get(key string) (value []byte, ok bool) {
	val, ok := lru.cache[key]

	if !ok {
		lru.stats.Misses += 1
	} else {
		ptr := lru.keyPointers[key]
		lru.keyList.MoveToFront(ptr)
		lru.stats.Hits += 1
	}

	return val, ok
}

// Remove removes and returns the value associated with the given key, if it exists.
// ok is true if a value was found and false otherwise
func (lecar *LeCar) Remove(key string) (value []byte, ok bool) {
	val, ok := lru.cache[key]

	if !ok {
		return val, ok
	}

	delete(lru.cache, key) // remove from the map

	// remove from the key list by searching for pointer
	ptr := lru.keyPointers[key]
	_ = lru.keyList.Remove(ptr)

	delete(lru.keyPointers, key) //remove from pointers map

	// decrease size by size of deletion
	deletionSize := len(val) + len(key)
	lru.size = lru.size - deletionSize

	return val, true
}

// Set associates the given value with the given key, possibly evicting values
// to make room. Returns true if the binding was added successfully, else false.
func (lecar *LeCaR) Set(key string, value []byte) bool {
	// check if sufficient storage is available
	setSize := len(key) + len(value)
	if setSize > lru.cap {
		return false
	}

	// check if the key already exists
	_, ok := lru.cache[key]
	if ok {
		lru.Remove(key)
	}

	// remove elements until there is enough space for the new key value pair
	if lru.RemainingStorage() < setSize {

		for lru.RemainingStorage() < setSize {
			keyTemp := lru.keyList.Back().Value
			keyRemove := keyTemp.(string)
			_, _ = lru.Remove(keyRemove)
		}
	}

	// Update cache, keylist, and size
	lru.cache[key] = value
	ptr := &list.Element{} // blank initialization
	if lru.keyList.Len() == 0 {
		ptr = lru.keyList.PushFront(key)
	} else {
		ptr = lru.keyList.InsertBefore(key, lru.keyList.Front())
	}
	lru.keyPointers[key] = ptr
	lru.size += setSize

	return true
}

// Len returns the number of bindings in the LeCaR.
func (lecar *LeCaR) Len() int {
	return lru.keyList.Len()
}

// Stats returns statistics about how many search hits and misses have occurred.
func (lecar *LeCaR) Stats() *Stats {
	return &lru.stats
}
