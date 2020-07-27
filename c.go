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
}

type Function func(*Parameters)

type argBridge struct {
	function Function
}

type Engine struct {
	ptr   C.bia_engine_t
	fptrs []*argBridge
}

func NewEngine() (Engine, error) {
	if ptr := C.bia_engine_new(); ptr != nil {
		return Engine{ptr, nil}, nil
	}

	return Engine{}, errors.New("failed to create engine")
}

func (e *Engine) Close() error {
	C.bia_engine_free(e.ptr)

	*e = Engine{}

	return nil
}

func (e *Engine) UseBSL() error {
	if C.bia_engine_use_bsl(e.ptr) != 0 {
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

	e.fptrs = append(e.fptrs, fptr)

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
	p := &Parameters{params}

	defer p.invalidate()

	(*argBridge)(arg).function(p)
}

func (p *Parameters) invalidate() {

}
