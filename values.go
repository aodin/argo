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

func ValidateUsing(values sql.Values, types map[string]sql.Type) *APIError {
	// Create an empty error scaffold
	err := NewError(400)
	for key, value := range values {
		t, exists := types[key]
		if !exists {
			err.SetField(key, "does not exist")
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
