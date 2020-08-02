package gobia

/*
#cgo LDFLAGS: -lbia
#include <bia/cbia.h>
#include <stdlib.h>
#include <stdint.h>

extern bia_creation_t functionBridge(bia_parameters_t params, void* arg);

static int engine_put_function_bridge(bia_engine_t engine, const char* name, bia_function_t function, uintptr_t arg)
{
	return bia_engine_put_function(engine, name, function, (void*)arg);
}
*/
import "C"

import (
	"errors"
	"sync"
	"unsafe"
)

type Parameters struct {
	ptr  C.bia_parameters_t
	lock sync.Locker
}

type Function func(*Parameters) interface{}

type argBridge struct {
	function Function
}

type Engine struct {
	ptr   C.bia_engine_t
	fptrs map[string]*argBridge
}

type Member struct {
	ptr C.bia_member_t
}

// NewEngine creates a new Bia engine.
func NewEngine() (Engine, error) {
	if ptr := C.bia_engine_new(); ptr != nil {
		return Engine{ptr, make(map[string]*argBridge)}, nil
	}

	return Engine{}, errors.New("failed to create engine")
}

// Close frees the resources associated with the engine.
func (e *Engine) Close() error {
	C.bia_engine_free(e.ptr)

	*e = Engine{}

	return nil
}

// UseBSL binds Bia's standard library.
func (e *Engine) UseBSL(args []string) error {
	cargs := C.malloc(C.size_t(len(args)) * C.size_t(unsafe.Sizeof(uintptr(0))))
	a := (*[1<<30 - 1]*C.char)(cargs)

	defer func() {
		for i := 0; i < len(args); i++ {
			C.free(unsafe.Pointer(a[i]))
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

// PutFunction binds a Go function to Bia.
func (e *Engine) PutFunction(name string, function Function) error {
	cname := C.CString(name)

	defer C.free(unsafe.Pointer(cname))

	fptr := &argBridge{function}

	if C.engine_put_function_bridge(e.ptr, cname, (C.bia_function_t)(unsafe.Pointer(C.functionBridge)), C.uintptr_t(uintptr(unsafe.Pointer(fptr)))) != 0 {
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
func functionBridge(params C.bia_parameters_t, arg unsafe.Pointer) C.bia_creation_t {
	p := &Parameters{params, &sync.Mutex{}}

	defer p.invalidate()

	if val := (*argBridge)(arg).function(p); val != nil {
		var result C.bia_creation_t = nil

		switch v := val.(type) {
		case int:
			C.bia_create_integer(C.longlong(v), &result)
		case int8:
			C.bia_create_integer(C.longlong(v), &result)
		case int16:
			C.bia_create_integer(C.longlong(v), &result)
		case int32:
			C.bia_create_integer(C.longlong(v), &result)
		case int64:
			C.bia_create_integer(C.longlong(v), &result)
		case float32:
			C.bia_create_double(C.double(v), &result)
		case float64:
			C.bia_create_double(C.double(v), &result)
		case string:
			c := C.CString(v)

			defer C.free(unsafe.Pointer(c))

			C.bia_create_cstring(c, &result)
		}

		return result
	}

	return nil
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
		return Member{}, errors.New("invalid parameters")
	} else if C.bia_parameters_at(p.ptr, C.size_t(index), &s) != 0 {
		return Member{}, errors.New("failed to get count")
	}

	return Member{s}, nil
}

func (p *Parameters) Get(name string) (Member, error) {
	p.lock.Lock()

	defer p.lock.Unlock()

	var s C.bia_member_t

	if p.ptr == nil {
		return Member{}, errors.New("invalid parameters")
	}

	cname := C.CString(name)

	defer C.free(unsafe.Pointer(cname))

	if C.bia_parameters_kwargs_find(p.ptr, cname, &s) != 0 {
		return Member{}, errors.New("failed to get count")
	}

	return Member{s}, nil
}

func (p *Parameters) invalidate() {
	p.lock.Lock()

	defer p.lock.Unlock()

	p.ptr = nil
}

func (m Member) getInt() (int64, error) {
	var c C.longlong

	if C.bia_member_cast_llong(m.ptr, &c) != 0 {
		return 0, errors.New("failed to cast to int")
	}

	return int64(c), nil
}

func (m Member) Cast(out interface{}) error {
	switch v := out.(type) {
	case *int:
		c, err := m.getInt()
		*v = int(c)

		return err
	case *int8:
		c, err := m.getInt()
		*v = int8(c)

		return err
	case *int16:
		c, err := m.getInt()
		*v = int16(c)

		return err
	case *int32:
		c, err := m.getInt()
		*v = int32(c)

		return err
	case *int64:
		c, err := m.getInt()
		*v = c

		return err
	case *string:
		var c *C.char

		if C.bia_member_cast_cstring(m.ptr, &c) != 0 {
			return errors.New("failed to cast to string")
		}

		*v = C.GoString(c)
	case *float32:
		var c C.double

		if C.bia_member_cast_double(m.ptr, &c) != 0 {
			return errors.New("failed to cast to double")
		}

		*v = float32(c)
	case *float64:
		var c C.double

		if C.bia_member_cast_double(m.ptr, &c) != 0 {
			return errors.New("failed to cast to double")
		}

		*v = float64(c)
	default:
		return errors.New("invalid type")
	}

	return nil
}
