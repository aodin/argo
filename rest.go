package argo

import ()

// Rest is the common interface for REST-ful resources
type Rest interface {
	List(*Request) (Response, *APIError)
	Post(*Request) (Response, *APIError)
	Get(*Request) (Response, *APIError)
	Patch(*Request) (Response, *APIError)
	Delete(*Request) (Response, *APIError)
}

// Handle is an alias for Rest
type Handle interface {
	Rest
}
