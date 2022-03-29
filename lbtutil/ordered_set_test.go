package lbtutil

import (
	"testing"
	"fmt"
	"math/rand"
	"time"
)

func TestOrderedSet(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	s := NewOrderedSet()
	for _, n := range []int{1,2,3,4,5,7,5,3,1} {
		s.Add(n)
	}
	s.Remove(5)
	s.Remove(3)
	s.Add(5)
	fmt.Printf("OrderedSet: %v\n", s)
	fmt.Printf("OrderedSet random get: %v\n", s.RandomGetOne())
	fmt.Printf("OrderedSet random get: %v\n", s.RandomGetOne())
	fmt.Printf("OrderedSet random get: %v\n", s.RandomGetOne())
}
