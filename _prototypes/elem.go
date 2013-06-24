package main

import (
	"log"
	"encoding/json"
)

type AttrType int

const (
	StringAttr AttrType = iota
	Int64Attr
)

type ElementStruct struct {
	Name string
	attrs []*AttrStruct
}

func (e *ElementStruct) Unmarshal(raw string) map[string] interface {} {
	dirty := make(map[string] interface {})
	json.Unmarshal([]byte(raw), &dirty)
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

func Elem(name string, attrStructs ...*AttrStruct) *ElementStruct {
	// TODO There has to be an easiier way to do this
	attrs := make([]*AttrStruct, len(attrStructs))
	for index, attr := range attrStructs {
		attrs[index] = attr
	}
	return &ElementStruct{Name: name, attrs: attrs}
}

type AttrStruct struct {
	Name string
	Type AttrType
}

func Attr(attrName string, attrType AttrType) *AttrStruct {
	return &AttrStruct{Name: attrName, Type: attrType}
}

var item = Elem("item",
	Attr("id", StringAttr),
	Attr("name", Int64Attr),
)

// TODO interface for storage
// Example integer Collection
var initial = map[int64] map[string] interface {} {
	1: map[string] interface {} {"id": 1, "name": "Alaska"},
	2: map[string] interface {} {"id": 2, "name": "Michigan"},
	3: map[string] interface {} {"id": 3, "name": "Nevada"},
	4: map[string] interface {} {"id": 4, "name": "California"},
	5: map[string] interface {} {"id": 5, "name": "Oregon"},
	6: map[string] interface {} {"id": 6, "name": "Washington"},
}

func TestArray() []byte {
	toJSON := make([]map[string] interface {}, len(initial))
	count := 0
	for _, item := range initial {
		toJSON[count] = item
		count += 1
	}
	result, err := json.Marshal(toJSON)
	// TODO what to do with errors?
	if err != nil {
		panic(err)
	}
	return result
}

// TODO Registry


func main() {
	log.Printf("item: %+v\n", item)

	rawString := `{"name":"I am a string","bogus":"not an attr","id":3}`
	result := item.Unmarshal(rawString)
	log.Printf("json: %+v\n", result)
	
	log.Println(string(TestArray()))

}