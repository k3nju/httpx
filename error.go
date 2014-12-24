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

	v := []string{e.From.Error(), e.msg}
	return strings.Join(v, "\n")
}

func (e *Error) String() string {
	return e.Error()
}
