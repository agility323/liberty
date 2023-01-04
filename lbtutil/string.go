/*
By Thomas Wade, 2021.12.14
*/
package lbtutil

import (
	"unicode"
	"hash/fnv"
)

var simpleSymbolMap map[rune]struct{} = map[rune]struct{} {
	'-': struct{}{},
	'_': struct{}{},
}

func IsSimpleString(s string) bool {
	rs := []rune(s)
	for _, r := range rs {
		if unicode.IsLetter(r) || unicode.IsDigit(r) { continue }
		if _, ok := simpleSymbolMap[r]; ok { continue }
		return false
	}
	return true
}

func StringHash(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32())
}
