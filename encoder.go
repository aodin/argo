package argo

import (
	"encoding/json"
	"fmt"
	"io"

	sql "github.com/aodin/aspect"
)

// Encoder is the common encoding and decoding interface
type Encoder interface {
	Decode(io.Reader) (sql.Values, *Error)
	Encode(interface{}) []byte
	MediaType() string
}

// JSONEncoder implements JSON encoding and decoding
type JSONEncoder struct{}

func (c JSONEncoder) Decode(data io.Reader) (sql.Values, *Error) {
	values := sql.Values{}
	if err := json.NewDecoder(data).Decode(&values); err != nil {
		return values, NewError(400, err.Error())
	}
	return values, nil
}

func (c JSONEncoder) Encode(i interface{}) []byte {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		panic(fmt.Sprintf(
			"argo: could not json encode response: %s",
			err,
		))
	}
	return b
}

func (c JSONEncoder) MediaType() string {
	return "application/json"
}
