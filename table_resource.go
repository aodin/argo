package argo

import (
	"fmt"

	sql "github.com/aodin/aspect"
)

type JSONResource struct {
	encoder JSONEncoding
	table   *sql.TableElem
	conn    sql.Connection
}

func (c *JSONResource) Encoder() Encoder {
	return c.encoder
}

func (c *JSONResource) List(r *Request) (Response, *Error) {
	stmt := c.table.Select()
	results := make([]sql.Values, 0)
	if err := c.conn.QueryAll(stmt, &results); err != nil {
		panic(fmt.Sprintf(
			"argo: could not query all in table resource list (%s): %s",
			stmt,
			err,
		))
	}
	c.encoder.Fix(results...)
	return MultiResponse{Results: results}, nil
}

func (c *JSONResource) Post(r *Request) (Response, *Error) {
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
			"argo: could not insert table resource post (%s): %s",
			stmt,
			dbErr,
		))
	}

	// TODO Return the whole object (same as a GET?)
	c.encoder.Fix(values)
	return values, nil
}

func (c *JSONResource) Get(r *Request) (Response, *Error) {
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
			"argo: could not query one in table resource get (%s): %s",
			stmt,
			err,
		))
	}
	c.encoder.Fix(result)
	return result, nil
}

func (c *JSONResource) Patch(r *Request) (Response, *Error) {
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
			"argo: could not execute table resource patch (%s): %s",
			stmt,
			err,
		))
	}

	// If no rows were affected, then no row exists at this id
	rows, err := result.RowsAffected()
	if err != nil {
		panic(fmt.Sprintf(
			"argo: unsupported RowsAffected in table resource patch %s",
			err,
		))
	}
	if rows == 0 {
		return nil, NewError(404, "No resource with %s %s", pkKey, pkValue)
	}
	return values, nil
}

func (c *JSONResource) Delete(r *Request) (Response, *Error) {
	// Get the primary keys
	// TODO Just one for now - but composites soon!
	pkKey := c.table.PrimaryKey()[0]
	pkValue := r.Params.ByName(pkKey)

	// TODO cast to id type?

	stmt := c.table.Delete().Where(c.table.C[pkKey].Equals(pkValue))
	result, err := c.conn.Execute(stmt)
	if err != nil {
		panic(fmt.Sprintf(
			"argo: could not execute table resource delete (%s): %s",
			stmt,
			err,
		))
	}

	// If no rows were affected, then no row exists at this id
	rows, err := result.RowsAffected()
	if err != nil {
		panic(fmt.Sprintf(
			"argo: unsupported RowsAffected in table resource delete %s",
			err,
		))
	}
	if rows == 0 {
		return nil, NewError(404, "No resource with %s %s", pkKey, pkValue)
	}
	return nil, nil
}

func NewJSONResource(conn sql.Connection, t *sql.TableElem) *JSONResource {
	// Handle only single primary keys
	return &JSONResource{
		encoder: JSONEncoding{},
		table:   t,
		conn:    conn,
	}
}
