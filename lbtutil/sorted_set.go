package lbtutil

import (
	"math/rand"
	"sync"
)

type SortedSetStr struct {
	arr []string
	m map[string]int
	hsize uint64
	lock sync.RWMutex
}

func NewSortedSetStr(hsize uint64) *SortedSetStr {
	s := &SortedSetStr{
		arr: make([]string, 0),
		m: make(map[string]int),
		hsize: hsize,
	}
	return s
}

func (s *SortedSetStr) Add(v string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.m[v]; ok { return }

	// ascending order
	tail := len(s.arr)
	s.arr = append(s.arr, v)
	s.m[v] = tail
	for i := tail; i > 0; i-- {
		if s.arr[i - 1] < s.arr[i] {
			break
		}
		s.arr[i - 1], s.arr[i] = s.arr[i], s.arr[i - 1]
		s.m[s.arr[i - 1]] = i - 1
		s.m[s.arr[i]] = i
	}
}

func (s *SortedSetStr) Remove(v string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	idx, ok := s.m[v]
	if !ok { return }

	for i := idx; i < len(s.arr) - 1 ; i++ {
		s.arr[i] = s.arr[i + 1]
		s.m[s.arr[i]] = i
	}
	s.arr = s.arr[:len(s.arr) - 1]
	delete(s.m, v)
}

func (s *SortedSetStr) RandomGet() string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if len(s.arr) == 0 { return "" }
	return s.arr[rand.Intn(len(s.arr))]
}

func (s *SortedSetStr) GetAll() []string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	arr := make([]string, len(s.arr))
	if len(s.arr) > 0 {
		copy(arr, s.arr)
	}
	return arr
}

func (s *SortedSetStr) HashGet(hval uint64) string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if len(s.arr) == 0 { return "" }
	hval = hval % s.hsize
	idx := int(hval * uint64(len(s.arr)) / s.hsize) % len(s.arr)
	return s.arr[idx]
}
