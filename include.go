package argo

import (
	"fmt"

	sql "github.com/aodin/aspect"
)

type IncludeElem struct {
	name        string // name where values will be added to parent table
	fkName      string // c[fkName].Equals(fkValue)
	fkValueName string
	table       *sql.TableElem
	resource    *ResourceSQL
	selects     Columns
	asMap       *struct {
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
	if resource.table == nil {
		return fmt.Errorf("argo: Many statements can only modify resources with an existing table")
	}

	// Search the foreign keys of the included element to find a
	// foreign key that matches the resource's table
	// TODO It doesn't need to be only foreign keys
	fks := elem.table.ForeignKeys()
	for _, fk := range fks {
		if fk.ReferencesTable() == resource.table {
			elem.fkName = fk.Name()
			elem.fkValueName = fk.ForeignName() // For getting from values
			break
		}
	}
	if elem.fkName == "" || elem.fkValueName == "" {
		return fmt.Errorf(
			"argo: could not match the many field '%s' to any foreign key column in '%s'",
			elem.name,
			resource.Name,
		)
	}

	// The include name can't also be taken
	// TODO set a field to prevent multiple includes at the same name
	if _, exists := resource.table.C[elem.name]; exists {
		return fmt.Errorf(
			"argo: the parent table already has a field named %s",
			elem.name,
		)
	}

	// Set the resource of the include
	elem.resource = resource

	// TODO Create a common field struct, with validation / create?

	// Add the included table to the requested methods
	// TODO detail only for now
	resource.detailIncludes = append(resource.detailIncludes, elem)

	return nil
}

// Query is the database query method used for single result detail methods.
func (elem IncludeElem) Query(conn sql.Connection, values sql.Values) error {
	// TODO panic or errors
	// TODO Query by a value that doesn't exist in values?
	fkValue, ok := values[elem.fkValueName]
	if !ok {
		panic(fmt.Sprintf(
			"argo: cannot query an included table by a values key '%s' - it does not exist in the given values map",
			elem.fkValueName,
		))
	}

	stmt := sql.Select(
		elem.selects,
	).Where(
		elem.table.C[elem.fkName].Equals(fkValue),
	)

	var results []sql.Values
	if err := conn.QueryAll(stmt, &results); err != nil {
		panic(fmt.Sprintf(
			"argo: error while querying included many for key '%d' (%s): %s",
			fkValue,
			stmt,
			err,
		))
	}

	fixValues(results...)
	if elem.asMap == nil {
		values[elem.name] = results
		return nil
	}

	// Conver to a map
	key := elem.asMap.Key
	value := elem.asMap.Value

	// All key results must be of type string
	// TODO if not unique allow mapping as map[string][]interface{}
	mapping := make(map[string]interface{})
	for _, result := range results {
		keyValue, ok := result[key].(string)
		if !ok {
			return fmt.Errorf(
				"argo: cannot creating mapping using key '%s' - it is non-string type %T",
				key,
				keyValue,
			)
		}
		// TODO error for non-unique?
		mapping[keyValue] = result[value]
	}
	values[elem.name] = mapping
	return nil
}

func Many(name string, table *sql.TableElem) IncludeElem {
	if table == nil {
		panic("argo: tables in many statements cannot be nil")
	}
	if err := validateFieldName(name); err != nil {
		panic(err.Error())
	}
	return IncludeElem{
		name:    name,
		table:   table,
		selects: ColumnSet(table.Columns()...),
	}
}
