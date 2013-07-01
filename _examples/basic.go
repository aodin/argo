package main

import (
	"log"
	"net/http"
	"github.com/aodin/argonaut"
)

var item = argonaut.Resource("item",
	argonaut.Field("id", argonaut.Integer{}),
	argonaut.Field("name", argonaut.String{Required: true}),
)

/*
This basic example can be tested using cURL:

To get a resource:
curl -i -X GET localhost:9000/api/item/1

To create a new resource:
curl -i -X PUT -d '{"id":3,"name":"HiberNol"}' localhost:9000/api/item

To update an existing resource:
curl -i -X PUT -d '{"id":2,"name":"Super Colon Blow"}' localhost:9000/api/item/2

To see example error:
curl -i -X PUT -d '{"id":5}' localhost:9000/api/item

*/


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