package lbtutil

import (
	"math/rand"
)

// thread unsafe
type OrderedSet struct {
	data []interface{}
	m map[interface{}]int
}

func NewOrderedSet() *OrderedSet {
	s := &OrderedSet{
		data: make([]interface{}, 0),
		m: make(map[interface{}]int),
	}
	return s
}

func (s *OrderedSet) Add(v interface{}) {
	if _, ok := s.m[v]; ok { return }
	s.data = append(s.data, v)
	s.m[v] = len(s.data) - 1
}

func (s *OrderedSet) Remove(v interface{}) {
	i, ok := s.m[v]	// TODO: s(nil) bug, panic: runtime error: invalid memory address or nil pointer dereference
	if !ok { return }
	delete(s.m, v)
	s.data[i] = s.data[len(s.data) - 1]
	s.m[s.data[i]] = i
	s.data = s.data[:len(s.data) - 1]
}

func (s *OrderedSet) RandomGetOne() interface{} {
	if len(s.data) == 0 { return nil }
	return s.data[rand.Intn(len(s.data))]
}

func (s *OrderedSet) GetAll() []interface{} {
	data := make([]interface{}, len(s.data))
	copy(data, s.data)
	return data
}

func (s *OrderedSet) Size() int {
	return len(s.data)
}

func (s *OrderedSet) Clear() {
	s.data = make([]interface{}, 0)
	s.m = make(map[interface{}]int)
}
