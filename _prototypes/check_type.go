package main

import (
	"errors"
	"encoding/json"
	"log"
	"strings"
)

// Types
// -----

type Integer struct {
	Required bool
}

func (s Integer) Check(raw interface{}, exists bool) error {
	// All JSON numbers are parsed as float64
	_, ok := raw.(float64)

	// TODO more processing (or silent cast) to check for integer
	if ok {
		return nil
	}
	// TODO More specific error message
	return errors.New("Value was not an integer")
}

type String struct {
	Required bool
	MaxLength int
}
// TODO How to handle additional pseudo-keyword attributes on the type
// Intermediate type that handles cast / check attributes

func (s String) Check(raw interface{}, exists bool) error {
	if !exists && !s.Required {
		// No need to check type, it doesn't exist and isn't required
		return nil
	}

	value, ok := raw.(string)
	if !ok {
		return errors.New("Value was not a string")
	}

	// TODO Need the attr name for readable errors
	if s.Required && len(value) == 0 {
		return errors.New("Value is required")
	}

	// TODO Aggregate errors - common error system

	// TODO More specific error message
	return nil
} 

type TypeChecker interface {
	// Ugly, but typesafe
	Check(interface{}, bool) []errors
}


// Element
// -------

type ElementStruct struct {
	Name string
	attrs []*AttributeStruct
}

func (e *ElementStruct) Unmarshal(raw []byte) (map[string] interface {}, error) {
	dirty := make(map[string] interface {})
	json.Unmarshal(raw, &dirty)
	// TODO required fields?

	// TODO turn into an attribute map
	attrErrors := make([]string, 0)

	clean := make(map[string] interface {})
	// TODO non-arbitrary arbitrary decode?
	for _, attr := range e.attrs {
		attrValue, exists := dirty[attr.Name]
		attrErr := attr.Type.Check(attrValue, exists);
		// TODO Add it to the "clean" map even if error / value doesn't exist?
		if attrErr != nil {
			attrErrors = append(attrErrors, attrErr.Error())
		} else {
			clean[attr.Name] = attrValue
		}
	}
	if len(attrErrors) != 0 {
		// TODO The error creation should be an interface
		return clean, errors.New(strings.Join(attrErrors, "; "))
	}

	// TODO return multiple errors?
	return clean, nil
}

func (e *ElementStruct) Marshal(elem interface {}) ([]byte, error) {
	return json.Marshal(elem);
}

func Elem(name string, attrStructs ...*AttributeStruct) *ElementStruct {
	// TODO There has to be an easiier way to do this
	attrs := make([]*AttributeStruct, len(attrStructs))
	for index, attr := range attrStructs {
		attrs[index] = attr
	}
	return &ElementStruct{Name: name, attrs: attrs}
}


// Attributes
// ----------

type AttributeStruct struct {
	Name string
	Type TypeChecker
}

func Attr(attrName string, typeCheck TypeChecker) *AttributeStruct {
	return &AttributeStruct{Name: attrName, Type: typeCheck}
}


// Example Schema
// --------------

var item = Elem("item",
	Attr("id", Integer{Required: true}),
	Attr("name", String{Required: true}),
)

func main() {
	log.Printf("Item: %+v\n", item)

	good := []byte(`{"name":"Super Bass-o-Matic 1976","id":1}`)
	result, err := item.Unmarshal(good)
	if err != nil {
		log.Printf("Error on good: %s\n", err.Error())
	}
	log.Printf("Good result: %+v\n", result)


	empty := []byte(`{"name":"","id":1}`)
	result, err = item.Unmarshal(empty)
	if err != nil {
		log.Printf("Error on empty: %s\n", err.Error())
	}
	log.Printf("Empty result: %+v\n", result)


	bad := []byte(`{"id":"Super Bass-o-Matic 1976","name":1}`)
	result, err = item.Unmarshal(bad)
	if err != nil {
		log.Printf("Error on bad: %s\n", err.Error())
	}
	log.Printf("Bad result: %+v\n", result)

	bogus := []byte(`{}`)
	result, err = item.Unmarshal(bogus)
	if err != nil {
		log.Printf("Error on bogus: %s\n", err.Error())
	}
	log.Printf("Bogus result: %+v\n", result)

}