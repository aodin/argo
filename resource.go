package argonaut

import (
	"encoding/json"
)

type ResourceStruct struct {
	Name string
	attrs []*FieldStruct
}

func (e *ResourceStruct) Unmarshal(raw []byte) map[string] interface {} {
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

func (e *ResourceStruct) Marshal(elem interface {}) ([]byte, error) {
	return json.Marshal(elem)
}

func Resource(name string, attrStructs ...*FieldStruct) *ResourceStruct {
	// TODO There has to be an easiier way to do this
	attrs := make([]*FieldStruct, len(attrStructs))
	for index, attr := range attrStructs {
		attrs[index] = attr
	}
	return &ResourceStruct{Name: name, attrs: attrs}
}

type FieldStruct struct {
	Name string
	// TODO type?
}

func Field(attrName string) *FieldStruct {
	return &FieldStruct{Name: attrName}
}