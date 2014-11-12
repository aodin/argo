package argo

import (
	"fmt"

	sql "github.com/aodin/aspect"
	"github.com/aodin/aspect/postgres"
)

type Modifier interface {
	Modify(*ResourceSQL) error
}

// ResourceSQL is the internal representation of a REST resource backed by SQL.
type ResourceSQL struct {
	Name    string
	conn    sql.Connection
	table   *sql.TableElem
	selects Columns
	inserts Columns

	// Includes
	listIncludes   []ManyElem
	detailIncludes []ManyElem

	// TODO save pk columns
	// TODO Unique and foreign keys that must be checked
}

func (c *ResourceSQL) Validate(values sql.Values) *APIError {
	// Create an empty error scaffold
	err := NewError(400)
	for key, value := range values {
		column, exists := c.inserts[key]
		if !exists {
			err.SetField(key, "does not exist")
			continue
		}
		clean, validateErr := column.Type().Validate(value)
		if validateErr != nil {
			err.SetField(column.Name(), validateErr.Error())
			continue
		}
		values[key] = clean
	}
	if err.Exists() {
		return err
	}
	return nil
}

func (c *ResourceSQL) HasRequired(values sql.Values) *APIError {
	// TODO use an existing error scaffold?
	err := NewError(400)
	for _, column := range c.inserts {
		// TODO precompute required fields
		if column.Type().IsRequired() {
			if _, exists := values[column.Name()]; !exists {
				err.SetField(column.Name(), "is required")
			}
		}
	}
	if err.Exists() {
		return err
	}
	return nil
}

func (c *ResourceSQL) List(r *Request) (Response, *APIError) {
	stmt := sql.Select(c.selects)
	results := make([]sql.Values, 0)
	if err := c.conn.QueryAll(stmt, &results); err != nil {
		panic(fmt.Sprintf(
			"argo: could not query all in table resource list (%s): %s",
			stmt,
			err,
		))
	}
	FixValues(results...)
	return MultiResponse{Results: results}, nil
}

func (c *ResourceSQL) Post(r *Request) (Response, *APIError) {
	values, apiErr := r.Decoding.Decode(r.Body)
	if apiErr != nil {
		return nil, apiErr
	}
	if len(values) == 0 {
		return nil, MetaError(
			400,
			"refusing to create an entry without values",
		)
	}

	// TODO persist errors?
	// Validate all fields
	if apiErr = c.Validate(values); apiErr != nil {
		return nil, apiErr
	}

	// Check required fields
	if apiErr = c.HasRequired(values); apiErr != nil {
		return nil, apiErr
	}

	// TODO Check existence of foreign keys
	// TODO Check unique fields - case insensitive if string?

	// TODO only one pk for now
	key := c.table.PrimaryKey()[0]

	// Check for uniques using strict equality
	uniques := c.table.UniqueConstraints()
	for _, unique := range uniques {
		// TODO Alternate forms of equality
		columns := make([]sql.Selectable, len(unique))
		clauses := make([]sql.Clause, len(unique))
		for i, name := range unique {
			columns[i] = c.table.C[name]
			clauses[i] = c.table.C[name].Equals(values[name])
		}
		stmt := sql.Select(columns...).Where(sql.AllOf(clauses...))

		result := sql.Values{}
		dbErr := c.conn.QueryOne(stmt, result)
		if dbErr == nil {
			return nil, MetaError(400, "duplicate entry for values %s", result)
		} else if dbErr != sql.ErrNoResult {
			panic(fmt.Sprintf(
				"argo: could not select uniques in sql resource post (%s): %s",
				stmt,
				dbErr,
			))
		}
	}

	stmt := postgres.Insert(c.inserts).Returning(c.table.C[key]).Values(values)
	if stmtErr := stmt.Error(); stmtErr != nil {
		return nil, MetaError(400, stmtErr.Error())
	}

	var pk interface{}
	if dbErr := c.conn.QueryOne(stmt, &pk); dbErr != nil {
		panic(fmt.Sprintf(
			"argo: could not insert in sql resource post (%s): %s",
			stmt,
			dbErr,
		))
	}

	// Send the created resource back
	selectStmt := sql.Select(c.selects).Where(c.table.C[key].Equals(pk))

	// If we get ErrNoResult then something is fucked
	result := sql.Values{}
	if dbErr := c.conn.QueryOne(selectStmt, result); dbErr != nil {
		panic(fmt.Sprintf(
			"argo: could not query one in sql resource post (%s): %s",
			selectStmt,
			dbErr,
		))
	}
	FixValues(result)
	return result, nil
}

