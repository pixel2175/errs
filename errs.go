package errs

import (
	"errors"
	"runtime/debug"
)

type Error struct {
	Err   error
	Stack []byte
}

func (e *Error) Error() string { return e.Err.Error() }
func (e *Error) Unwrap() error { return e.Err }

func throw(e error) {
	var x *Error
	if  errors.As(e, &x) {
		panic(&Error{e, x.Stack})
	}
	panic(&Error{e, debug.Stack()})
}

func Catch(err *error, fn ...func(*error)) {
	r := recover()
	if r == nil {
		return
	}
	e, ok := r.(*Error)
	if !ok {
		panic(r)
	}
	*err = e
	if len(fn) > 0 && fn[0] != nil {
		fn[0](err)
	}
}

func resolve(err error, h []any) error {
	if len(h) == 0 {
		return err
	}
	switch f := h[0].(type) {
	case func() error:
		if f != nil {
			return f()
		}
	case func(error) error:
		if f != nil {
			return f(err)
		}
	}
	return err
}

func TryE(err error, h ...any) {
	if err == nil {
		return
	}
	if e := resolve(err, h); e != nil {
		throw(e)
	}
}

func TryV[T any](v T, err error, h ...any) T {
	if err == nil {
		return v
	}
	if e := resolve(err, h); e != nil {
		throw(e)
	}
	return v
}
