package argo

import (
	"fmt"

	sql "github.com/aodin/aspect"
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

	// TODO Columns that will be queried

	// Default fields

	// Unique

	// Validations

}

func (c *ResourceSQL) List(r *Request) (Response, *Error) {
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

func (c *ResourceSQL) Post(r *Request) (Response, *Error) {
	values, err := c.encoder.Decode(r.Body)
	if err != nil {
		return nil, err
	}

	// Confirm all fields exist before posting

	// Check unique fields - case insensitive if string?

	stmt := c.table.Insert().Values(values)
	if stmtErr := stmt.Error(); stmtErr != nil {
		return nil, NewError(400, stmtErr.Error())
	}

	if _, dbErr := c.conn.Execute(stmt); dbErr != nil {
		panic(fmt.Sprintf(
			"argo: could not insert sql resource post (%s): %s",
			stmt,
			dbErr,
		))
	}

	// TODO Return the whole object (same as a GET?)
	fixValues(values)
	return values, nil
}

func (c *ResourceSQL) Get(r *Request) (Response, *Error) {
	// Get the primary keys
	// TODO Just one for now - but composites soon!
	pkKey := c.table.PrimaryKey()[0]
	pkValue := r.Params.ByName(pkKey)

	stmt := c.table.Select().Where(c.table.C[pkKey].Equals(pkValue))
	result := sql.Values{}
	err := c.conn.QueryOne(stmt, result)
	if err == sql.ErrNoResult {
		return nil, NewError(404, "No resource with %s %s", pkKey, pkValue)
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

func (c *ResourceSQL) Patch(r *Request) (Response, *Error) {
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
		return nil, NewError(400, stmtErr.Error())
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
		return nil, NewError(404, "No resource with %s %s", pkKey, pkValue)
	}
	return values, nil
}

func (c *ResourceSQL) Delete(r *Request) (Response, *Error) {
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
		return nil, NewError(404, "No resource with %s %s", pkKey, pkValue)
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
			panic(err)
		}
	}
	return resource
}
