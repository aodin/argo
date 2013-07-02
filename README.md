Argonaut
========

A JSON REST API in Go.

### Quick Start

An API is an empty placeholder for collections. Create one using:

```go
import "github.com/aodin/argonaut"

api, _ := argonaut.API("/api/")
```

The API can now be attached to Go's default HTTP route multiplexer using:

```go
http.Handle(api.BaseURL(), api)
```

This API is empty. Navigating to the URL `/api/` will return an empty list of resources.

To attach a resource, first declare a schema:

```go
import "github.com/aodin/argonaut"

var item = argonaut.Resource("item",
    argonaut.Field("id", argonaut.Integer{}),
    argonaut.Field("name", argonaut.String{Required: true}),
)
```

Or you can directly import the package into the default namespace using `.`:

```go
import . "github.com/aodin/argonaut"

var item = Resource("item",
    Field("id", Integer{}),
    Field("name", String{Required: true}),
)
```

This schema is used to create a collection. For now, there's an example in-memory key-value store called `IntegerStore`. More to come.

```go
items := IntegerStore(item)
resourceErr := endpoint.Register("item", items)
```

This will expose standard REST methods at the URL `/api/item/`.

A full example looks like:

```go
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
```

aodin, 2013