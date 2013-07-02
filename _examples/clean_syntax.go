package main

import (
	"log"
	"net/http"
	// A period in front of the package adds it directly to the namespace
	. "github.com/aodin/argonaut"
)

var item = Resource("item",
	Field("id", Integer{}),
	Field("name", String{Required: true}),
)

func main() {
	// Create an item and register the item schema
	endpoint, apiErr := API("/api")
	if apiErr != nil {
		panic("Could not create the specified API")
	}

	// Attach it to the handler
	http.Handle(endpoint.BaseURL(), endpoint)

	items := IntegerStore(item)
	items.Create([]byte(`{"name":"Super Bass-O-Matic 1976"}`))
	items.Create([]byte(`{"name":"Swill"}`))

	resourceErr := endpoint.Register("item", items)
	if resourceErr != nil {
		panic("Could not register the given resource")
	}

	// Serve forever
	address := ":9000"
	log.Println("Running on address:", address)
	http.ListenAndServe(address, nil)
}