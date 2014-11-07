package argo

import (
	"fmt"

	sql "github.com/aodin/aspect"
)

type IncludeElem struct {
	name    string
	src     string
	dest    string
	table   *sql.TableElem
	selects Columns
	asMap   *struct {
		Key   string
		Value string
	}
}

func (elem IncludeElem) AsMap(key, value string) IncludeElem {
	// Both key and value must be selectable columns
	if !elem.selects.Has(key) {
		panic(fmt.Sprintf(
			"argo: the column %s is not a valid key - it either does not exist or has been excluded",
			key,
		))
	}
	if !elem.selects.Has(value) {
		panic(fmt.Sprintf(
			"argo: the column %s is not a valid value - it either does not exist or has been excluded",
			value,
		))
	}

	// Save the new mapping
	elem.asMap = &struct {
		Key   string
		Value string
	}{
		Key:   key,
		Value: value,
	}
	return elem
}

func (elem IncludeElem) Exclude(names ...string) IncludeElem {
	for _, name := range names {
		if _, ok := elem.table.C[name]; !ok {
			panic(fmt.Sprintf(
				"argo: cannot exclude %s, table %s does not have column with this name",
				name,
				elem.table.Name,
			))
		}
		// Remove the column from the list of selected columns
		if err := elem.selects.Remove(name); err != nil {
			panic(fmt.Sprintf(
				"argo: the column %s cannot be excluded - it either does not exist or has already been excluded",
				name,
			))
		}
	}
	return elem
}

func (elem IncludeElem) Modify(resource *ResourceSQL) error {
	// Confirm that the source column exists in the resource
	if _, ok := resource.table.C[elem.src]; !ok {
		return fmt.Errorf(
			"argo: source field %s of include does not exist in parent table",
			elem.src,
		)
	}

	// The include name can't also be taken
	if _, exists := resource.table.C[elem.name]; exists {
		return fmt.Errorf(
			"argo: the parent table already has a field named %s",
			elem.name,
		)
	}

	// TODO Create a common field struct, with validation / create?

	// Add the included table to the requested methods
	// TODO toggle which it gets added to
	resource.listIncludes = append(resource.listIncludes, elem)
	resource.detailIncludes = append(resource.detailIncludes, elem)
	return nil
}

// TODO Automatic matching of foreign key tables
// TODO Toggle Collection or List only
func Include(name string, table *sql.TableElem, src, dest string) IncludeElem {
	if table == nil {
		panic("argo: a table cannot be nil")
	}
	if len(table.PrimaryKey()) == 0 {
		panic("argo: tables must have a primary key")
	}
	if _, ok := table.C[dest]; !ok {
		panic(fmt.Sprintf(
			"argo: destination field %s does not exist in the included table",
			dest,
		))
	}
	return IncludeElem{
		name:    name,
		table:   table,
		selects: ColumnSet(table.Columns()...),
		src:     src,
		dest:    dest,
	}
}
