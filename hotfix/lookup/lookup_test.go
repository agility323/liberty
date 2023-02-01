package lookup

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/agility323/liberty/lbtnet"
)

func TestHotfix(t *testing.T) {
	var p uintptr
	var err error

	f1 := lbtnet.NewTcpClient
	p, err = FindFuncWithName("github.com/agility323/liberty/lbtnet.NewTcpClient")
	fmt.Printf("0x%x 0x%x %v\n", **(**uintptr)(unsafe.Pointer(&f1)), p, err)

	//github.com/agility323/liberty/lbtnet.(*TcpClient).LocalAddr
	f2 := (*lbtnet.TcpClient).LocalAddr
	p, err = FindFuncWithName("github.com/agility323/liberty/lbtnet.(*TcpClient).LocalAddr")
	fmt.Printf("0x%x 0x%x %v\n", **(**uintptr)(unsafe.Pointer(&f2)), p, err)

}
