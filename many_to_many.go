package argo

import (
	"fmt"

	sql "github.com/aodin/aspect"
)

type ManyToManyElem struct {
	name       string // Key where values will be added to parent table
	resourceFK sql.ForeignKeyElem
	elementFK  sql.ForeignKeyElem
	table      *sql.TableElem
	through    *sql.TableElem
	resource   *ResourceSQL
	selects    Columns
	detailOnly bool
}

func (elem ManyToManyElem) DetailOnly() ManyToManyElem {
	elem.detailOnly = true
	return elem
}

// Exclude removes fields on the  element table from the query
func (elem ManyToManyElem) Exclude(names ...string) ManyToManyElem {
	for _, name := range names {
		if _, ok := elem.table.C[name]; !ok {
			panic(fmt.Sprintf(
				"argo: cannot exclude '%s' from many to many, table '%s' does not have a column with this name",
				name,
				elem.table.Name,
			))
		}
		// Remove the column from the list of selected columns
		if err := elem.selects.Remove(name); err != nil {
			panic(fmt.Sprintf(
				"argo: the column '%s' cannot be excluded - it either does not exist or has already been excluded",
				name,
			))
		}
	}
	return elem
}

// IncludeThrough adds fields from the through table to the query
func (elem ManyToManyElem) IncludeThrough(names ...string) ManyToManyElem {
	for _, name := range names {
		if _, ok := elem.through.C[name]; !ok {
			panic(fmt.Sprintf(
				"argo: cannot exclude '%s' from many to many, through table '%s' does not have a column with this name",
				name,
				elem.through.Name,
			))
		}

		// TODO for now, there can be no name collision between the fields
		// on the element table and those on the through table
		if err := elem.selects.Add(elem.through.C[name]); err != nil {
			panic(fmt.Sprintf(
				"argo: cannot add column '%s' from through table '%s' to selects: %s",
				name,
				elem.through.Name,
				err,
			))
		}
	}
	return elem
}

func (elem ManyToManyElem) Modify(resource *ResourceSQL) error {
	if resource.table == nil {
		return fmt.Errorf("argo: Many To Many statements can only modify resources with an existing table")
	}

	// The through table should contain foreign keys to both the current
	// element's table and the resource table.
	// TODO What if there are multiple matches?
	for _, fk := range elem.through.ForeignKeys() {
		if fk.ReferencesTable() == resource.table {
			elem.resourceFK = fk
		} else if fk.ReferencesTable() == elem.table {
			elem.elementFK = fk
		}
	}
	if elem.resourceFK.Name() == "" || elem.elementFK.Name() == "" {
		return fmt.Errorf(
			"argo: could not match the many to many relationship of '%s' to '%s' through the table '%s'",
			elem.table.Name,
			resource.table.Name,
			elem.through.Name,
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

	// TODO Add the included table to the requested methods
	resource.detailIncludes = append(resource.detailIncludes, elem)
	if !elem.detailOnly {
		resource.listIncludes = append(resource.listIncludes, elem)
	}
	return nil
}

// Query is the database query method used for single result detail methods.
func (elem ManyToManyElem) Query(c sql.Connection, values sql.Values) error {
	// The values must include the referencing name of the element foreign
	// key. The rest of the relationship is built from there.

	// TODO panic or errors?
	// TODO Query by a value that doesn't exist in values? a default value?
	fkValue, ok := values[elem.resourceFK.ForeignName()]
	if !ok {
		panic(fmt.Sprintf(
			"argo: cannot query an included table by a values key '%s' - it does not exist in the given values map",
			elem.resourceFK.ForeignName(),
		))
	}

	stmt := sql.Select(
		elem.selects,
	).Join(
		elem.through.C[elem.resourceFK.Name()],
		elem.resource.table.C[elem.resourceFK.ForeignName()],
	).Join(
		elem.through.C[elem.elementFK.Name()],
		elem.table.C[elem.elementFK.ForeignName()],
	).Where(
		elem.through.C[elem.resourceFK.Name()].Equals(fkValue),
	)

	results := make([]sql.Values, 0)
	if err := c.QueryAll(stmt, &results); err != nil {
		panic(fmt.Sprintf(
			"argo: error while querying included many for key '%d' (%s): %s",
			fkValue,
			stmt,
			err,
		))
	}

	FixValues(results...)
	values[elem.name] = results
	return nil
}

// QueryAll is the database query method used for multiple result list methods.
func (elem ManyToManyElem) QueryAll(c sql.Connection, v []sql.Values) error {
	// Get all foreign name values
	fkValues := make([]interface{}, 0)

	// TODO panic or errors
	// TODO Query by a value that doesn't exist in values?
	for _, value := range v {
		fkValue, ok := value[elem.resourceFK.ForeignName()]
		if !ok {
			panic(fmt.Sprintf(
				"argo: cannot query an included many to many table by a values key '%s' - it does not exist in the given values map",
				elem.resourceFK.ForeignName(),
			))
		}
		fkValues = append(fkValues, fkValue)
	}

	// If there are no values to query, stop here
	if len(fkValues) == 0 {
		return nil
	}

	// TODO conditional query toggles
	stmt := sql.Select(
		elem.selects,
	).Join(
		elem.through.C[elem.resourceFK.Name()],
		elem.resource.table.C[elem.resourceFK.ForeignName()],
	).Join(
		elem.through.C[elem.elementFK.Name()],
		elem.table.C[elem.elementFK.ForeignName()],
	).Where(
		elem.through.C[elem.resourceFK.Name()].In(fkValues),
	)

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
	byFkValue := make(map[interface{}][]interface{})
	for _, result := range results {
		key := result[elem.resourceFK.Name()]
		byFkValue[key] = append(byFkValue[key], result)
	}

	// Add them back into the original values array
	for _, value := range v {
		fkValues, ok := byFkValue[value[elem.resourceFK.ForeignName()]]
		if ok {
			value[elem.name] = fkValues
		} else {
			value[elem.name] = make([]interface{}, 0) // JSON output as []
		}
	}
	return nil
}

func ManyToMany(name string, table, through *sql.TableElem) ManyToManyElem {
	if table == nil || through == nil {
		panic("argo: tables in many to many statements cannot be nil")
	}
	if err := validateFieldName(name); err != nil {
		panic(err.Error())
	}
	return ManyToManyElem{
		name:    name,
		table:   table,
		through: through,
		selects: ColumnSet(table.Columns()...), // No include values by default
	}
}
