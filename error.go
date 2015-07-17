package httpx

import (
	"strings"
)

type Error struct {
	msg  string
	From error
}

func NewError(msg string) *Error {
	return &Error{
		msg: msg,
	}
}

func NewErrorFrom(msg string, from error) *Error {
	return &Error{
		msg:  msg,
		From: from,
	}
}

func (e *Error) Error() string {
	if e.From == nil {
		return e.msg
	}

	v := []string{e.msg, e.From.Error()}
	return strings.Join(v, " caused by error ")
}

func (e *Error) String() string {
	return e.Error()
}
