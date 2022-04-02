package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestPrintError(t *testing.T) {
	err := errors.New("hi")
	err = CauseErr(err, "hello")
	err = CauseErr(err, "world")
	fmt.Println(err.Error())
}

func TestNil(t *testing.T) {
	err := BaseErr("nil",nil)
	fmt.Println(err)
}
