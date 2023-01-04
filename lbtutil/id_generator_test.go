package lbtutil

import (
	"testing"
	"fmt"
)

func TestNewObjectID(t *testing.T) {
	n := 5
	for i := 0; i < n; i++ {
		id := NewObjectID()
		fmt.Printf("NewObjectID: [%s]\n", id.Hex())
	}
}
