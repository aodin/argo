package main

import (
    "log"
)

const (
    INSERT = iota
    UPDATE
    DELETE
)

// Just the name of the Key that changed and its values
type ChangeEvent struct {
    Key string
    PrevValue interface{}
    NewValue interface{}
}

type Entry struct  {
    listeners map[chan *ChangeEvent] struct{}
}

func (e *Entry) Listen() chan *ChangeEvent {
    // Create a new listener and add it to the Entry
    listener := make(chan *ChangeEvent)
    e.listeners[listener] = struct{}{}
    return listener
}

func (e *Entry) Trigger() {
    // Send an event to every listener
    event := &ChangeEvent{Key: "WAT"}
    for listener, _ := range e.listeners {
        listener <- event
    }
}

func NewEntry() *Entry {
    listeners := make(map[chan *ChangeEvent] struct{})
    return &Entry{listeners}
}

func main() {

    log.Println(INSERT, UPDATE, DELETE)


    e := NewEntry()
    listener := e.Listen()
    end := make(chan struct{})

    go func() {
        // Wait for an event to occur on the listener, then log
        log.Println("Received message:", <- listener)
        end <- struct{}{}
    }()
    // Trigger an event
    e.Trigger()

    <- end
}