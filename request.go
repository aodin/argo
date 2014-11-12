package argo

import (
	"io"
	"net/http"
	"net/url"

	sql "github.com/aodin/aspect"
)

type Request struct {
	*http.Request
	Encoding Encoder
	Decoding Decoder
	Params   Params
	Values   url.Values
}

func (r *Request) Decode(data io.Reader) (sql.Values, *APIError) {
	if r.Decoding == nil {
		// Default to JSON if no decoder was specified
		r.Decoding = JSON{}
	}
	return r.Decoding.Decode(data)
}

// Get gets a GET parameter and ONLY a get parameter - never POST form data
func (r *Request) Get(key string) string {
	if r.Values != nil {
		return r.Values.Get(key)
	}
	r.Values = r.Request.URL.Query()
	return r.Values.Get(key)
}
