package argonaut

import (
	"encoding/json"
)

type ElementStruct struct {
	Name string
	attrs []*AttributeStruct
}

func (e *ElementStruct) Unmarshal(raw []byte) map[string] interface {} {
	dirty := make(map[string] interface {})
	json.Unmarshal(raw, &dirty)
	// TODO required fields?

	clean := make(map[string] interface {})
	// TODO non-arbitrary arbitrary decode?
	for _, attr := range e.attrs {
		if attrValue, exists := dirty[attr.Name]; exists {
			clean[attr.Name] = attrValue
		}
		// TODO warning otherwise?
	}
	return clean
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

type AttributeStruct struct {
	Name string
	// TODO type?
}

func Attr(attrName string) *AttributeStruct {
	return &AttributeStruct{Name: attrName}
}