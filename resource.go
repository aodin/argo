package argonaut

import (
	"encoding/json"
	"errors"
)

type ResourceStruct struct {
	Name string
	fields []*FieldStruct
}

func (e *ResourceStruct) Unmarshal(raw []byte) (map[string] interface {}, error) {
	// Dump the entire JSON into a empty map
	dirty := make(map[string] interface {})
	unmarshalErr := json.Unmarshal(raw, &dirty)

	clean := make(map[string] interface {})

	if unmarshalErr != nil {
		return clean, unmarshalErr
	}

	// Aggregate errors by the field Name
	fieldErrors := make(map[string] []string)

	// Use the type checkers to clean the given map
	for _, field := range e.fields {
		fieldValue, exists := dirty[field.Name]
		typeErrors := field.Type.Check(fieldValue, exists);
		// TODO Are all errors fatal?
		if typeErrors == nil {
			clean[field.Name] = fieldValue
		} else {
			stringErrors := make([]string, len(typeErrors))
			for index, err := range typeErrors {
				stringErrors[index] = err.Error()
			}
			fieldErrors[field.Name] = stringErrors
		}
	}
	// TODO Excess fields are not an error, should they be?
	// If errors, marshal into a JSON object
	if len(fieldErrors) > 0 {
		jsonFieldErrors, _ := json.Marshal(fieldErrors)
		// TODO It is a bit silly to create a field error object, json encode
		// it and then plug it back into an error as a string
		// There should be a common interface to handle errors throughout the
		// entire system
		return clean, errors.New(string(jsonFieldErrors))
	}
	return clean, nil
}

func (e *ResourceStruct) Marshal(elem interface {}) ([]byte, error) {
	return json.Marshal(elem)
}

func Resource(name string, fieldStructs ...*FieldStruct) *ResourceStruct {
	// TODO There has to be an easiier way to do this
	fields := make([]*FieldStruct, len(fieldStructs))
	for index, field := range fieldStructs {
		fields[index] = field
	}
	return &ResourceStruct{Name: name, fields: fields}
}

type FieldStruct struct {
	Name string
	Type TypeChecker
}

func Field(fieldName string, typeCheck TypeChecker) *FieldStruct {
	return &FieldStruct{Name: fieldName, Type: typeCheck}
}