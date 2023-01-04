package lbtutil

import (
	"testing"
)

var idToHash []ObjectID
var strToHash []string

func init() {
	n := 127
	idToHash = make([]ObjectID, n, n)
	strToHash = make([]string, n, n)
	for i := 0; i < n; i++ {
		id := NewObjectID()
		idToHash[i] = id
		strToHash[i] = id.Hex()[:16]
	}
}

/*
cpu: Intel(R) Xeon(R) Silver 4216 CPU @ 2.10GHz
BenchmarkHashOfString-8         33847371                35.04 ns/op
BenchmarkHashOfObjectID-8       106751695               11.18 ns/op
*/
func BenchmarkHashOfString(b *testing.B) {
	n := len(strToHash)
	for i := 0; i < b.N; i++ {
		_ = StringHash(strToHash[i % n])
	}
}

func BenchmarkHashOfObjectID(b *testing.B) {
	n := len(idToHash)
	m := int32(n)
	for i := 0; i < b.N; i++ {
		_ = ObjectIDCounter(idToHash[i % n]) % m
	}
}
