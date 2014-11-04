package argo

import (
	"encoding/json"
	"fmt"
	"io"

	sql "github.com/aodin/aspect"
)

// Resource is the common interface for JSON REST resources
type Resource interface {
	Encoder() Encoder

	List(*Request) (Response, *Error)
	Post(*Request) (Response, *Error)
	Get(r *Request) (Response, *Error)
	Patch(r *Request) (Response, *Error)
	// Put(r *Request) (Response, Error)
	Delete(r *Request) (Response, *Error)
}

// TODO Alias Handle to Resource as a hack
type Handle interface {
	Resource
}

// Common Encoding interface
type Encoder interface {
	Decode(io.Reader) (sql.Values, *Error)
	Encode(interface{}) []byte
	MediaType() string
}

// JSONResource implements JSON encoding and decoding
type JSONEncoding struct{}

// sql.Values objects encode []byte as base64. Cast them to strings.
func (c JSONEncoding) Fix(results ...sql.Values) {
	for _, result := range results {
		for k, v := range result {
			switch v.(type) {
			case []byte:
				result[k] = string(v.([]byte))
			}
		}
	}
}

func (c JSONEncoding) Decode(data io.Reader) (sql.Values, *Error) {
	values := sql.Values{}
	if err := json.NewDecoder(data).Decode(&values); err != nil {
		return values, NewError(400, err.Error())
	}
	return values, nil
}

func (c JSONEncoding) Encode(i interface{}) []byte {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		panic(fmt.Sprintf(
			"argo: could not encode response: %s",
			err,
		))
	}
	return b
}

func (c JSONEncoding) MediaType() string {
	return "application/json"
}
