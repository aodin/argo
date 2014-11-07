package argo

import (
	"bytes"
	"encoding/json"
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

	// POST valid
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

	// POST error
	// Malformed json
	_, errAPI = users.Post(mockRequest([]byte(`{fsfds`)))
	assert.NotNil(errAPI.Meta)

	// Extra fields
	_, errAPI = users.Post(mockRequest([]byte(`{"extra":"yup"}`)))
	assert.NotNil(errAPI.Meta)

	// TODO GET - valid

	// TODO GET - invalid id
}
