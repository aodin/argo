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

type Any struct {}

// Allow any type through
func (t Any) Check(raw interface{}, exists bool) []error {
	return nil
}

type Integer struct {
	Required bool
}

func (t Integer) Check(raw interface{}, exists bool) []error {
	if !t.Required && (!exists || raw == nil) {
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

type Number struct {
	Required bool
}

func (t Number) Check(raw interface{}, exists bool) []error {
	if !t.Required && (!exists || raw == nil) {
		// No need to check type, it doesn't exist and isn't required
		return nil
	}
	if t.Required && !exists {
		return []error{errors.New("Field is required")}
	}
	// All JSON numbers are parsed as float64
	_, ok := raw.(float64)
	if !ok {
		return []error{errors.New("Field must be a number")}
	}
	return nil
}

type Boolean struct {
	Required bool
}

func (t Boolean) Check(raw interface{}, exists bool) []error {
	if !t.Required && (!exists || raw == nil) {
		// No need to check type, it doesn't exist and isn't required
		return nil
	}
	if t.Required && !exists {
		return []error{errors.New("Field is required")}
	}
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
	if !t.Required && (!exists || raw == nil) {
		// No need to check type, it doesn't exist and isn't required
		return nil
	}
	if t.Required && !exists {
		return []error{errors.New("Field is required")}
	}

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

type Array struct {
	Type TypeChecker
}

func (t Array) Check(raw interface{}, exists bool) []error {
	if !exists {
		return nil
	}
	// All JSON Arrays are of type []interface{}
	elements, ok := raw.([]interface{})
	if !ok {
		return []error{errors.New("Field must be an array")}
	}

	// If a type was given, check each element
	if t.Type != nil {
		elemTypeErrors := make([]error, 0)
		for _, elem := range elements {
			// TODO aggregate errors
			elemErrors := t.Type.Check(elem, true)
			if elemErrors != nil {
				for _, elemErr := range elemErrors {
					elemTypeErrors = append(elemTypeErrors, elemErr)
				} 
			}
		}
		if len(elemTypeErrors) != 0 {
			return elemTypeErrors
		}
	}
	return nil
}

// TODO allow resources to be infinitely nested?
