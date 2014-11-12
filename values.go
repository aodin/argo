package argo

import (
	sql "github.com/aodin/aspect"
)

// FixValues converts []byte types to string, since sql.Values objects JSON
// encode []byte as base64.
func FixValues(results ...sql.Values) {
	for _, result := range results {
		for k, v := range result {
			switch v.(type) {
			case []byte:
				result[k] = string(v.([]byte))
			}
		}
	}
}

// HasRequired confirms that all requested keys exist in the given values map.
func HasRequired(values sql.Values, keys ...string) *APIError {
	// TODO use an existing error scaffold?
	err := NewError(400)
	for _, key := range keys {
		if _, exists := values[key]; !exists {
			err.SetField(key, "is required")
		}
	}
	if err.Exists() {
		return err
	}
	return nil
}

func ValidateUsing(values sql.Values, types map[string]Validator) *APIError {
	// Create an empty error scaffold
	err := NewError(400)

	// Are all the required values present?
	// for key, validator := range types {
	// 	if validator.IsRequired() {
	// 		if _, exists := values[key]; !exists {
	// 			err.SetField(key, "is required")
	// 		}
	// 	}
	// }

	// Are the values valid?
	for key, value := range values {
		t, exists := types[key]
		// Extra fields will produce an error
		if !exists {
			err.SetField(key, "does not exist in this resource")
			continue
		}
		clean, validateErr := t.Validate(value)
		if validateErr != nil {
			err.SetField(key, validateErr.Error())
			continue
		}
		values[key] = clean
	}
	if err.Exists() {
		return err
	}
	return nil
}
