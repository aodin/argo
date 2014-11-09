package argo

import (
	"fmt"

	sql "github.com/aodin/aspect"
)

// TableElem wraps the given SQL table. The sql.TableElem is not directly
// added to the resource in case we want to provide options / modifiers
// before altering the ResourceSQL (see the Resource() constructor)
type TableElem struct {
	table   *sql.TableElem
	selects Columns
}

// Exclude removes the given field names from the selects
func (elem TableElem) Exclude(names ...string) TableElem {
	for _, name := range names {
		// Remove the column name from the list of selected columns
		if err := elem.selects.Remove(name); err != nil {
			panic(fmt.Sprintf(
				"argo: the column %s cannot be excluded from the table - it either does not exist or has already been excluded",
				name,
			))
		}
	}
	return elem
}

func FromTable(table *sql.TableElem) (elem TableElem) {
	if table == nil {
		panic("argo: a table cannot be nil")
	}
	if len(table.PrimaryKey()) == 0 {
		panic("argo: tables must have a primary key")
	}
	elem.table = table
	elem.selects = ColumnSet(table.Columns()...)
	return
}
