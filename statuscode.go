package httpx

import (
	"fmt"
)

var (
	scMap map[uint]string = map[uint]string{
		// status code: "reason phrase"
		400: "Bad Request",
	}
)

type StatusCode struct {
	StatusCode   uint
	ReasonPhrase string
	Message      string
}

func NewStatusCode(sc uint, msg string) *StatusCode {
	v, ok := scMap[sc]
	if !ok {
		panic(fmt.Errorf("Undefined status code specified: status code = %d", sc))
	}

	return &StatusCode{
		StatusCode:   sc,
		ReasonPhrase: v,
		Message:      msg,
	}
}
