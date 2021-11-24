package utils

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

type serr struct {
	Msg  string
	Pc   uintptr
	File string
	Line int
}

func (e *serr) Error() string {
	return e.Msg + " : " + e.File + ":" + strconv.Itoa(e.Line)
}

func (e *serr) Base(err error) *serr {
	if se, ok := err.(*serr); ok {
		e.Msg += " : " + se.Msg
		e.Pc = se.Pc
		e.File = se.File
		e.Line = se.Line
	} else {
		e.Msg += " : " + err.Error()
	}
	return e
}

func NewErrf(msg string, v ...interface{}) *serr {
	return NewErr(fmt.Sprintf(msg, v...))
}
func NewErr(msg string) *serr {
	err := &serr{
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
func BaseErrf(msg string, e error, v ...interface{}) *serr {
	return NewErrf(msg, v).Base(e)
}
func BaseErr(msg string, e error) *serr {
	return NewErr(msg).Base(e)
}
