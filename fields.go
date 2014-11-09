package argo

import (
	"fmt"

	sql "github.com/aodin/aspect"
)

// Fields maintains a set of fields by name
type Fields map[string]ColumnField

func (set Fields) Add(f ColumnField) error {
	if _, exists := set[f.Name]; exists {
		return fmt.Errorf(
			"argo: a field named %s already exists in this set",
			f.Name,
		)
	}
	set[f.Name] = f
	return nil
}

func (set Fields) Has(name string) bool {
	_, ok := set[name]
	return ok
}

func (set Fields) Remove(name string) error {
	if _, exists := set[name]; !exists {
		return fmt.Errorf(
			"argo: no field named %s exists in this set",
			name,
		)
	}
	delete(set, name)
	return nil
}

// Selectable implement's aspect's Selectable interface so the set can be
// queried directly.
// func (set Fields) Selectable() []sql.ColumnElem {
// 	columns := make([]sql.ColumnElem, len(set))
// 	var i int
// 	for _, column := range set {
// 		columns[i] = column
// 		i += 1
// 	}
// 	return columns
// }

func FieldsFromColumns(columns ...sql.ColumnElem) Fields {
	set := Fields{}
	for _, column := range columns {
		set[column.Name()] = ColumnField{c: column}
	}
	return set
}
