package argo

import (
	"encoding/json"
	"net/url"
	"testing"
	"time"

	sql "github.com/aodin/aspect"
	"github.com/aodin/aspect/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type user struct {
	ID       int64  `json:"id,omitempty"`
	Name     string `json:"name"`
	Age      int64  `json:"age"`
	IsActive bool   `json:"is_active"`
	Password string `json:"password"`
}

var usersDB = sql.Table("users",
	sql.Column("id", postgres.Serial{}),
	sql.Column("name", sql.String{NotNull: true}),
	sql.Column("age", sql.Integer{NotNull: true}),
	sql.Column("is_active", sql.Boolean{Default: sql.True}),
	sql.Column("password", sql.String{NotNull: true}),
	sql.Column("created", sql.Timestamp{Default: "now() at time zone 'utc'"}),
	sql.PrimaryKey("id"),
	sql.Unique("name"),
)

type edge struct {
	A int64 `json:"a"`
	B int64 `json:"b"`
}

var edgesDB = sql.Table("edges",
	sql.Column("id", postgres.Serial{}),
	sql.Column("a", sql.Integer{NotNull: true}),
	sql.Column("b", sql.Integer{NotNull: true}),
	sql.PrimaryKey("id"),
	sql.Unique("a", "b"),
)

func initSchemas(t *testing.T, tables ...*sql.TableElem) (*sql.DB, sql.Transaction) {
	// Connect to the database specified in the test db.json config
	// Default to the Travis CI settings if no file is found
	conf, err := sql.ParseTestConfig("./db.json")
	if err != nil {
		t.Fatalf(
			"argo: failed to parse test configuration, test aborted: %s",
			err,
		)
	}
	conn, err := sql.Connect(conf.Driver, conf.Credentials())
	require.Nil(t, err)

	// Perform all tests in a transaction and always rollback
	tx, err := conn.Begin()
	require.Nil(t, err)

	// Create the given schemas
	for _, table := range tables {
		_, err = tx.Execute(table.Create())
		require.Nil(t, err)
	}
	return conn, tx
}

func TestParseOrder(t *testing.T) {
	assert := assert.New(t)

	users := Resource(FromTable(usersDB))

	assert.Equal(
		[]sql.Orderable(nil),
		users.parseOrder(""),
	)
	assert.Equal(
		[]sql.Orderable{usersDB.C["id"].Asc()},
		users.parseOrder("id"),
	)
	assert.Equal(
		[]sql.Orderable{usersDB.C["id"].Desc()},
		users.parseOrder("-id"),
	)
	assert.Equal(
		[]sql.Orderable{usersDB.C["name"].Asc(), usersDB.C["id"].Desc()},
		users.parseOrder("name,-id"),
	)

	// Malformed input
	assert.Equal(
		[]sql.Orderable(nil),
		users.parseOrder(",,,,"),
	)
	assert.Equal(
		[]sql.Orderable(nil),
		users.parseOrder(",,what,,"),
	)
}

func TestParseMeta(t *testing.T) {
	assert := assert.New(t)

	users := Resource(FromTable(usersDB))

	// Test with no url Values
	mock := MockRequest(nil, nil)
	meta := users.parseMeta(mock)

	assert.Equal(meta.Limit, 10000)

	// With a different limit and offset
	mock = MockRequest(nil, url.Values{
		"offset": []string{"1"},
		"limit":  []string{"1"},
	})
	meta = users.parseMeta(mock)
	assert.Equal(meta.Limit, 1)
	assert.Equal(meta.Offset, 1)

	// Add a filter
	mock = MockRequest(nil, url.Values{
		"is_active": []string{"true"},
	})
	meta = users.parseMeta(mock)
	assert.Equal(
		[]sql.Clause{usersDB.C["is_active"].Equals("true")},
		meta.filters,
	)

	mock = MockRequest(nil, url.Values{
		"name": []string{"g"},
	})
	meta = users.parseMeta(mock)
	assert.Equal(
		[]sql.Clause{usersDB.C["name"].ILike(`%g%`)},
		meta.filters,
	)
}

