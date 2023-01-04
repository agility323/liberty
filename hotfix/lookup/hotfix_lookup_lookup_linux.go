package lookup

/*
   #cgo linux LDFLAGS: -ldl
   #include <dlfcn.h>
   #include <limits.h>
   #include <stdlib.h>
   #include <stdint.h>

   #include <stdio.h>
   #ifndef RTLD_MAIN_ONLY
   #define RTLD_MAIN_ONLY 0
   #endif

   static void* localLookup(const char* name, char** err) {
   	void* r = dlsym(RTLD_MAIN_ONLY, name);
   	if (r == NULL) {
   		*err = (char*)dlerror();
   	}
   	return r;
   }
*/
import "C"
import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"
)

func FindFuncWithName(name string) (uintptr, error) {
	var cErr *C.char
	nameStr := make([]byte, len(name)+1)
	copy(nameStr, name)

	handle := C.localLookup((*C.char)(unsafe.Pointer(&nameStr[0])), &cErr)
	if handle == nil {
		return 0, errors.New(C.GoString(cErr))
	}
	return uintptr(handle), nil
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
