package argo

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	sql "github.com/aodin/aspect"
	"gopkg.in/yaml.v2"
)

// Decoder is the common decoding interface
type Decoder interface {
	Decode(io.Reader) (sql.Values, *APIError)
}

// Encoder is the common encoding interface
type Encoder interface {
	Encode(interface{}) []byte // Our responses should never error
	MediaType() string
}

// JSON implements JSON encoding and decoding
type JSON struct{}

func (c JSON) Decode(data io.Reader) (sql.Values, *APIError) {
	values := sql.Values{}
	if err := json.NewDecoder(data).Decode(&values); err != nil {
		return values, MetaError(400, err.Error())
	}
	return values, nil
}

func (c JSON) Encode(i interface{}) []byte {
	// TODO turn off pretty printing by default?
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		panic(fmt.Sprintf(
			"argo: could not JSON encode response: %s",
			err,
		))
	}
	return b
}

func (c JSON) MediaType() string {
	return "application/json"
}

// YAML implements YAML encoding and decoding
type YAML struct{}

func (c YAML) Decode(data io.Reader) (sql.Values, *APIError) {
	values := sql.Values{}
	// TODO limit reader
	b, err := ioutil.ReadAll(data)
	if err != nil {
		return values, MetaError(400, err.Error())
	}
	if err = yaml.Unmarshal(b, &values); err != nil {
		return values, MetaError(400, err.Error())
	}
	return values, nil
}

func (c YAML) Encode(i interface{}) []byte {
	b, err := yaml.Marshal(i)
	if err != nil {
		panic(fmt.Sprintf(
			"argo: could not YAML encode response: %s",
			err,
		))
	}
	return b
}

func (c YAML) MediaType() string {
	return "application/x-yaml"
}
