package argonaut

import (
	"errors"
	"fmt"
	"strconv"
)

type Collection interface {
	Create([]byte) (string, []byte, error)
	Read(string) ([]byte, error)
	Update(string, []byte) ([]byte, error)
	Delete(string) error
	List() []byte
}

type IntegerCollection struct {
	pkey int64
	length int64
	schema *ResourceStruct
	resources map[int64] map[string] interface {}
	// TODO default ordering
}

var DoesNotExist = errors.New("No resource exists at this key")
var DuplicateResource = errors.New("A resource aready exists at this key")
var IdMismatch = errors.New("Id of resource and id provided do not match")
var ImproperKey = errors.New("An improper key was given")

// TODO another option is Create(id, elem)
func (c *IntegerCollection) Create(elem []byte) (string, []byte, error) {
	// Unpack the resource from the given byte array
	clean, unmarshalErr := c.schema.Unmarshal(elem)
	if unmarshalErr != nil {
		return "", nil, unmarshalErr
	}
	// TODO Aggregate errors?

	// An "id" may be specified in the item to be created
	// If the key did not exist, it will be handled by the type assertion
	givenId, _ := clean["id"]

	// All JSON numbers are float64
	fid, ok := givenId.(float64)
	// But we want int64
	id := int64(fid)

	// Confirm that the id is a positive integer
	if !ok || id <= 0 {
		// Increment first, because it initializes at 0
		c.pkey += 1
		id = c.pkey

		// Set the new id on the resource
		clean["id"] = id
	} else {
		// Confirm that no resource exists at the provided id
		if _, exists := c.resources[id]; exists {
			return fmt.Sprintf("%d", id), nil, DuplicateResource
		}
		// If the given key is greater than the auto-increment key, set the
		// auto-increment key equal to the given key
		// This allows the auto-increment key to keep functioning without
		// complex checks for empty id slots
		if id > c.pkey {
			c.pkey = id
		}
		// Update the clean resource to use an int64 "id" attr
		clean["id"] = id
	}
	c.resources[id] = clean
	c.length += 1

	// Return the clean copy of the newly created item
	cleanJSON, marshalErr := c.schema.Marshal(clean)
	return fmt.Sprintf("%d", id), cleanJSON, marshalErr
}

func (c *IntegerCollection) Read(key string) ([]byte, error) {
	id, keyErr := strconv.ParseInt(key, 10, 64)
	if keyErr != nil {
		return nil, ImproperKey
	}

	elem, exists := c.resources[id]
	if !exists {
		return nil, DoesNotExist
	}
	return c.schema.Marshal(elem)
}

func (c *IntegerCollection) Update(key string, elem []byte) ([]byte, error) {
	id, keyErr := strconv.ParseInt(key, 10, 64)
	if keyErr != nil {
		return nil, ImproperKey
	}
	// Unpack the resource from the given byte array
	clean, unmarshalErr := c.schema.Unmarshal(elem)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}
	// TODO Aggregate errors?

	// TODO replace the whole item or just the fields in the new elem?
	// TODO Allow resources to be moved

	// An "id" may be specified in the item to be created
	// If the key did not exist, it will be handled by the type assertion
	givenId, _ := clean["id"]

	// All JSON numbers are float64
	fid, ok := givenId.(float64)
	// But we want int64
	if !ok || id != int64(fid) {
		return nil, IdMismatch
	}

	if _, exists := c.resources[id]; !exists {
		return nil, DoesNotExist
	}
	c.resources[id] = clean

	// Return the clean copy of the updated item
	return c.schema.Marshal(clean)
}

func (c *IntegerCollection) Delete(key string) error {
	id, keyErr := strconv.ParseInt(key, 10, 64)
	if keyErr != nil {
		return ImproperKey
	}
	if _, exists := c.resources[id]; !exists {
		return DoesNotExist
	}
	delete(c.resources, id)
	c.length -= 1
	return nil
}

// TODO return an error?
func (c *IntegerCollection) List() []byte {
	output := make([]map[string] interface{}, c.length)
	count := 0
	for _, elem := range c.resources {
		output[count] = elem
		count += 1
	}
	listJSON, marshalErr := c.schema.Marshal(output)
	if marshalErr != nil {
		return nil
	}
	return listJSON
}

// TODO allow default resources - JSON or maps?
func IntegerStore(schema *ResourceStruct) *IntegerCollection {
	resources := make(map[int64] map[string] interface {})
	return &IntegerCollection{schema: schema, resources: resources}
} 