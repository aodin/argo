package argo

import (
	"net/http"
	"net/url"
)

type Request struct {
	*http.Request
	Encoding Encoder
	Params   Params
	Values   url.Values
}

// Get gets a GET parameter and ONLY a get parameter - never POST form data
func (r *Request) Get(key string) string {
	if r.Values != nil {
		return r.Values.Get(key)
	}
	r.Values = r.Request.URL.Query()
	return r.Values.Get(key)
}
