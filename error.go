package argo

import (
	"fmt"
	"net/http"
	"strings"
)

// APIError is an error structure with meta and field-specific errors
type APIError struct {
	code   int
	Meta   []string          `json:"meta"`
	Fields map[string]string `json:"fields"`
}

// AddMeta appends a meta error
func (e *APIError) AddMeta(msg string, args ...interface{}) {
	e.Meta = append(e.Meta, fmt.Sprintf(msg, args...))
}

// Code returns the status code that should be written to the response
func (e APIError) Code() int {
	return e.code
}

// Error implements the error interface
func (e APIError) Error() string {
	return fmt.Sprintf("%d: %s", e.code, strings.Join(e.Meta, "; "))
}

// Exists returns true if there are either meta or fields errors
func (e APIError) Exists() bool {
	return len(e.Meta) > 0 || len(e.Fields) > 0
}

// SetField sets the error message for the field. Mutiple calls will overwrite
// previous messages for that field.
func (e *APIError) SetField(field, msg string, args ...interface{}) {
	e.Fields[field] = fmt.Sprintf(msg, args...)
}

// Write writes the error to the response using the given encoder
func (e APIError) Write(w http.ResponseWriter, encoder Encoder) {
	w.WriteHeader(e.code)
	w.Header().Set("Content-Type", encoder.MediaType())
	w.Write(encoder.Encode(e))
}

// New creates a new empty API error scaffold
func NewError(code int) *APIError {
	return &APIError{
		code:   code,
		Meta:   make([]string, 0), // So JSON meta is never null
		Fields: make(map[string]string),
	}
}

// MetaError returns an API error with a pre-set meta error
func MetaError(code int, msg string, args ...interface{}) *APIError {
	return &APIError{
		code:   code,
		Meta:   []string{fmt.Sprintf(msg, args...)},
		Fields: make(map[string]string),
	}
}
