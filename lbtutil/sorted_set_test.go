package lbtutil

import (
	"testing"
	"fmt"
	"math/rand"
	"time"
)

func TestSortedSetStr(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	hsize := uint64((1 << 16) - 1)
	s := NewSortedSetStr(hsize)
	for _, n := range []string{"a","b","c","d","e","g","e","c","a"} {
		s.Add(n)
	}
	s.Remove("e")
	s.Remove("c")
	s.Add("e")
	fmt.Printf("SortedSetStr: %v\n", s)
	fmt.Printf("SortedSetStr random get: %v\n", s.RandomGet())
	fmt.Printf("SortedSetStr random get: %v\n", s.RandomGet())
	fmt.Printf("SortedSetStr random get: %v\n", s.RandomGet())
	fmt.Printf("SortedSetStr hash get: %v\n", s.HashGet(0))
	fmt.Printf("SortedSetStr hash get: %v\n", s.HashGet(13106))
	fmt.Printf("SortedSetStr hash get: %v\n", s.HashGet(13107))
	fmt.Printf("SortedSetStr hash get: %v\n", s.HashGet(52427))
	fmt.Printf("SortedSetStr hash get: %v\n", s.HashGet(52428))
	fmt.Printf("SortedSetStr size: %v\n", s.Size())
}
