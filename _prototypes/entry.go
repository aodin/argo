package main

import (
    "encoding/json"
    "log"
)

type Entry struct {
    json map[string] interface{}
    Keys map[string] struct{} // Used as a set
}

// Completely rewrite the json / Keys attributes with a new message
func (e *Entry) Set(m, ) {
    // Determine which keys are actually changing
}

// Perform a merge, overwriting existing key/values in favor of the new ones 
func (e *Entry) Merge(m, ) {

}

func (e *Entry) Save() {
    // Emit an SQL INSERT or UPDATE statement for the entry, depending on
    // The existance of a key
}

func (e *Entry) Delete() {
    // Emit an SQL DELETE statement for the entry
}

// Given JSON must be an object
func ParseMessageJSON(m []byte) map[string] interface{} {
    // Decode into an interface
    var f interface{}
    json.Unmarshal(m, &f) // If you don't use the address, you'll get <nil>
    // TODO should also return an error in case of bad parse / type cast
    return f.(map[string] interface{})
}

func CreateEntry(rawMsg []byte) *Entry {
    // TODO Again, there should be error checking
    msg := ParseMessageJSON(rawMsg)
    keys := make(map[string] struct{}, len(msg))
    for key, _ := range msg {
        keys[key] = struct{}{}
    }
    return &Entry{json: msg, Keys: keys}
}

func main() {
    // Arbitrary JSON
    raw := []byte(`{"Name":"Alice","Home":"NV","Lived in":["CA","WA", "OR"]}`)

    entry := CreateEntry(raw)
    log.Println("New Entry:", entry)
}
