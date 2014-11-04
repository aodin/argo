package argo

import "fmt"

type Error struct {
	msg  string
	code int
}

func (e Error) Error() string {
	return e.msg
}

func (e Error) Code() int {
	return e.code
}

func NewError(code int, msg string, args ...interface{}) *Error {
	return &Error{
		msg:  fmt.Sprintf(msg, args...),
		code: code,
	}
}
