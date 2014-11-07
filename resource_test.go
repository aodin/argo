package argo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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
	sql.Column("name", sql.String{}),
	sql.Column("age", sql.Integer{}),
	sql.Column("is_active", sql.Boolean{Default: sql.True}),
	sql.Column("password", sql.String{}),
	sql.Column("created", sql.Timestamp{Default: "now() at time zone 'utc'"}),
	sql.PrimaryKey("id"),
)

type ClosingBuffer struct {
	*bytes.Buffer
}

func (cb ClosingBuffer) Close() (err error) {
	return
}

func mockRequest(body []byte) *Request {
	return &Request{
		Request: &http.Request{Body: ClosingBuffer{bytes.NewBuffer(body)}},
	}
}

func mockRequestID(body []byte, id interface{}) *Request {
	return &Request{
		Request: &http.Request{
			Body: ClosingBuffer{bytes.NewBuffer(body)},
		},
		Params: Params{{Key: "id", Value: fmt.Sprintf("%d", id)}},
	}
}

func TestSimpleResourceSQL(t *testing.T) {
	assert := assert.New(t)

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
	defer conn.Close()

	// Perform all tests in a transaction and always rollback
	tx, err := conn.Begin()
	require.Nil(t, err)
	defer tx.Rollback()

	// Create the users schema
	_, err = tx.Execute(usersDB.Create())
	require.Nil(t, err)

	// Resources must be created with a connection
	users := Resource(
		tx,
		Table(usersDB),
	)

	// Since *APIErr implements error, explicitly request an API error
	var errAPI *APIError

	// Get the empty list
	response, errAPI := users.List(&Request{})
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

	response, errAPI = users.Post(mockRequest(b))
	assert.Nil(errAPI)
	result, ok := response.(sql.Values)
	require.Equal(t, true, ok)
	assert.Equal(true, result["id"].(int64) > 0)
	assert.Equal(admin.Name, result["name"])
	assert.Equal(admin.Age, result["age"])
	assert.Equal(admin.IsActive, result["is_active"])
	assert.Equal(admin.Password, result["password"])
	assert.Equal(true, result["created"].(time.Time).Before(time.Now()))

	// POST - include primary key
	client := user{ID: 2, Name: "client"}
	b, err = json.Marshal(client)
	require.Nil(t, err)
	_, errAPI = users.Post(mockRequest(b))
	assert.Equal(400, errAPI.code)
	assert.NotNil(errAPI.Fields["id"])

	// POST - malformed json
	_, errAPI = users.Post(mockRequest([]byte(`{fsfds`)))
	assert.Equal(true, errAPI.Exists())
	assert.Equal(400, errAPI.code)
	assert.Equal(1, len(errAPI.Meta))

	// POST - extra fields
	_, errAPI = users.Post(mockRequest([]byte(`{"extra":"yup"}`)))
	assert.Equal(true, errAPI.Exists())
	assert.Equal(400, errAPI.code)
	assert.NotNil(errAPI.Fields["extra"])

	// GET - valid
	uid := result["id"].(int64)
	response, errAPI = users.Get(mockRequestID(nil, uid))
	assert.Nil(errAPI)
	result, ok = response.(sql.Values)
	require.Equal(t, true, ok)
	assert.Equal(uid, result["id"])

	// GET - missing id
	response, errAPI = users.Get(mockRequestID(nil, 0))
	assert.Equal(true, errAPI.Exists())
	assert.Equal(404, errAPI.code)
	assert.Equal(1, len(errAPI.Meta))

	// GET - invalid id
	response, errAPI = users.Get(mockRequestID(nil, "whatever"))
	assert.Equal(true, errAPI.Exists())
	assert.Equal(400, errAPI.code)
	assert.NotNil(errAPI.Fields["id"])

	// PATCH - missing id (data must be valid)
	// response, errAPI = users.Patch(mockRequestID(nil, 0))
	// assert.Equal(true, errAPI.Exists())
	// assert.Equal(404, errAPI.code)
	// assert.Equal(1, len(errAPI.Meta))

	// PATCH - invalid id
	response, errAPI = users.Patch(mockRequestID(nil, "whatever"))
	assert.Equal(true, errAPI.Exists())
	assert.Equal(400, errAPI.code)
	assert.NotNil(errAPI.Fields["id"])

	// PATCH - malformed JSON
	_, errAPI = users.Patch(mockRequestID([]byte(`{fsfds`), uid))
	assert.Equal(true, errAPI.Exists())
	assert.Equal(400, errAPI.code)
	assert.Equal(1, len(errAPI.Meta))

	// PATCH - extra fields
	_, errAPI = users.Patch(mockRequestID([]byte(`{"extra":"yup"}`), uid))
	assert.Equal(true, errAPI.Exists())
	assert.Equal(400, errAPI.code)
	assert.NotNil(errAPI.Fields["extra"])

	// TODO PATCH - duplicates

	// TODO PATCH - id
	_, errAPI = users.Patch(mockRequestID([]byte(`{"id":"3"}`), uid))
	assert.Equal(true, errAPI.Exists())
	assert.Equal(400, errAPI.code)
	assert.NotNil(errAPI.Fields["id"])

	// PATCH - valid
	response, errAPI = users.Patch(mockRequestID([]byte(`{"name":"Q"}`), uid))
	assert.Nil(errAPI)
	result, ok = response.(sql.Values)
	require.Equal(t, true, ok)
	assert.Equal(uid, result["id"])
	assert.Equal("Q", result["name"])
	assert.Equal(admin.Age, result["age"])
	assert.Equal(admin.IsActive, result["is_active"])
	assert.Equal(admin.Password, result["password"])

	// DELETE - invalid id
	response, errAPI = users.Delete(mockRequestID(nil, "whatever"))
	assert.Equal(true, errAPI.Exists())
	assert.Equal(400, errAPI.code)
	assert.NotNil(errAPI.Fields["id"])

	// DELETE - missing id
	response, errAPI = users.Delete(mockRequestID(nil, 0))
	assert.Equal(true, errAPI.Exists())
	assert.Equal(404, errAPI.code)
	assert.Equal(1, len(errAPI.Meta))

	// DELETE - valid id
	response, errAPI = users.Delete(mockRequestID(nil, uid))
	assert.Nil(errAPI)
	assert.Nil(response)
}
