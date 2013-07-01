package argonaut

import (
	"testing"
)

func TestTypeChecker(t *testing.T) {
	// JSON unmarshal returns numbers as a float64
	var intValue float64 = 10 

	// Check an optional integer
	optionalInt := Integer{}
	typeErrors := optionalInt.Check(intValue, true)
	if typeErrors != nil {
		t.Error("Errors during optional integer check:", typeErrors)
	}
	typeErrors = optionalInt.Check(nil, false)
	if typeErrors != nil {
		t.Error("Errors during optional integer check:", typeErrors)
	}
	typeErrors = optionalInt.Check("What", true)
	if typeErrors == nil {
		t.Error("An error was expected during an optional integer check but did not occur")
	}

	// Check a required integer
	requiredInt := Integer{Required: true}
	typeErrors = requiredInt.Check(intValue, true)
	if typeErrors != nil {
		t.Error("Errors during required integer check:", typeErrors)
	}
	typeErrors = requiredInt.Check(nil, false)
	if typeErrors == nil {
		t.Error("An error was expected during a required integer check but did not occur")
	}
	typeErrors = requiredInt.Check("What", true)
	if typeErrors == nil {
		t.Error("An error was expected during a required integer check but did not occur")
	}

	// TODO Floats will pass the integer check

	// Check an optional string
	optionalString := String{}
	typeErrors = optionalString.Check("What", true)
	if typeErrors != nil {
		t.Error("Errors during optional string check:", typeErrors)
	}
	typeErrors = optionalString.Check(nil, false)
	if typeErrors != nil {
		t.Error("Errors during optional string check:", typeErrors)
	}
	typeErrors = optionalString.Check(10, true)
	if typeErrors == nil {
		t.Error("An error was expected during an optional string check but did not occur")
	}

	// Check an required string
	requiredString := String{Required: true}
	typeErrors = requiredString.Check("What", true)
	if typeErrors != nil {
		t.Error("Errors during required string check:", typeErrors)
	}
	typeErrors = requiredString.Check(nil, false)
	if typeErrors == nil {
		t.Error("An error was expected during an required string check but did not occur")
	}
	typeErrors = requiredString.Check(10, true)
	if typeErrors == nil {
		t.Error("An error was expected during an required string check but did not occur")
	}

	// Check a optional string with a max length
	optionalMaxLengthString := String{MaxLength: 10}
	typeErrors = optionalMaxLengthString.Check("What", true)
	if typeErrors != nil {
		t.Error("Errors during optional max length string check:", typeErrors)
	}
	typeErrors = optionalMaxLengthString.Check(nil, false)
	if typeErrors != nil {
		t.Error("Errors during optional max length string check:", typeErrors)
	}
	typeErrors = optionalMaxLengthString.Check(10, true)
	if typeErrors == nil {
		t.Error("An error was expected during an optional max length string check but did not occur")
	}
	typeErrors = optionalMaxLengthString.Check("Longer than Max", true)
	if typeErrors == nil {
		t.Error("An error was expected during an optional max length string check but did not occur")
	}

	// Check an optional boolean
	optionalBoolean := Boolean{}
	typeErrors = optionalBoolean.Check(true, true)
	if typeErrors != nil {
		t.Error("Errors during optional boolean check:", typeErrors)
	}
	typeErrors = optionalBoolean.Check(nil, false)
	if typeErrors != nil {
		t.Error("Errors during optional boolean check:", typeErrors)
	}
	typeErrors = optionalBoolean.Check(10, true)
	if typeErrors == nil {
		t.Error("An error was expected during an optional boolean check but did not occur")
	}

	// Check an required boolean
	requiredBoolean := Boolean{Required: true}
	typeErrors = requiredBoolean.Check(false, true)
	if typeErrors != nil {
		t.Error("Errors during required boolean check:", typeErrors)
	}
	typeErrors = requiredBoolean.Check(nil, false)
	if typeErrors == nil {
		t.Error("An error was expected during an required boolean check but did not occur")
	}
	typeErrors = requiredBoolean.Check(10, true)
	if typeErrors == nil {
		t.Error("An error was expected during an required boolean check but did not occur")
	}
}
