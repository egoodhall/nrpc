package nrpc

import (
	"strconv"

	"github.com/nats-io/nats.go"
)

type Error struct {
	code    int
	message string
	wrapped error
}

func (err *Error) Code() int {
	return err.code
}

func (err *Error) Error() string {
	return "[" + strconv.Itoa(err.code) + "] " + err.message
}

func (err *Error) Is(other error) bool {
	_, ok := other.(*Error)
	return ok
}

func (err *Error) Unwrap() error {
	return err.wrapped
}

func NewError(code int, message string) *Error {
	return &Error{code, message, nil}
}

func ParseError(msg *nats.Msg) error {
	if msg.Header.Get("Nats-Service-Error") != "" && msg.Header.Get("Nats-Service-Error-Code") != "" {
		code, err := strconv.Atoi(msg.Header.Get("Nats-Service-Error"))
		if err != nil {
			return err
		}
		return NewError(code, msg.Header.Get("Nats-Service-Error-Code"))
	}
	return nil
}
