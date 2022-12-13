package main

import (
	"fmt"
	"testing"
)

var CACHE_SIZE = 20
var LEARNING_RATE = 0.45
var DISCOUNT_RATE = 0.005
var c = NewLeCaR(CACHE_SIZE, LEARNING_RATE, DISCOUNT_RATE)

func TestBasics(t *testing.T) {
	m := c.MaxStorage()

	if m != CACHE_SIZE {
		t.Errorf("incorrect MaxStorage() result: %d", m)
		t.FailNow()
	}

	r := c.RemainingStorage()
	if r != CACHE_SIZE {
		t.Errorf("incorrect RemainingStorage() result: %d", r)
	}
}

func TestBasicSetAndGet(t *testing.T) {
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
	CACHE_SIZE = 20

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

	fmt.Println(c.toString())

}
