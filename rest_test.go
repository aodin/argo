package argonaut

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
)

// TODO Use httptest package

func TestEndpoint(t *testing.T) {
	// Test various names
	endpoint, apiErr := API("/v1")
	if apiErr != nil {
		t.Errorf("Error while creating API: %s\n", apiErr.Error())
	}
	// A slash should have been appended
	if endpoint.BaseURL() != "/v1/" {
		t.Errorf("Unexpected endpoint base URL: %s\n", endpoint.baseUrl)
	}

	item := Resource("item",
		Field("id", Integer{}),
		Field("name", String{Required: true}),
	)
	items := IntegerStore(item)
	items.Create([]byte(`{"name":"Super Bass-O-Matic 1976"}`))
	items.Create([]byte(`{"name":"Swill"}`))

	resourceErr := endpoint.Register("item", items)
	if resourceErr != nil {
		t.Errorf("Error while registering the given collection: %s\n", resourceErr.Error())
	}
}

// Create an example server and test methods against it
func TestRestMethods(t *testing.T) {
	// TODO Respect errors!
	item := Resource("item",
		Field("id", Integer{}),
		Field("name", String{Required: true}),
	)
	endpoint, _ := API("/api")
	endpoint.Register("item", IntegerStore(item))

	// TODO Custom test server address
	port := ":8889"
	address := "http://localhost" + port

	go func() {
		http.Handle(endpoint.BaseURL(), endpoint)
		http.ListenAndServe(port, nil)
	}()

	// TODO Use URL package to build these addresses?
	// TODO don't forget to close the body of the response
	apiURL := address + "/api/"
	response, httpErr := http.Get(apiURL)
	if httpErr != nil {
		t.Fatalf("Unexpected GET error: %s\n", httpErr.Error())
	}
	if response.StatusCode != 200 {
		t.Fatalf("Received unexpected status code during GET: %d\n", response.StatusCode)
	}

	// Add an item
	collectionURL := apiURL + "item/"
	response, httpErr = http.Post(collectionURL, "application/json", bytes.NewBuffer([]byte(`{"name":"Super Bass-O-Matic 1976"}`)))
	if httpErr != nil {
		t.Fatalf("Unexpected POST error: %s\n", httpErr.Error())
	}
	if response.StatusCode != 200 {
		t.Fatalf("Received unexpected status code during POST: %d\n", response.StatusCode)
	}

	body, readErr := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if readErr != nil {
		t.Errorf("Unexpected error during POST response: %s\n", readErr.Error())
	}

	// TODO Deep comparison of JSON?
	item1 := make(map[string]interface{})
	json.Unmarshal(body, &item1)

	id, ok := item1["id"].(float64)
	if !ok {
		t.Errorf("Returned id was not a float64")
	}
	name, ok := item1["name"].(string)
	if !ok {
		t.Errorf("Returned name was not a string")
	}
	var expectedId float64 = 1
	if id != expectedId {
		t.Errorf("Returned id was not 1")
	}
	if name != "Super Bass-O-Matic 1976" {
		t.Errorf("Unexpected name was returned: %s\n", name)
	}

	// TODO Test the required, max length, etc... type checks
}
