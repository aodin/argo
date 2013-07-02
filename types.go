package argonaut

import (
	"errors"
	"fmt"
)

// The type system is not designed to cast unknown types to a known type,
// but to merely confirm that the field has the requested type.
// A Check may return multiple errors. It is provided an interface of
// the unmarshaled JSON, and whether or not it exists.
type TypeChecker interface {
	Check(interface{}, bool) []error
}

type Integer struct {
	Required bool
}

func (t Integer) Check(raw interface{}, exists bool) []error {
	if !exists && !t.Required {
		// No need to check type, it doesn't exist and isn't required
		return nil
	}
	if t.Required && !exists {
		return []error{errors.New("Field is required")}
	}
	// All JSON numbers are parsed as float64
	_, ok := raw.(float64)
	// TODO more processing (or silent cast) to check for integer
	if !ok {
		return []error{errors.New("Field must be an integer")}
	}
	return nil
}

type Float struct {
	Required bool
}

func (t Float) Check(raw interface{}, exists bool) []error {
	if !exists && !t.Required {
		// No need to check type, it doesn't exist and isn't required
		return nil
	}
	if t.Required && !exists {
		return []error{errors.New("Field is required")}
	}
	// All JSON numbers are parsed as float64
	_, ok := raw.(float64)
	if !ok {
		return []error{errors.New("Field must be a float")}
	}
	return nil
}

type Boolean struct {
	Required bool
}

func (t Boolean) Check(raw interface{}, exists bool) []error {
	if !exists && !t.Required {
		// No need to check type, it doesn't exist and isn't required
		return nil
	}
	if t.Required && !exists {
		return []error{errors.New("Field is required")}
	}
	// All JSON numbers are parsed as float64
	_, ok := raw.(bool)
	if !ok {
		return []error{errors.New("Field must be a boolean")}
	}
	return nil
}

type String struct {
	Required  bool
	MaxLength int
}

func (t String) Check(raw interface{}, exists bool) []error {
	if !exists && !t.Required {
		// No need to check type, it doesn't exist and isn't required
		return nil
	}
	if t.Required && !exists {
		return []error{errors.New("Field is required")}
	}

	// All JSON numbers are parsed as float64
	value, ok := raw.(string)
	if !ok {
		return []error{errors.New("Field must be a string")}
	}
	if len(value) == 0 {
		return []error{errors.New("Field is required")}
	}
	if t.MaxLength > 0 && len(value) > t.MaxLength {
		return []error{errors.New(fmt.Sprintf("Field cannot be longer than %d characters", t.MaxLength))}
	}
	return nil
}
