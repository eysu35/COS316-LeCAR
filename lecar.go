package main

import (
	"container/heap"
	"container/list"
	"math"
	"math/rand"
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
	clock int               // keeps track of the universal passing of time

	// LFU

	LFUKeyToFreq  map[string]int         // maps keys to frequencies
	LFUFreqToKeys map[int]map[string]int // maps freqs to list of keys
	LFUFreqOrder  IntHeap                // keep a min-ordered list of freqs

	// LRU
	LRU         *list.List               // queue of least recently accessed keys
	LRUPointers map[string]*list.Element // keeps track of location of each key in list

	historyLFU map[string]int // keeps track of history of evictions by LFU, int = time evicted
	historyLRU map[string]int // keeps track of history of evictions by LRU, int = time evicted

	wLRU         float64 // weight of LRU policy
	wLFU         float64 // weight of LFU policy
	learningRate float64 // hyperparameter used for how quickly we update weights
	discountRate float64 // hyperparameter used for determine regret factor

	stats Stats // stats struct
}

// NewLeCaR returns a pointer to a new LeCaR with a capacity to store limit bytes
func NewLeCaR(limit int, learningRate float64, discountRate float64) *LeCaR {
	r := LeCaR{cache: map[string][]byte{},
		cap:           limit,
		size:          0,
		clock:         0,
		LFUKeyToFreq:  map[string]int{},
		LFUFreqToKeys: map[int]map[string]int{},
		LFUFreqOrder:  IntHeap{},
		LRU:           list.New(),
		LRUPointers:   map[string]*list.Element{},
		historyLFU:    map[string]int{},
		historyLRU:    map[string]int{},
		wLRU:          0.5,
		wLFU:          0.5,
		learningRate:  learningRate,                             // initialized via LeCar paper
		discountRate:  math.Pow(discountRate, 1/float64(limit))} // initialized via LeCaR paper

	heap.Init(&r.LFUFreqOrder)
	r.LFUFreqToKeys[1] = make(map[string]int) // add the map for frequency of 1
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

func (lecar *LeCaR) UpdateWeight(key string, time int, policy string) {
	timePassed := lecar.clock - time
	r := math.Pow(lecar.discountRate, float64(timePassed)) // discount rate, by LeCaR paper

	// update relevant weight
	if policy == "LFU" {
		lecar.wLFU = lecar.wLFU * math.Pow(math.E, lecar.learningRate*r) // update by LeCaR paper
	}

	if policy == "LRU" {
		lecar.wLRU = lecar.wLRU * math.Pow(math.E, lecar.learningRate*r) // update by LeCaR paper
	}

	// normalize weights
	lecar.wLFU = lecar.wLFU / (lecar.wLFU + lecar.wLRU)
	lecar.wLRU = 1 - lecar.wLFU

}

// Get returns the value associated with the given key, if it exists.
// This operation counts as a "use" for that key-value pair
// ok is true if a value was found and false otherwise.
func (lecar *LeCaR) Get(key string) (value []byte, ok bool) {
	val, ok := lecar.cache[key]

	// cache miss
	if !ok {
		// check if request is in either history, if so, update policy weights

		//LFU
		if time, ok := lecar.historyLFU[key]; ok {
			delete(lecar.historyLFU, key) // so we don't double update the weights
			lecar.UpdateWeight(key, time, "LFU")
		}

		//LRU
		if time, ok := lecar.historyLFU[key]; ok {
			delete(lecar.historyLRU, key) // so we don't double update the weights
			lecar.UpdateWeight(key, time, "LRU")
		}

		lecar.stats.Misses += 1
	} else {
		// cache hit

		// LFU
		oldFreq := lecar.LFUKeyToFreq[key]   // old freq
		keys := lecar.LFUFreqToKeys[oldFreq] // map of all keys with oldFreq

		delete(keys, key) // remove key from old frequency list

		// increment frequency upon access
		lecar.LFUKeyToFreq[key] = oldFreq + 1

		if lecar.LFUFreqToKeys[oldFreq+1] == nil {
			lecar.LFUFreqToKeys[oldFreq+1] = make(map[string]int)
		}

		lecar.LFUFreqToKeys[oldFreq+1][key] = 1  // add key to new freq list with arbitrary val
		heap.Push(lecar.LFUFreqOrder, oldFreq+1) // add new frequency to heap, noop if it already exists

		// LRU
		ptr := lecar.LRUPointers[key]
		lecar.LRU.MoveToFront(ptr)

		lecar.stats.Hits += 1
	}

	return val, ok
}

// Remove removes and returns the value associated with the given key, if it exists.
// ok is true if a value was found and false otherwise
func (lecar *LeCaR) Remove(key string, policy string) (value []byte, ok bool) {
	val, ok := lecar.cache[key]

	// if key not found, just return nil, false
	if !ok {
		return val, ok
	}

	// if key exists, evict the key based on the chosen policy
	// LFU
	if policy == "LFU" {
		delete(lecar.cache, key) // remove from the cache

		// remove from LFU
		freq := lecar.LFUKeyToFreq[key] // store the frequency of the key to be deleted
		delete(lecar.LFUKeyToFreq, key) // remove from LFU map

		// remove the key from the list of keys corresponding to one frequency
		delete(lecar.LFUFreqToKeys[freq], key)

		// add to historyLFU
		lecar.historyLFU[key] = lecar.clock // value = current time

		// decrease size by size of deletion
		deletionSize := len(val) + len(key)
		lecar.size = lecar.size - deletionSize

	} else if policy == "LRU" { // LRU
		delete(lecar.cache, key) // remove from the cache

		// remove from the key list by searching for pointer
		ptr := lecar.LRUPointers[key]
		_ = lecar.LRU.Remove(ptr)
		delete(lecar.LRUPointers, key) //remove from pointers map

		// add to historyLRU
		lecar.historyLRU[key] = lecar.clock // value = current time

		// decrease size by size of deletion
		deletionSize := len(val) + len(key)
		lecar.size = lecar.size - deletionSize

	}

	// increment the clock to keep track of number of evictions
	lecar.clock = lecar.clock + 1

	return val, true
}

// Set associates the given value with the given key, possibly evicting values
// to make room. Returns true if the binding was added successfully, else false.
func (lecar *LeCaR) Set(key string, value []byte) bool {
	// check if sufficient storage is available
	setSize := len(key) + len(value)
	if setSize > lecar.cap {
		return false
	}

	// check if the key already exists
	_, ok := lecar.cache[key]
	if ok {
		// if key exists, remove it from the cache since we want to update value
		sample := rand.Float64() // returns a float in [0.0. 1.0)
		// let random sample determine policy based on which weight interval it falls in
		policy := ""
		if sample <= lecar.wLFU {
			policy = "LFU"
		} else {
			policy = "LFU"
		}
		lecar.Remove(key, policy)
	}

	// remove elements until there is enough space for the new key value pair
	if lecar.RemainingStorage() < setSize {

		for lecar.RemainingStorage() < setSize {
			// again, sample eviction policy from weights and evict accordingly
			sample := rand.Float64() // returns a float in [0.0. 1.0)
			// let random sample determine policy based on which weight interval it falls in
			policy := ""
			if sample <= lecar.wLFU {
				policy = "LFU"
			} else {
				policy = "LRU"
			}

			if policy == "LFU" {
				// find the lowest freq for which there is a key
				minFreq := lecar.LFUFreqOrder.Pop().(int)
				for len(lecar.LFUFreqToKeys[minFreq]) == 0 {
					minFreq = lecar.LFUFreqOrder.Pop().(int)
				}

				// if more than one item has this frequency, put the freq back in the heap
				if len(lecar.LFUFreqToKeys[minFreq]) > 1 {
					heap.Push(lecar.LFUFreqOrder, minFreq)
				}

				// get an arbitrary key with the desired frequency
				for keyRemove, _ := range lecar.LFUFreqToKeys[minFreq] {
					_, _ = lecar.Remove(keyRemove, policy)
					break
				}
			}

			if policy == "LRU" {
				keyTemp := lecar.LRU.Back().Value
				keyRemove := keyTemp.(string)
				_, _ = lecar.Remove(keyRemove, policy)
			}
		}
	}

	// Update cache, LFU cache, and LRU cache
	lecar.cache[key] = value
	lecar.size += setSize

	// LFU
	lecar.LFUKeyToFreq[key] = 1
	lecar.LFUFreqToKeys[1][key] = 1
	lecar.LFUFreqOrder.Push(1)

	//LRU
	ptr := &list.Element{} // blank initialization
	if lecar.LRU.Len() == 0 {
		ptr = lecar.LRU.PushFront(key)
	} else {
		ptr = lecar.LRU.InsertBefore(key, lecar.LRU.Front())
	}
	lecar.LRUPointers[key] = ptr

	return true
}

// Len returns the number of bindings in the LeCaR.
func (lecar *LeCaR) Len() int {
	return len(lecar.cache)
}

// Stats returns statistics about how many search hits and misses have occurred.
func (lecar *LeCaR) Stats() *Stats {
	return &lecar.stats
}
