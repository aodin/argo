package argo

import (
	sql "github.com/aodin/aspect"
)

type Validator interface {
	IsRequired() bool
	Validate(interface{}) (interface{}, error)
}

// OptionalType wraps a database field. It keeps the type's underlying
// validation while marking it as optional.
type OptionalType struct {
	sql.Type
}

func (opt OptionalType) IsRequired() bool {
	return false
}

func MakeOptional(t sql.Type) OptionalType {
	return OptionalType{Type: t}
}
