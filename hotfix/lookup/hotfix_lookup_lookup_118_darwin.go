//go:build go1.18

//WARNING:test if the go version not satisfy, 1.16版本以下，自行查找Mod定义匹配以下吧。

// 1.18 相比于之前版本 ModuleData 新加字段，新增函数 textAddr 寻址。

package lookup

import (
	"fmt"
	"reflect"
	"runtime"
	_ "strings"
	"unsafe"
)

// GetFunc gets the function defined by the given fully-qualified name. The
// outFuncPtr parameter should be a pointer to a function with the appropriate
// type (e.g. the address of a local variable), and is set to a new function
// value that calls the specified function. If the specified function does not
// exist, outFuncPtr is not set and an error is returned.
func GetFunc(outFuncPtr interface{}, name string) error {
	codePtr, err := FindFuncWithName(name)
	if err != nil {
		return err
	}
	CreateFuncForCodePtr(outFuncPtr, codePtr)
	return nil
}

// Convenience struct for modifying the underlying code pointer of a function
// value. The actual struct has other values, but always starts with a code
// pointer.
type Func struct {
	codePtr uintptr
}

// CreateFuncForCodePtr is given a code pointer and creates a function value
// that uses that pointer. The outFun argument should be a pointer to a function
// of the proper type (e.g. the address of a local variable), and will be set to
// the result function value.
func CreateFuncForCodePtr(outFuncPtr interface{}, codePtr uintptr) {
	outFuncVal := reflect.ValueOf(outFuncPtr).Elem()
	// Use reflect.MakeFunc to create a well-formed function value that's
	// guaranteed to be of the right type and guaranteed to be on the heap
	// (so that we can modify it). We give a nil delegate function because
	// it will never actually be called.
	newFuncVal := reflect.MakeFunc(outFuncVal.Type(), nil)
	// Use reflection on the reflect.Value (yep!) to grab the underling
	// function value pointer. Trying to call newFuncVal.Pointer() wouldn't
	// work because it gives the code pointer rather than the function value
	// pointer. The function value is a struct that starts with its code
	// pointer, so we can swap out the code pointer with our desired value.
	funcValuePtr := reflect.ValueOf(newFuncVal).FieldByName("ptr").Pointer()
	funcPtr := (*Func)(unsafe.Pointer(funcValuePtr))
	funcPtr.codePtr = codePtr
	outFuncVal.Set(newFuncVal)
}

// FindFuncWithName searches through the moduledata table created by the linker
// and returns the function's code pointer. If the function was not found, it
// returns an error. Since the data structures here are not exported, we copy
// them below (and they need to stay in sync or else things will fail
// catastrophically).
func FindFuncWithName(name string) (uintptr, error) {
	mods := activeModules()
	for _, moduleData := range mods {
		for _, ftab := range moduleData.ftab {
			entry := textAddr(moduleData, ftab.entry)
			if moduleData.minpc > uintptr(entry) || uintptr(entry) > moduleData.maxpc {
				continue
			}

			f := (*runtime.Func)(unsafe.Pointer(&moduleData.pclntable[ftab.funcoff]))
			// (*Func).Name() assumes that the *Func was created by some exported
			// method that would have returned a nil *Func pointer IF the
			// desired function's datap resolves to nil.
			// (a.k.a. if findmoduledatap(pc) returns nil)
			// Since the last element of the moduleData.ftab has a datap of nil
			// (from experimentation), .Name() Seg Faults on the last element.
			//
			// If we instead ask the external function FuncForPc to fetch
			// our *Func object, it will check the datap first and give us
			// a proper nil *Func, that .Name() understands.
			// The down side of doing this is that internally, the
			// findmoduledatap(pc) function is called twice for every element
			// we loop over.
			f = runtime.FuncForPC(f.Entry())
			cur := f.Name()
			if cur == name {
				return f.Entry(), nil
			}
		}
	}
	return 0, fmt.Errorf("invalid function name: %s", name)
}

// Everything below is taken from the runtime package, and must stay in sync
// with it.

//go:linkname activeModules runtime.activeModules
func activeModules() []*Moduledata

//go:linkname textAddr runtime.(*moduledata).textAddr
func textAddr(m *Moduledata, off32 uint32) uintptr

// moduledata records information about the layout of the executable
// image. It is written by the linker. Any changes here must be
// matched changes to the code in cmd/internal/ld/symtab.go:symtab.
// moduledata is stored in statically allocated non-pointer memory;
// none of the pointers here are visible to the garbage collector.
type Moduledata struct {
	pcHeader     *struct{}
	funcnametab  []byte
	cutab        []uint32
	filetab      []byte
	pctab        []byte
	pclntable    []byte
	ftab         []Functab
	findfunctab  uintptr
	minpc, maxpc uintptr

	text, etext           uintptr
	noptrdata, enoptrdata uintptr
	data, edata           uintptr
	bss, ebss             uintptr
	noptrbss, enoptrbss   uintptr
	end, gcdata, gcbss    uintptr
	types, etypes         uintptr
	rodata                uintptr
	gofunc                uintptr // go.func.*

	textsectmap []textsect
	typelinks   []int32 // offsets from types
	itablinks   []*struct{}

	ptab []PtabEntry

	pluginpath string
	pkghashes  []interface{}

	modulename   string
	modulehashes []interface{}

	hasmain uint8 // 1 if module contains the main function, 0 otherwise

	gcdatamask, gcbssmask Bitvector

	typemap map[typeOff]*struct{} // offset to *_rtype in previous module

	bad bool // module failed to load and should be ignored

	next *Moduledata
}

type Functab struct {
	entry   uint32
	funcoff uint32
}

type Bitvector struct {
	n        int32 // # of bits
	bytedata *uint8
}

type textsect struct {
	vaddr    uintptr // prelinked section vaddr
	end      uintptr // vaddr + section length
	baseaddr uintptr // relocated section address
}

type nameOff int32
type typeOff int32
type textOff int32

type PtabEntry struct {
	name nameOff
	typ  typeOff
}

// 防止逃逸检测失败
//
//go:noinline
func makePtr(p uintptr) unsafe.Pointer {
	return unsafe.Pointer(&p)
}

func MakeValueByFunctionName(target interface{}, name string) (reflect.Value, error) {
	src := reflect.ValueOf(target)
	if src.Kind() != reflect.Func {
		return src, fmt.Errorf("%s is not function", src.String())
	}
	ptr, err := FindFuncWithName(name)
	if err != nil {
		return src, err
	}
	val := (*[2]uintptr)(unsafe.Pointer(&src))
	(*val)[1] = uintptr(makePtr(ptr))
	return src, nil
}
