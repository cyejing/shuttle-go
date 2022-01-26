package utils

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

// SErr for trace err
type SErr struct {
	Msg  string
	Pc   uintptr
	File string
	Line int
}

func (e *SErr) Error() string {
	return e.Msg + " : " + e.File + ":" + strconv.Itoa(e.Line)
}

// Base trace err
func (e *SErr) Base(err error) *SErr {
	if se, ok := err.(*SErr); ok {
		e.Msg += " : " + se.Msg
		e.Pc = se.Pc
		e.File = se.File
		e.Line = se.Line
	} else {
		e.Msg += " : " + err.Error()
	}
	return e
}

// NewErrf for trace
func NewErrf(msg string, v ...interface{}) *SErr {
	return NewErr(fmt.Sprintf(msg, v...))
}

// NewErr for trace
func NewErr(msg string) *SErr {
	err := &SErr{
		Msg: msg,
	}
	for i := 0; i < 3; i++ {
		if pc, file, line, ok := runtime.Caller(i); ok && !strings.Contains(file, "pkg/utils/error.go") {
			err.Pc = pc
			err.File = file
			err.Line = line
			break
		}
	}

	return err
}

// BaseErrf new err formate
func BaseErrf(msg string, e error, v ...interface{}) *SErr {
	return NewErrf(msg, v).Base(e)
}

// BaseErr new err
func BaseErr(msg string, e error) *SErr {
	return NewErr(msg).Base(e)
}

func Err(e error) *SErr {
	return NewErr("").Base(e)
}
