package argo

import (
	"fmt"

	sql "github.com/aodin/aspect"
)

// ManyElem is the internal representation of an included Many resource.
// The included table must be connected by a Foreign key to the parent
// resource table.
type ManyElem struct {
	name       string // name where values will be added to parent table
	fk         sql.ForeignKeyElem
	table      *sql.TableElem
	resource   *ResourceSQL
	selects    Columns
	showFK     bool // By default, foreign key fields will be dropped
	detailOnly bool
	asMap      *struct {
		Key   string
		Value string
	}
}

// AsMap converts the list of many elements into a map of the given key: value.
func (elem ManyElem) AsMap(key, value string) ManyElem {
	// TODO How to guarantee that the key is unique per result?

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

// DetailOnly will attach the ManyElem to only the detail views of the API.
func (elem ManyElem) DetailOnly() ManyElem {
	elem.detailOnly = true
	return elem
}

// Exclude removes the given fields by name from the included ManyElem.
func (elem ManyElem) Exclude(names ...string) ManyElem {

	// TODO The foreign key column cannot be excluded! it is needed for
	// matching - it can be dropped automatically later

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

// Modify implements the Modifier resource that allows an element to
// modify a resource. It will add the ManyElem to the list of included
// elements for the given resource.
func (elem ManyElem) Modify(resource *ResourceSQL) error {
	if resource.table == nil {
		return fmt.Errorf("argo: Many statements can only modify resources with an existing table")
	}

	// Search the foreign keys of the included element to find a
	// foreign key that matches the resource table
	// TODO It doesn't need to be only foreign keys
	for _, fk := range elem.table.ForeignKeys() {
		if fk.ReferencesTable() == resource.table {
			elem.fk = fk
			break
		}
	}
	if elem.fk.Name() == "" {
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
	// TODO specify the HTTP methods were the include should be active
	resource.detailIncludes = append(resource.detailIncludes, elem)
	if !elem.detailOnly {
		resource.listIncludes = append(resource.listIncludes, elem)
	}
	return nil
}

// Query is the database query method used for single result detail methods.
func (elem ManyElem) Query(conn sql.Connection, values sql.Values) error {
	// TODO panic or errors
	// TODO Query by a value that doesn't exist in values?
	fkValue, ok := values[elem.fk.ForeignName()]
	if !ok {
		panic(fmt.Sprintf(
			"argo: cannot query an included table by a values key '%s' - it does not exist in the given values map",
			elem.fk.ForeignName(),
		))
	}

	stmt := sql.Select(
		elem.selects,
	).Where(
		elem.table.C[elem.fk.Name()].Equals(fkValue),
	)

	results := make([]sql.Values, 0)
	if err := conn.QueryAll(stmt, &results); err != nil {
		panic(fmt.Sprintf(
			"argo: error while querying included many for key '%d' (%s): %s",
			fkValue,
			stmt,
			err,
		))
	}

	if !elem.showFK {
		// TODO multiple fks
		for _, result := range results {
			delete(result, elem.fk.Name())
		}
	}

	FixValues(results...)
	if elem.asMap == nil {
		values[elem.name] = results
		return nil
	}

	// Convert to a map
	key := elem.asMap.Key
	value := elem.asMap.Value

	// All key results must be of type string
	// TODO if not unique allow mapping as map[string][]interface{}
	mapping := make(map[string]interface{})
	for _, result := range results {
		keyValue, ok := result[key].(string)
		if !ok {
			return fmt.Errorf(
				"argo: cannot create mapping using key '%s' - it is non-string type %T",
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

// QueryAll is the database query method used for building a many
// relationship with many tables.
func (elem ManyElem) QueryAll(c sql.Connection, values []sql.Values) error {
	// Get all foreign name values
	fkValues := make([]interface{}, 0)

	// TODO panic or errors
	// TODO Query by a value that doesn't exist in values?
	for _, value := range values {
		fkValue, ok := value[elem.fk.ForeignName()]
		if !ok {
			panic(fmt.Sprintf(
				"argo: cannot query an included table by a values key '%s' - it does not exist in the given values map",
				elem.fk.ForeignName(),
			))
		}
		fkValues = append(fkValues, fkValue)
	}

	// If there are no values to query, stop here
	if len(fkValues) == 0 {
		return nil
	}

	// TODO conditional query toggles
	// TODO the included fk field must be selected even if it is removed
	// later - it is needed to match resources
	// TODO custom order bys
	// TODO composite primary keys
	stmt := sql.Select(
		elem.selects,
	).Where(
		elem.table.C[elem.fk.Name()].In(fkValues),
	).OrderBy(elem.table.C[elem.table.PrimaryKey()[0]])

	results := make([]sql.Values, 0)
	if err := c.QueryAll(stmt, &results); err != nil {
		panic(fmt.Sprintf(
			"argo: error in query all for many with keys '%v' (%s): %s",
			fkValues, // TODO pretty print value array?
			stmt,
			err,
		))
	}

	FixValues(results...)

	// Separate them by fk value
	byFkValue := make(map[interface{}][]sql.Values)
	for _, result := range results {
		key := result[elem.fk.Name()]
		if !elem.showFK {
			// TODO multiple fks
			delete(result, elem.fk.Name())
		}
		byFkValue[key] = append(byFkValue[key], result)
	}

	// Add them back into the original values array
	// TODO as map
	for _, value := range values {
		fkValues, ok := byFkValue[value[elem.fk.ForeignName()]]
		if ok {
			value[elem.name] = fkValues
		} else {
			value[elem.name] = make([]sql.Values, 0) // JSON output as []
		}
	}
	return nil
}

// Many creates a new Many respresentation of the given table at the given name
func Many(name string, table *sql.TableElem) ManyElem {
	if table == nil {
		panic("argo: tables in many statements cannot be nil")
	}
	if err := validateFieldName(name); err != nil {
		panic(err.Error())
	}
	return ManyElem{
		name:    name,
		table:   table,
		selects: ColumnSet(table.Columns()...),
	}
}
