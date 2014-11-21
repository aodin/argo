package argo

import (
	"io"
	"net/http"
	"net/url"

	sql "github.com/aodin/aspect"
)

// TODO A Request constructor function

// GetEncoder matches the request Accept-Encoding header with an Encoder.
// TODO this could be done with routes / headers / auth
func GetEncoder(r *http.Request) Encoder {
	return JSON{}
}

// GetDecoder matches the request Content-Type header with a Decoder.
func GetDecoder(r *http.Request) Decoder {
	return JSON{}
}

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
	return r.QueryValues().Get(key)
}

func (r *Request) QueryValues() url.Values {
	if r.Values == nil {
		r.Values = r.Request.URL.Query()
	}
	return r.Values
}
