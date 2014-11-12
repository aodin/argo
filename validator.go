package argo

import ()

type Validator interface {
	IsRequired() bool
	Validate(interface{}) (interface{}, error)
}
