package main

import (
	"testing"
)

var CACHE_SIZE = 64000
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
