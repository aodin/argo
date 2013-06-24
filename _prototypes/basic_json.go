package main

import (
    "encoding/json"
    "log"
)

type Message struct {
    Name string
    Id int64
    Amount float64
}

func main() {
    // Create a Message
    msg := Message{Name: "Alice", Id: 3, Amount: 5.66}

    // Create a json-encoded []byte
    // Ignore errors
    msgJSON, _ := json.Marshal(msg)
    log.Printf("JSON: %s\n", msgJSON)

    // And restore to a message
    var m Message
    json.Unmarshal(msgJSON, &m)

    // The following also works
    // m := &Message{}
    // json.Unmarshal(msgJSON, m)
    log.Printf("Restored message: %v\n", m)
}
