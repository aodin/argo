package main

import (
	"log"
	"reflect"
)

type Entry struct {
	Keys map[string] interface{}
}

func TestKeyEquality(a map[string] interface{}, b map[string] interface{}) {
	// TODO another method, create a union of the keys and test all of them

	bKeysTested := make(map[string] struct{})
	inequality := make([]string, 0)
	for key, aValue := range a {
		// Tests the keys that exist in both a and b for deep equality
		if bValue, exists := b[key]; exists {
			if !reflect.DeepEqual(aValue, bValue) {
				inequality = append(inequality, key)
			}
			// Add the tested key to the "already tested" map
			bKeysTested[key] = struct{}{}
		// Add keys that exist in a but not b to the inequality array
		} else {
			inequality = append(inequality, key)
		}
	}

	// Value doesn't matter: we're testing if a key exists in b and not a
	for key, _ := range b {
		if _, exists := bKeysTested[key]; !exists {
			// Add keys that exist in b and not a to the inequality list
			inequality = append(inequality, key)
		}
	}

	log.Println("Unequal keys:", inequality)
}

func main() {

	aKeys := map[string] interface{} {
		"one": "Goodbye",
		"two": 1.99,
		"three": []string{"NV", "CA"},
	}
	a := &Entry{Keys: aKeys}

	// Keys "one", "three" and "four" differ from a
	bKeys := map[string] interface{} {
		"one": "Hello",
		"two": 1.99,
		"three": []string{"CA", "NV"},
		"four": true,
	}
	b := &Entry{Keys: bKeys}

	// Same as b
	cKeys := map[string] interface{} {
		"one": "Hello",
		"two": 1.99,
		"three": []string{"CA", "NV"},
		"four": true,
	}
	c := &Entry{Keys: cKeys}

	// Same as c except for "one"
	dKeys := map[string] interface{} {
		"two": 1.99,
		"three": []string{"CA", "NV"},
		"four": true,
	}
	d := &Entry{Keys: dKeys}

	log.Println("a == b ?", reflect.DeepEqual(a, b))
	log.Println("b == c ?", reflect.DeepEqual(b, c))

	// Test individual keys
	TestKeyEquality(a.Keys, b.Keys)
	TestKeyEquality(c.Keys, d.Keys)

}