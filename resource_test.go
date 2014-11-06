package argo

import (
	"testing"

	sql "github.com/aodin/aspect"
	_ "github.com/aodin/aspect/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var usersDB = sql.Table("users",
	sql.Column("id", sql.Integer{}),
	sql.Column("name", sql.String{}),
	sql.Column("age", sql.Integer{}),
	sql.Column("password", sql.String{}),
)

func TestResourceSQL(t *testing.T) {
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
	users := Resource(tx, Table(usersDB))

	// Get the empty list
	_, err = users.List(&Request{})
	assert.Nil(err)
}
