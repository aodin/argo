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
	encoder Encoder
	conn    sql.Connection
	table   *sql.TableElem

	// TODO Whitelist columns that will be queried
	// TODO Unique and foreign keys that must be checked
}

func (c *ResourceSQL) Validate(values sql.Values) *APIError {
	// Create an empty error scaffold
	err := NewError(400)
	for key, value := range values {
		column, exists := c.table.C[key]
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

func (c *ResourceSQL) List(r *Request) (Response, *APIError) {
	stmt := c.table.Select()
	results := make([]sql.Values, 0)
	if err := c.conn.QueryAll(stmt, &results); err != nil {
		panic(fmt.Sprintf(
			"argo: could not query all in table resource list (%s): %s",
			stmt,
			err,
		))
	}
	fixValues(results...)
	return MultiResponse{Results: results}, nil
}

func (c *ResourceSQL) Post(r *Request) (Response, *APIError) {
	values, apiErr := c.encoder.Decode(r.Body)
	if apiErr != nil {
		return nil, apiErr
	}

	// Validate all fields
	if apiErr = c.Validate(values); apiErr != nil {
		return nil, apiErr
	}

	// Check required fields

	// Check existence of foreign keys

	// Check unique fields - case insensitive if string?

	// TODO only one pk for now
	key := c.table.PrimaryKey()[0]

	stmt := postgres.Insert(c.table).Returning(c.table.C[key]).Values(values)
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

	// Get the object back
	selectStmt := c.table.Select().Where(c.table.C[key].Equals(pk))

	// If we get ErrNoResult then something is fucked
	result := sql.Values{}
	if dbErr := c.conn.QueryOne(selectStmt, result); dbErr != nil {
		panic(fmt.Sprintf(
			"argo: could not query one in sql resource post (%s): %s",
			selectStmt,
			dbErr,
		))
	}
	fixValues(result)
	return result, nil
}

func (c *ResourceSQL) Get(r *Request) (Response, *APIError) {
	// Get the primary keys
	// TODO Just one for now - but composites soon!
	pkKey := c.table.PrimaryKey()[0]
	pkValue := r.Params.ByName(pkKey)

	stmt := c.table.Select().Where(c.table.C[pkKey].Equals(pkValue))
	result := sql.Values{}
	err := c.conn.QueryOne(stmt, result)
	if err == sql.ErrNoResult {
		return nil, MetaError(404, "no resource with %s %s", pkKey, pkValue)
	} else if err != nil {
		panic(fmt.Sprintf(
			"argo: could not query one in sql resource get (%s): %s",
			stmt,
			err,
		))
	}
	fixValues(result)
	return result, nil
}

func (c *ResourceSQL) Patch(r *Request) (Response, *APIError) {
	// Get the primary keys
	// TODO Just one for now - but composites soon!
	pkKey := c.table.PrimaryKey()[0]
	pkValue := r.Params.ByName(pkKey)

	// TODO cast to id type?

	values, decodeErr := c.encoder.Decode(r.Body)
	if decodeErr != nil {
		return nil, decodeErr
	}

	// Confirm all fields exist before posting

	// Check unique fields - case insensitive if string?
	stmt := c.table.Update().Values(values).Where(
		c.table.C[pkKey].Equals(pkValue),
	)
	if stmtErr := stmt.Error(); stmtErr != nil {
		return nil, MetaError(400, stmtErr.Error())
	}

	// TODO Return the whole object (same as a GET?)

	result, err := c.conn.Execute(stmt)
	if err != nil {
		panic(fmt.Sprintf(
			"argo: could not execute sql resource patch (%s): %s",
			stmt,
			err,
		))
	}

	// If no rows were affected, then no row exists at this id
	rows, err := result.RowsAffected()
	if err != nil {
		panic(fmt.Sprintf(
			"argo: unsupported RowsAffected in sql resource patch %s",
			err,
		))
	}
	if rows == 0 {
		return nil, MetaError(404, "No resource with %s %s", pkKey, pkValue)
	}
	return values, nil
}

func (c *ResourceSQL) Delete(r *Request) (Response, *APIError) {
	// Get the primary keys
	// TODO Just one for now - but composites soon!
	pkKey := c.table.PrimaryKey()[0]
	pkValue := r.Params.ByName(pkKey)

	// TODO cast to id type?

	stmt := c.table.Delete().Where(c.table.C[pkKey].Equals(pkValue))
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
		return nil, MetaError(404, "No resource with %s %s", pkKey, pkValue)
	}
	return nil, nil
}

func (c *ResourceSQL) Encoder() Encoder {
	return c.encoder
}

// sql.Values objects encode []byte as base64. Cast them to strings.
func fixValues(results ...sql.Values) {
	for _, result := range results {
		for k, v := range result {
			switch v.(type) {
			case []byte:
				result[k] = string(v.([]byte))
			}
		}
	}
}

func invalidName(name string) error {
	if name == "" {
		return fmt.Errorf("argo: invalid resource name '%s'", name)
	}
	return nil
}

// Resource created a new ResourceSQL from the given table and modifiers.
// Panic on errors.
func Resource(c sql.Connection, t TableElem, fields ...Modifier) *ResourceSQL {
	name := t.table.Name
	if err := invalidName(name); err != nil {
		panic(err)
	}

	// Resources are JSON encoded by default
	resource := &ResourceSQL{
		Name:    name,
		encoder: JSONEncoder{},
		conn:    c,
		table:   t.table, // the parameter table is argo.TableElem
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
