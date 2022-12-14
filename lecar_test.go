package main

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var LEARNING_RATE = 0.45
var DISCOUNT_RATE = 0.005

func TestBasics(t *testing.T) {
	cache_size := 1000
	c := NewLeCaR(cache_size, LEARNING_RATE, DISCOUNT_RATE, 0.5)
	m := c.MaxStorage()

	if m != cache_size {
		t.Errorf("incorrect MaxStorage() result: %d", m)
		t.FailNow()
	}

	r := c.RemainingStorage()
	if r != cache_size {
		t.Errorf("incorrect RemainingStorage() result: %d", r)
	}
}

func TestBasicSetAndGet(t *testing.T) {
	cache_size := 1000
	c := NewLeCaR(cache_size, LEARNING_RATE, DISCOUNT_RATE, 0.5)

	for i := 0; i < 4; i++ {
		key := fmt.Sprintf("key%d", i)
		val := []byte(key)
		ok := c.Set(key, val)
		if !ok {
			t.Errorf("Failed to add binding with key: %s", key)
			t.FailNow()
		}

		res, _ := c.Get(key)
		if !bytesEqual(res, val) {
			t.Errorf("Wrong value %s for binding with key: %s", res, key)
			t.FailNow()
		}
	}
}

func TestBasicEviction(t *testing.T) {
	cache_size := 20
	c := NewLeCaR(cache_size, LEARNING_RATE, DISCOUNT_RATE, 0.5)

	for i := 0; i < 2; i++ {
		key := fmt.Sprintf("!key%d", i)
		val := []byte(key)
		ok := c.Set(key, val)
		if !ok {
			t.Errorf("Failed to add binding with key: %s", key)
			t.FailNow()
		}
	}

	// arbitrarily increase the freq of one entry
	for i := 0; i < 10; i++ {
		c.Get("!key0")
	}

	size := c.RemainingStorage()
	if size != 0 {
		t.Errorf("Incorrect remaining storage, expected 0 but got %d", size)
		t.FailNow()
	}

	// try to add something else
	key := "!key2"
	val := []byte(key)
	ok := c.Set(key, val)

	if !ok {
		t.Errorf("Insert failure")
		t.FailNow()
	}

	// make sure key1 got evicted
	_, ok = c.Get("!key1")
	if ok {
		t.Errorf("Cache should not contain key1")
		t.FailNow()
	}

	fmt.Println(c.CacheToString())
	fmt.Println(c.WeightsToString())
	fmt.Println(c.HistoryToString())

	_, ok = c.Get("!key1")
	if ok {
		t.Errorf("Cache should not contain key1")
		t.FailNow()
	}

	fmt.Println(c.WeightsToString())
	fmt.Println(c.stats.toString())
}

func TestReweighting1(t *testing.T) {
	numEntries := 2
	cache_size := 8 * numEntries
	c := NewLeCaR(cache_size, LEARNING_RATE, DISCOUNT_RATE, 0.5)

	for i := 0; i < numEntries; i++ {
		key := fmt.Sprintf("key%d", i)
		val := []byte(key)
		ok := c.Set(key, val)
		if !ok {
			t.Errorf("Failed to add binding with key: %s", key)
			t.FailNow()
		}
	}

	// for i := 0; i < 5; i++ {
	// 	// make key0 the most frequently
	// 	for i := 0; i < 10; i++{
	// 		c.Get("key0")
	// 	}

	// 	// make key1 the most recently used
	// 	c.Get("key1")

	// 	c.Set("key2", "key2")
	// 	fmt.Println(c.WeightsToString)
	// 	}

	for i := 0; i < 10; i++ {
		c.Get("key0")
	}

	// make key1 the most recently used
	c.Get("key1")

	c.Set("key2", []byte("key2"))
	fmt.Println(c.HistoryToString())

	c.Get("key1")
	c.Get("key0")
	fmt.Println(c.WeightsToString())

}

func TestReweighting2(t *testing.T) {
	cache_size := 1000
	c := NewLeCaR(cache_size, LEARNING_RATE, DISCOUNT_RATE, 0.5)

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%d", i)
		val := []byte(key)
		ok := c.Set(key, val)
		if !ok {
			t.Errorf("Failed to add binding with key: %s", key)
			t.FailNow()
		}
	}

	for i := 0; i < 1000; i++ {
		// with 80% prob, get a key 0-99, with 20% insert a new key
		rand.Seed(time.Now().UnixNano())
		p := rand.Float64()

		if p < 0.60 {
			rand.Seed(time.Now().UnixNano())
			d := rand.Intn(100)
			fmt.Println(d)

			key := fmt.Sprintf("key%d", d)
			_, ok := c.Get(key)
			if ok {
				fmt.Println("hit!")
			}
		} else {
			e := rand.Intn(100) + 101
			key := fmt.Sprintf("key%d", e)
			val := []byte(key)
			ok := c.Set(key, val)
			if !ok {
				t.Errorf("Failed to add binding with key: %s", key)
				t.FailNow()
			}
		}
	}

	fmt.Println(c.CacheToString())
	fmt.Println(c.WeightsToString())
	fmt.Println(c.stats.toString())
	// fmt.Println(c.HistoryToString())
}
