package argo

import ()

// Rest is the common interface for JSON REST resources
type Rest interface {
	List(*Request) (Response, *APIError)
	Post(*Request) (Response, *APIError)
	Get(*Request) (Response, *APIError)
	Patch(*Request) (Response, *APIError)
	Delete(*Request) (Response, *APIError)
}

// TODO Alias Handle to Rest as a hack
type Handle interface {
	Rest
}
