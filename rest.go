package argo

import ()

// Rest is the common interface for JSON REST resources
type Rest interface {
	Encoder() Encoder

	List(*Request) (Response, *Error)
	Post(*Request) (Response, *Error)
	Get(r *Request) (Response, *Error)
	Patch(r *Request) (Response, *Error)
	// Put(r *Request) (Response, Error)
	Delete(r *Request) (Response, *Error)
}

// TODO Alias Handle to Rest as a hack
type Handle interface {
	Rest
}
