package main

import (
	"container/list"
)

// An LRU is a fixed-size in-memory cache with least-recently-used eviction

type LRU struct {
	cache       map[string][]byte        // map of string keys to slice values
	cap         int                      // the max capacity of the cache
	keyList     *list.List               // keeps track of lru order
	keyPointers map[string]*list.Element // keeps track of location of each key in list
	size        int                      // number of bytes currently in the cache
	stats       Stats                    // stats struct
}

// NewLRU returns a pointer to a new LRU with a capacity to store limit bytes
func NewLru(limit int) *LRU {
	r := LRU{cap: limit, cache: map[string][]byte{},
		keyList: list.New(), size: 0, stats: Stats{Hits: 0, Misses: 0},
		keyPointers: map[string]*list.Element{}}
	return &r
}

// MaxStorage returns the maximum number of bytes this LRU can store
func (lru *LRU) MaxStorage() int {
	return lru.cap
}

// RemainingStorage returns the number of unused bytes available in this LRU
func (lru *LRU) RemainingStorage() int {
	return lru.cap - lru.size
}

// Get returns the value associated with the given key, if it exists.
// This operation counts as a "use" for that key-value pair
// ok is true if a value was found and false otherwise.
func (lru *LRU) Get(key string) (value []byte, ok bool) {
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
func (lru *LRU) Remove(key string) (value []byte, ok bool) {
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
func (lru *LRU) Set(key string, value []byte) bool {
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

// Len returns the number of bindings in the LRU.
func (lru *LRU) Len() int {
	return lru.keyList.Len()
}

// Stats returns statistics about how many search hits and misses have occurred.
func (lru *LRU) Stats() *Stats {
	return &lru.stats
}