func TestSimpleResourceSQL(t *testing.T) {
	assert := assert.New(t)
	conn, tx := initSchemas(t, usersDB)
	defer tx.Rollback()
	defer conn.Close()

	// Resources must be created with a connection
	users := Resource(FromTable(usersDB))
	users.conn = tx

	// Since *APIErr implements error, explicitly request an API error
	var errAPI *APIError

	// Get the empty list
	response, errAPI := users.List(MockRequest(nil, nil))
	assert.Nil(errAPI)
	multiResponse, ok := response.(MultiResponse)
	require.Equal(t, true, ok)
	assert.Equal(multiResponse.Meta.Offset, 0)

	results, ok := multiResponse.Results.([]sql.Values)
	require.Equal(t, true, ok)
	assert.Equal(len(results), 0)

	// POST - valid
	admin := user{Name: "admin", Age: 57, IsActive: true, Password: "haX0r"}
	b, err := json.Marshal(admin)
	require.Nil(t, err)

	response, errAPI = users.Post(MockRequest(b, nil))
	assert.Nil(errAPI)
	result, ok := response.(sql.Values)
	require.Equal(t, true, ok)
	assert.Equal(true, result["id"].(int64) > 0)
	assert.Equal(admin.Name, result["name"])
	assert.Equal(admin.Age, result["age"])
	assert.Equal(admin.IsActive, result["is_active"])
	assert.Equal(admin.Password, result["password"])
	assert.Equal(true, result["created"].(time.Time).Before(time.Now()))

	// GET - valid
	uid := result["id"].(int64)
	response, errAPI = users.Get(MockRequest(nil, nil, uid))
	assert.Nil(errAPI)
	result, ok = response.(sql.Values)
	require.Equal(t, true, ok)
	assert.Equal(uid, result["id"])

	// GET - missing id
	response, errAPI = users.Get(MockRequest(nil, nil, 0))
	assert.Equal(true, errAPI.Exists())
	assert.Equal(404, errAPI.code)
	assert.Equal(1, len(errAPI.Meta))

	// GET - invalid id
	response, errAPI = users.Get(MockRequest(nil, nil, "whatever"))
	assert.Equal(true, errAPI.Exists())
	assert.Equal(400, errAPI.code)
	assert.NotNil(errAPI.Fields["id"])

	// PATCH - missing id (data must be valid)
	response, errAPI = users.Patch(MockRequest([]byte(`{"name":"Q"}`), nil, 0))
	assert.Equal(true, errAPI.Exists())
	assert.Equal(404, errAPI.code)
	assert.Equal(1, len(errAPI.Meta))

	// PATCH - invalid id
	response, errAPI = users.Patch(MockRequest(nil, nil, "whatever"))
	assert.Equal(true, errAPI.Exists())
	assert.Equal(400, errAPI.code)
	assert.NotNil(errAPI.Fields["id"])

	// PATCH - malformed JSON
	_, errAPI = users.Patch(MockRequest([]byte(`{fsfds`), nil, uid))
	assert.Equal(true, errAPI.Exists())
	assert.Equal(400, errAPI.code)
	assert.Equal(1, len(errAPI.Meta))

	// PATCH - extra fields
	_, errAPI = users.Patch(MockRequest([]byte(`{"extra":"yup"}`), nil, uid))
	assert.Equal(true, errAPI.Exists())
	assert.Equal(400, errAPI.code)
	assert.NotNil(errAPI.Fields["extra"])

	// TODO PATCH - duplicates

	// PATCH - id
	_, errAPI = users.Patch(MockRequest([]byte(`{"id":"3"}`), nil, uid))
	assert.Equal(true, errAPI.Exists())
	assert.Equal(400, errAPI.code)
	assert.NotNil(errAPI.Fields["id"])

	// PATCH - valid
	response, errAPI = users.Patch(MockRequest([]byte(`{"name":"Q"}`), nil, uid))
	assert.Nil(errAPI)
	result, ok = response.(sql.Values)
	require.Equal(t, true, ok)
	assert.Equal(uid, result["id"])
	assert.Equal("Q", result["name"])
	assert.Equal(admin.Age, result["age"])
	assert.Equal(admin.IsActive, result["is_active"])
	assert.Equal(admin.Password, result["password"])

	// DELETE - invalid id
	response, errAPI = users.Delete(MockRequest(nil, nil, "whatever"))
	assert.Equal(true, errAPI.Exists())
	assert.Equal(400, errAPI.code)
	assert.NotNil(errAPI.Fields["id"])

	// DELETE - missing id
	response, errAPI = users.Delete(MockRequest(nil, nil, 0))
	assert.Equal(true, errAPI.Exists())
	assert.Equal(404, errAPI.code)
	assert.Equal(1, len(errAPI.Meta))

	// DELETE - valid id
	response, errAPI = users.Delete(MockRequest(nil, nil, uid))
	assert.Nil(errAPI)
	assert.Nil(response)
}

func TestResource_Post(t *testing.T) {
	assert := assert.New(t)
	conn, tx := initSchemas(t, usersDB, edgesDB)
	defer tx.Rollback()
	defer conn.Close()

	users := Resource(FromTable(usersDB))
	users.conn = tx
	edges := Resource(FromTable(edgesDB))
	edges.conn = tx

	var errAPI *APIError

	// POST without all required fields
	_, errAPI = users.Post(MockRequest([]byte(`{}`), nil))
	assert.Equal(true, errAPI.Exists())
	assert.Equal(400, errAPI.code)

	// Required fields should have specific errors
	assert.NotNil(errAPI.Fields["password"])
	assert.NotNil(errAPI.Fields["age"])
	assert.NotNil(errAPI.Fields["name"])

	// POST - include primary key
	b, err := json.Marshal(user{ID: 2, Name: "client"})
	require.Nil(t, err)
	_, errAPI = users.Post(MockRequest(b, nil))
	assert.Equal(400, errAPI.code)
	assert.NotNil(errAPI.Fields["id"])

	// POST - malformed json
	_, errAPI = users.Post(MockRequest([]byte(`{fsfds`), nil))
	assert.Equal(true, errAPI.Exists())
	assert.Equal(400, errAPI.code)
	assert.Equal(1, len(errAPI.Meta))

	// POST - extra fields
	_, errAPI = users.Post(MockRequest([]byte(`{"extra":"yup"}`), nil))
	assert.Equal(true, errAPI.Exists())
	assert.Equal(400, errAPI.code)
	assert.NotNil(errAPI.Fields["extra"])

	// POST valid
	b, err = json.Marshal(user{Name: "admin", Password: "secret"})
	require.Nil(t, err)

	_, errAPI = users.Post(MockRequest(b, nil))
	assert.Nil(errAPI)

	// POST a duplicate name
	_, errAPI = users.Post(MockRequest(b, nil))
	assert.Equal(true, errAPI.Exists())
	assert.Equal(400, errAPI.code)
	assert.Equal(1, len(errAPI.Meta))

	// Check uniqueness of composite constraints
	b, err = json.Marshal(edge{A: 2, B: 3})
	require.Nil(t, err)
	_, errAPI = edges.Post(MockRequest(b, nil))
	assert.Nil(errAPI)

	_, errAPI = edges.Post(MockRequest(b, nil))
	assert.Equal(true, errAPI.Exists())
	assert.Equal(400, errAPI.code)
	assert.Equal(1, len(errAPI.Meta))
}