func (c *ResourceSQL) Get(r *Request) (Response, *APIError) {
	// Get the primary keys
	// TODO Just one for now - but composites soon!
	key := c.table.PrimaryKey()[0]
	dirtyPK := r.Params.ByName(key)

	// Validate the primary key value TODO multiple pk values
	cleanPK, err := c.table.C[key].Type().Validate(dirtyPK)
	if err != nil {
		apiErr := NewError(400)
		apiErr.SetField(key, err.Error())
		return nil, apiErr
	}

	stmt := sql.Select(c.selects).Where(c.table.C[key].Equals(cleanPK))
	result := sql.Values{}
	dbErr := c.conn.QueryOne(stmt, result)
	if dbErr == sql.ErrNoResult {
		return nil, MetaError(404, "no resource with %s %s", key, dirtyPK)
	} else if dbErr != nil {
		panic(fmt.Sprintf(
			"argo: could not query one in sql resource get (%s): %s",
			stmt,
			dbErr,
		))
	}

	FixValues(result)

	// Add the includes
	for _, include := range c.detailIncludes {
		if dbErr := include.Query(c.conn, result); dbErr != nil {
			panic(fmt.Sprintf(
				"argo: could not query includes in sql resource get (%s): %s",
				dbErr,
			))
		}
	}
	return result, nil
}

func (c *ResourceSQL) Patch(r *Request) (Response, *APIError) {
	// Get the primary keys
	// TODO Just one for now - but composites soon!
	key := c.table.PrimaryKey()[0]
	dirtyPK := r.Params.ByName(key)

	// Validate the primary key value TODO multiple pk values
	cleanPK, err := c.table.C[key].Type().Validate(dirtyPK)
	if err != nil {
		apiErr := NewError(400)
		apiErr.SetField(key, err.Error())
		return nil, apiErr
	}

	// Validate all fields
	values, apiErr := r.Decoding.Decode(r.Body)
	if apiErr != nil {
		return nil, apiErr
	}
	if apiErr := c.Validate(values); apiErr != nil {
		return nil, apiErr
	}

	// Check unique fields - case insensitive if string?
	stmt := c.table.Update().Values(values).Where(
		c.table.C[key].Equals(cleanPK),
	)
	if stmtErr := stmt.Error(); stmtErr != nil {
		return nil, MetaError(400, stmtErr.Error())
	}

	// Perform the UPDATE
	changes, err := c.conn.Execute(stmt)
	if err != nil {
		panic(fmt.Sprintf(
			"argo: could not execute sql resource patch (%s): %s",
			stmt,
			err,
		))
	}

	// If no rows were affected, then no row exists at this id
	rows, err := changes.RowsAffected()
	if err != nil {
		panic(fmt.Sprintf(
			"argo: unsupported RowsAffected in sql resource patch %s",
			err,
		))
	}
	if rows == 0 {
		return nil, MetaError(404, "No resource with %s %s", key, dirtyPK)
	}

	// Send the created resource back
	selectStmt := sql.Select(c.selects).Where(c.table.C[key].Equals(cleanPK))

	// If we get ErrNoResult then something is fucked
	result := sql.Values{}
	if dbErr := c.conn.QueryOne(selectStmt, result); dbErr != nil {
		panic(fmt.Sprintf(
			"argo: could not query one in sql resource patch (%s): %s",
			selectStmt,
			dbErr,
		))
	}

	// TODO includes?

	FixValues(result)
	return result, nil
}

func (c *ResourceSQL) Delete(r *Request) (Response, *APIError) {
	// Get the primary keys
	// TODO Just one for now - but composites soon!
	key := c.table.PrimaryKey()[0]
	dirtyPK := r.Params.ByName(key)

	// Validate the primary key value TODO multiple pk values
	cleanPK, err := c.table.C[key].Type().Validate(dirtyPK)
	if err != nil {
		apiErr := NewError(400)
		apiErr.SetField(key, err.Error())
		return nil, apiErr
	}

	stmt := c.table.Delete().Where(c.table.C[key].Equals(cleanPK))
	result, err := c.conn.Execute(stmt)
	if err != nil {
		panic(fmt.Sprintf(
			"argo: could not execute sql resource delete (%s): %s",
			stmt,
			err,
		))
	}

	// If no rows were affected, then no row exists at this id
	rows, err := result.RowsAffected()
	if err != nil {
		panic(fmt.Sprintf(
			"argo: unsupported RowsAffected in sql resource delete %s",
			err,
		))
	}
	if rows == 0 {
		return nil, MetaError(404, "No resource with %s %s", key, dirtyPK)
	}
	return nil, nil
}

func InvalidName(name string) error {
	if name == "" {
		return fmt.Errorf("argo: invalid resource name '%s'", name)
	}
	return nil
}

// Resource created a new ResourceSQL from the given table and modifiers.
// Panic on errors.
func Resource(c sql.Connection, t TableElem, fields ...Modifier) *ResourceSQL {
	name := t.Name
	if err := InvalidName(name); err != nil {
		panic(err)
	}

	// Resources are JSON encoded by default
	resource := &ResourceSQL{
		Name:    name,
		conn:    c,
		table:   t.table, // the parameter table is argo.TableElem
		selects: t.selects,
		inserts: ColumnSet(t.table.Columns()...),
	}

	// Remove the primary key column(s) from the directly inserted columns
	// TODO Allow this behavior to be toggled
	for _, pk := range t.table.PrimaryKey() {
		if err := resource.inserts.Remove(pk); err != nil {
			panic(err)
		}
	}

	for _, field := range fields {
		if err := field.Modify(resource); err != nil {
			panic(fmt.Sprintf(
				"argo: failed to modify resource: %s",
				err,
			))
		}
	}
	return resource
}
