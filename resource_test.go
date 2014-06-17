package argo

import (
	"net/url"
	"testing"
)

// Create an example resource
type exampleResource []string

func (r exampleResource) Get(parameters url.Values) Response {
	return Response{
		StatusCode:  200,
		ContentType: "application/json",
		Results:     r,
	}
}

func TestResource(t *testing.T) {
	e := exampleResource{"hello", "goodbye"}
	var _ Resource = e
}
