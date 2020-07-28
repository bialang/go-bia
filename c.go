package gobia

// #cgo LDFLAGS: -lbase -lbsl -lbvm -lcompiler -lcbia -lstdc++ -L /usr/local/lib/bia
// #include <bia/cbia.h>
// #include <stdlib.h>
// extern void functionBridge(bia_parameters_t params, void* arg);
import "C"

import (
	"errors"
	"unsafe"
)

type Parameters struct {
	ptr C.bia_parameters_t
	lock sync.Locker
}

type Function func(*Parameters)

type argBridge struct {
	function Function
}

type Engine struct {
	ptr   C.bia_engine_t
	fptrs map[string]*argBridge
}

type Member C.bia_member_t

func NewEngine() (Engine, error) {
	if ptr := C.bia_engine_new(); ptr != nil {
		return Engine{ptr, make(map[string]*argBridge)}, nil
	}

	return Engine{}, errors.New("failed to create engine")
}

func (e *Engine) Close() error {
	C.bia_engine_free(e.ptr)

	*e = Engine{}

	return nil
}

func (e *Engine) UseBSL(args []string) error {
	cargs := C.malloc(C.size_t(len(args)) * C.size_t(unsafe.Sizeof(uintptr(0))))
	a := (*[1<<30 - 1]*C.char)(cargs)

	defer func(){
		for _, e := range a {
			C.free(e)
		}

		C.free(cargs)
	}()


	for i, e := range args {
		a[i] = C.CString(e)
	}

	if C.bia_engine_use_bsl(e.ptr, (**C.char)(cargs), C.size_t(len(args))) != 0 {
		return errors.New("failed to register bsl modules")
	}

	return nil
}

func (e *Engine) PutFunction(name string, function Function) error {
	cname := C.CString(name)

	defer C.free(unsafe.Pointer(cname))

	fptr := &argBridge{function}

	if C.bia_engine_put_function(e.ptr, cname, (C.bia_function_t)(unsafe.Pointer(C.functionBridge)), unsafe.Pointer(fptr)) != 0 {
		return errors.New("failed to put function inplace")
	}

	// to prevent GC from destroying the function
	e.fptrs[name] = fptr
	
	return nil
}

func (e *Engine) Run(code []byte) error {
	ccode := C.CBytes(code)

	defer C.free(ccode)

	if C.bia_run(e.ptr, ccode, C.size_t(len(code))) != 0 {
		return errors.New("failed to run code")
	}

	return nil
}

//export functionBridge
func functionBridge(params C.bia_parameters_t, arg unsafe.Pointer) {
	p := &Parameters{params, &sync.Mutex{}}

	defer p.invalidate()

	(*argBridge)(arg).function(p)
}

func (p *Parameters) Size() (int, error) {
	p.lock.Lock()

	defer p.lock.Unlock()

	var s C.size_t

	if p.ptr == nil {
		return 0, errors.New("invalid parameters")
	} else if C.bia_parameters_count(p.ptr, &s) != 0 {
		return 0, errors.New("failed to get count")
	}

	return int(s), nil
}

func (p *Parameters) At(index int) (Member, error) {
	p.lock.Lock()

	defer p.lock.Unlock()

	var s C.bia_member_t

	if p.ptr == nil {
		return 0, errors.New("invalid parameters")
	} else if C.bia_parameters_at(p.ptr, C.size_t(index), &s) != 0 {
		return 0, errors.New("failed to get count")
	}

	return Member(s), nil
}

func (p *Parameters) invalidate() {
	p.lock.Lock()

	defer p.lock.Unlock()

	p.ptr = nil
}

func (m Member) Cast(out interface{}) error {
	switch v := out.(type) {
	case int*:
		var c C.int
		
		if C.bia_member_cast_int(m.ptr, &c) != 0 {
			return errors.New("failed to cast to int")
		}

		*v = c
	case int8*:
	case int16*:
	case int32*:
	case int64*:
	case string*:
	case float32*:
	case float64*:
	default:
		return errors.New("invalid type")
	}

	return nil
}
