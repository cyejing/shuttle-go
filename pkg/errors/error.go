package errors

import (
	"fmt"
	"io"
	"net"
	"runtime"
	"strings"
)

type CErr struct {
	Msg   string
	Pc    uintptr
	File  string
	Line  int
	raw   error
	Cause *CErr
}

func (c *CErr) Error() string {
	if c.Cause == nil {
		if c.File != "" && c.Line != 0 {
			return fmt.Sprintf("%s - %s:%d", c.Msg, c.File, c.Line)
		} else {
			return c.Msg
		}
	}
	return fmt.Sprintf("%s\n      at %s:%d : %s", c.Cause.Error(), c.File, c.Line, c.Msg)
}

func IsEOF(e error) bool {
	if c, ok := e.(*CErr); ok {
		return c.raw == io.EOF
	}
	return e == io.EOF
}
func IsNetErr(e error) bool {
	if c, ok := e.(*CErr); ok {
		return IsNetErr(c.raw)
	}
	if IsEOF(e) {
		return true
	}
	if _, ok := e.(*net.OpError); ok {
		return true
	}
	return false
}

func NewErr(msg string) *CErr {
	err := &CErr{
		Msg: msg,
	}
	for i := 0; i <= 5; i++ {
		if pc, file, line, ok := runtime.Caller(i); ok && !strings.Contains(file, "pkg/errors/error.go") {
			err.Pc = pc
			err.File = file
			err.Line = line
			break
		}
	}

	return err
}

func NewErrf(msg string, v ...interface{}) *CErr {
	return NewErr(fmt.Sprintf(msg, v...))
}

func CauseErr(e error, msg string) *CErr {
	nc := NewErr(msg)
	if c, ok := e.(*CErr); ok {
		nc.Cause = c
		nc.raw = c.raw
	} else {
		nc.Cause = &CErr{
			Msg: e.Error(),
			raw: e,
		}
	}
	return nc
}

func CauseErrf(e error, msg string, v ...interface{}) *CErr {
	return CauseErr(e, fmt.Sprintf(msg, v...))
}

func CauseErrN(e error) *CErr {
	return CauseErr(e, "")
}

func BaseErr(msg string, e error) *CErr {
	return CauseErr(e, msg)
}
func BaseErrf(msg string, e error, v ...interface{}) *CErr {
	return CauseErr(e, fmt.Sprintf(msg, v...))
}
