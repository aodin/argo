package main

import (
	"log"
	"net/http"
	"github.com/aodin/argonaut"
)

var item = argonaut.Elem("item",
	argonaut.Attr("id"),
	argonaut.Attr("name"),
)

func main() {
	// Create an item and register the item schema
	endpoint, apiErr := argonaut.API("/api")
	if apiErr != nil {
		panic("Could not create the specified API")
	}

	// Attach it to the handler
	http.Handle(endpoint.BaseURL(), endpoint)

	items := argonaut.IntegerStore(item)
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