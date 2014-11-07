package argo

import (
	"fmt"

	sql "github.com/aodin/aspect"
)

// Columns maintains a set of columns by name
type Columns map[string]sql.ColumnElem

func (set Columns) Add(c sql.ColumnElem) error {
	if _, exists := set[c.Name()]; exists {
		return fmt.Errorf(
			"argo: a column named %s already exists in this set",
			c.Name(),
		)
	}
	set[c.Name()] = c
	return nil
}

func (set Columns) Has(name string) bool {
	_, ok := set[name]
	return ok
}

func (set Columns) Remove(name string) error {
	if _, exists := set[name]; !exists {
		return fmt.Errorf(
			"argo: no column named %s exists in this set",
			name,
		)
	}
	delete(set, name)
	return nil
}

// Selectable implement's aspect's Selectable interface so the set can be
// queried directly.
func (set Columns) Selectable() []sql.ColumnElem {
	columns := make([]sql.ColumnElem, len(set))
	var i int
	for _, column := range set {
		columns[i] = column
		i += 1
	}
	return columns
}

func ColumnSet(columns ...sql.ColumnElem) Columns {
	set := Columns{}
	for _, column := range columns {
		set[column.Name()] = column
	}
	return set
}
