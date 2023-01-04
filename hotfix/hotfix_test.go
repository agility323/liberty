package hotfix

import (
	"testing"
	"fmt"
)

//go:noinline
func F() string {
	return "F"
}

//go:noinline
func NewF() string {
	return "NewF"
}

type S struct {
}

//go:noinline
func (s *S) F() string {
	return "s.F"
}

//go:noinline
func NewSF(s *S) string {
	return "NewSF"
}

func TestHotfix(t *testing.T) {
	entries := []interface{} {
		NewFuncEntry(F, NewF),
		NewMethodEntry((*S)(nil), "F", NewSF),
	}
	fmt.Printf("old F: %v\n", F())
	fmt.Printf("old s.F: %v\n", (*S)(nil).F())
	ApplyHotfix(entries)
	fmt.Printf("new F: %v\n", F())
	fmt.Printf("new s.F: %v\n", (*S)(nil).F())
}
