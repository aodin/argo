package argonaut

import (
	"encoding/json"
	"testing"
)

func TestCollection(t *testing.T) {
	// Example schema
	item := Elem("item",
		Attr("id"),
		Attr("name"),
	)
	collection := IntegerStore(item)

	// Create an item with a given id
	colonBlow := map[string] interface{} {
		"id": 2, // TODO should error is anything but int64/float64 key
		"name": "Colon Blow",
	}
	colonBlowJSON, _ := json.Marshal(colonBlow)
	t.Log(string(colonBlowJSON))
	createdId, created, err := collection.Create(colonBlowJSON)
	if err != nil {
		t.Error(err)
	}
	if createdId != "2" {
		t.Errorf("Expected id 2, received %d\n", createdId)
	}
	// TODO test content
	t.Log(string(created))

	// Create an item without a pre-specified id
	blankItem := []byte(`{"name":"New Shimmer"}`)
	createdId, created, err = collection.Create(blankItem)
	if err != nil {
		t.Error(err)
	}
	if createdId != "3" {
		t.Errorf("Expected id 3, received %d\n", createdId)
	}
	// TODO test content
	t.Log(string(created))

	// Attempt to create a duplicate item
	if _, _, dupCreateErr := collection.Create(colonBlowJSON); dupCreateErr == nil {
		t.Errorf("Expected an error when creating a duplicate item\n")
	}

	// Attempt to create a improperly formatted item
	// TODO For now, it just creates an empty item
	// if _, _, formatErr := collection.Create([]byte(`BOGUS`)); formatErr == nil {
	// 	t.Errorf("Expected an error when creating an improperly formatted item\n")
	// }

	// Fetch the created item
	createdItem, readErr := collection.Read("2")
	if readErr != nil {
		t.Error(readErr)
	}
	// TODO order doesn't matter for the JSON, how to compare?
	t.Log(string(createdItem))

	// Delete an item
	if deleteErr := collection.Delete("2"); deleteErr != nil {
		t.Error(deleteErr)
	}

	// Try to delete an item that does not exist
	if expectErr := collection.Delete("2"); expectErr == nil {
		t.Error("Expected an error while trying to delete non-existant item")
	}

	// Update an item
	updateJSON := []byte(`{"id":3,"name":"Super Colon Blow"}`)
	updated, updateErr := collection.Update("3", updateJSON)
	if updateErr != nil {
		t.Error(updateErr)
	}
	t.Log(string(updated))

	collection.Create([]byte(`{"name":"Jam Hawkers"}`))
	collection.Create([]byte(`{"name":"Super Bass-O-Matic 1976"}`))
	collection.Create([]byte(`{"name":"Swill"}`))
	collection.Create([]byte(`{"name":"HiberNol"}`))

	output := collection.List()
	// TODO test content
	t.Log(string(output))

}
