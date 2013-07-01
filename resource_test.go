package argonaut

import (
	"testing"
)

func TestResource(t *testing.T) {
	item := Resource("item",
		Field("id", Integer{}),
		Field("name", String{}),
	)
	if item.Name != "item" {
		t.Errorf("Unexpected element name: %s\n", item.Name)
	}
	if len(item.fields) != 2 {
		t.Fatalf("Unexpected length of attributes:", len(item.fields))
	}
	if item.fields[0].Name != "id" {
		t.Errorf("Unexpected attribute: %s\n", item.fields[0].Name)
	}
	if item.fields[1].Name != "name" {
		t.Errorf("Unexpected attribute: %s\n", item.fields[1].Name)
	}
}
