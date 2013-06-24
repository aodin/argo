package main

import (
    "encoding/json"
    "log"
)

type Entry struct {
    Keys map[string] struct{} // Used as a set
}

func main() {
    // Arbitrary JSON
    j := []byte(`{"Name":"Alice","Home":"NV","Lived in":["CA","WA", "OR"]}`)

    // Decode into an interface
    var f interface{}
    json.Unmarshal(j, &f) // If you don't use the address, you'll get <nil>

    log.Printf("Decoded message: %v\n", f)

    // Given JSON must be an object in this example
    msg := f.(map[string] interface{})

    // Iterate through the keys
    for key, _ := range msg {
        log.Println("Key:", key)
    }
}
