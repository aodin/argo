package argo

import (
	"fmt"
	"strconv"
	"strings"

	sql "github.com/aodin/aspect"
	"github.com/aodin/aspect/postgres"
)

type Modifier interface {
	Modify(*ResourceSQL) error
}

type Include interface {
	Query(sql.Connection, sql.Values) error
	QueryAll(sql.Connection, []sql.Values) error
}

// ResourceSQL is the internal representation of a REST resource backed by SQL.
type ResourceSQL struct {
	Name    string
	conn    sql.Connection
	table   *sql.TableElem
	selects Columns
	inserts Columns
	fields  map[string]Validator

	// Includes
	listIncludes   []Include
	detailIncludes []Include

	// Default values
	limit   int
	offset  int
	order   []sql.Orderable // Default ordering is the pks ascending
	filters map[string]Filter

	// TODO save pk columns
	// TODO Unique and foreign keys that must be checked
}

// parseMeta parses the GET variables of the request and creates a Meta object
// that can be directly added to the response. It will return defaults for the
// collection when the requested values are unsafe.
func (c *ResourceSQL) parseMeta(r *Request) (meta Meta) {
	var err error

	// Get all request parameters
	values := r.QueryValues()

	// TODO there should be limit max
	meta.Limit, err = strconv.Atoi(values.Get("limit"))
	if err != nil || meta.Limit < 1 {
		meta.Limit = c.limit
	}

	var ok bool
	if _, ok = values["limit"]; ok {
		delete(values, "limit")
	}

	meta.Offset, err = strconv.Atoi(r.Get("offset"))
	if err != nil || meta.Offset < 0 {
		meta.Offset = c.offset
	}

	if _, ok = values["offset"]; ok {
		delete(values, "offset")
	}

	meta.order = c.parseOrder(r.Get("order"))
	if len(meta.order) < 1 {
		// Fallback to default (primary keys ascending)
		meta.order = c.order
	}

	if _, ok = values["order"]; ok {
		delete(values, "order")
	}

	// Perform default filtering on the remaining fields
	for k, _ := range values {
		// The values of query values are slices, just get the first
		v := values.Get(k)
		if v == "" {
			continue
		}

		// If this is a filtering selection, add it to the meta filters
		var filter Filter
		if filter, ok = c.filters[k]; !ok {
			continue
		}
		meta.filters = append(meta.filters, filter.Filter(v))
	}

	// TODO foreign key matching?

	return
}

// parseOrder: field names are separated by commas, descending is
// marked by hyphens.
// TODO Sean hates this.
func (c *ResourceSQL) parseOrder(get string) (order []sql.Orderable) {
	parts := strings.Split(get, ",")
	for _, part := range parts {
		var desc bool
		// TODO error on bad columns / format
		if part != "" && part[0] == '-' {
			desc = true
			part = part[1:]
		}

		// TODO columns can't start with a hyphen
		column, exists := c.table.C[part]
		if !exists {
			continue
		}
		if desc {
			order = append(order, column.Desc())
		} else {
			order = append(order, column.Asc())
		}
	}
	return
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

// List returns the collection view of this sql resource.
func (c *ResourceSQL) List(r *Request) (Response, *APIError) {
	// Parse meta information for limit, offset, and order
	meta := c.parseMeta(r)
	stmt := sql.Select(
		c.selects,
	).OrderBy(meta.order...).Offset(meta.Offset).Limit(meta.Limit)

	if len(meta.filters) > 0 {
		stmt = stmt.Where(sql.AllOf(meta.filters...))
	}

	results := make([]sql.Values, 0)
	if err := c.conn.QueryAll(stmt, &results); err != nil {
		panic(fmt.Sprintf(
			"argo: could not query all in table resource list (%s): %s",
			stmt,
			err,
		))
	}
	FixValues(results...)

	// Add the includes
	for _, include := range c.listIncludes {
		if dbErr := include.QueryAll(c.conn, results); dbErr != nil {
			panic(fmt.Sprintf(
				"argo: could not query all includes in sql resource list (%s): %s",
				dbErr,
			))
		}
	}

	return MultiResponse{Meta: meta, Results: results}, nil
}

func (c *ResourceSQL) Post(r *Request) (Response, *APIError) {
	values, apiErr := r.Decode(r.Body)
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
	values, apiErr := r.Decode(r.Body)
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
func Resource(t TableElem, fields ...Modifier) *ResourceSQL {
	name := t.Name
	if err := InvalidName(name); err != nil {
		panic(err)
	}

	// Resources are JSON encoded by default
	resource := &ResourceSQL{
		Name:    name,
		table:   t.table, // the parameter table is argo.TableElem
		selects: t.selects,
		inserts: ColumnSet(t.table.Columns()...),
		fields:  make(map[string]Validator),

		// Default values - TODO how to set max?
		limit:   10000,
		filters: make(map[string]Filter),
	}

	// Set the default filters using the selectable columns
	for _, column := range t.selects {
		// TODO These clauses should not be hardcoded - but some kind
		// of "comparable" interface
		// For now, the actual filter will just modify the Post field
		switch column.Type().(type) {
		case sql.String:
			resource.filters[column.Name()] = StringFilter{column: column}
		default:
			resource.filters[column.Name()] = EqualsFilter{column: column}
		}
	}

	// Remove the primary key column(s) from the directly inserted columns
	// TODO Allow this behavior to be toggled
	for _, pk := range t.table.PrimaryKey() {
		if err := resource.inserts.Remove(pk); err != nil {
			panic(err)
		}

		// Construct the default ordering from the primary keys
		resource.order = append(resource.order, t.table.C[pk].Asc())
	}

	// TODO Make sure the table has no keywords, e.g. order, limit, offset

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
