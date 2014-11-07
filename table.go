package argo

import (
	sql "github.com/aodin/aspect"
)

// TableElem wraps the given SQL table. The sql.TableElem is not directly
// added to the resource in case we want to provide options / modifiers
// before altering the ResourceSQL (see the Resource() constructor)
type TableElem struct {
	table *sql.TableElem
}

func Table(table *sql.TableElem) TableElem {
	if table == nil {
		panic("argo: a table cannot be nil")
	}
	if len(table.PrimaryKey()) == 0 {
		panic("argo: tables must have a primary key")
	}
	return TableElem{
		table: table,
	}
}
