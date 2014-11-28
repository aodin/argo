package argo

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
)

type ClosingBuffer struct {
	*bytes.Buffer
}

func (cb ClosingBuffer) Close() (err error) {
	return
}

func MockRequest(b []byte, v url.Values, ids ...interface{}) *Request {
	// TODO how to handle multiple key - value pairs
	params := Params{}
	for _, id := range ids {
		params = append(
			params,
			// TODO values should be of any type
			Param{Key: "id", Value: fmt.Sprintf("%d", id)},
		)
	}

	if v == nil {
		v = url.Values{}
	}

	return &Request{
		Request: &http.Request{
			Body: ClosingBuffer{bytes.NewBuffer(b)},
		},
		Params: params,
		Values: v,
	}
}
