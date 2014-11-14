package argo

import (
	"encoding/json"
	"testing"

	sql "github.com/aodin/aspect"
	"github.com/aodin/aspect/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type contact struct {
	ID        int64  `json:"id,omitempty"`
	CompanyID int64  `json:"company_id"`
	Key       string `json:"key"`
	Value     string `json:"value"`
}

var contactsDB = sql.Table("contacts",
	sql.Column("id", postgres.Serial{NotNull: true}),
	sql.ForeignKey("company_id", companyDB.C["id"], sql.Integer{NotNull: true}),
	sql.Column("key", sql.String{NotNull: true}),
	sql.Column("value", sql.String{NotNull: true}),
	sql.PrimaryKey("id"),
	sql.Unique("company_id", "key", "value"),
)

func TestMany(t *testing.T) {
	assert := assert.New(t)

	// companyDB lives in many_to_many_test.go
	conn, tx := initSchemas(t, companyDB, contactsDB)
	defer tx.Rollback()
	defer conn.Close()

	companies := Resource(FromTable(companyDB), Many("contacts", contactsDB))
	companies.conn = tx
	contacts := Resource(FromTable(contactsDB))
	contacts.conn = tx

	var b []byte
	var err error
	var response interface{}
	var errAPI *APIError
	var values sql.Values
	var multiresults []sql.Values

	// Add a company and contacts
	// Get the created id from the company
	b, err = json.Marshal(company{Name: "Test Company"})
	require.Nil(t, err)
	response, errAPI = companies.Post(mockRequest(b))
	require.Nil(t, errAPI)
	values = response.(sql.Values)
	companyID := values["id"].(int64)
	assert.Equal(true, companyID > 0)

	b, err = json.Marshal(contact{
		CompanyID: companyID,
		Key:       "faceagram",
		Value:     "whatever",
	})
	require.Nil(t, err)
	_, errAPI = contacts.Post(mockRequest(b))
	require.Nil(t, errAPI)

	b, err = json.Marshal(contact{
		CompanyID: companyID,
		Key:       "flowslack",
		Value:     "whatever",
	})
	require.Nil(t, err)
	_, errAPI = contacts.Post(mockRequest(b))
	require.Nil(t, errAPI)

	// Get the companies resource with the many contacts included
	response, errAPI = companies.List(mockRequest(nil))
	require.Nil(t, errAPI)
	multiresults = response.(MultiResponse).Results.([]sql.Values)
	require.Equal(t, 1, len(multiresults))
	assert.Equal("Test Company", multiresults[0]["name"])

	// Contacts should be a list of values
	contactsValues := multiresults[0]["contacts"].([]sql.Values)
	assert.Equal(2, len(contactsValues))

	// FKs should be removed by default
	assert.Nil(contactsValues[0]["company_id"])
	assert.Nil(contactsValues[1]["company_id"])

	// TODO as map in the list view?

	response, errAPI = companies.Get(mockRequestID(nil, companyID))
	require.Nil(t, errAPI)
	values = response.(sql.Values)
	assert.Equal("Test Company", values["name"])

	contactsValues = values["contacts"].([]sql.Values)
	require.Equal(t, 2, len(contactsValues))
	assert.Equal("faceagram", contactsValues[0]["key"])
	assert.Equal("whatever", contactsValues[0]["value"])
	assert.Nil(contactsValues[0]["company_id"])

	// Output the include as a map
	asMap := Resource(
		FromTable(companyDB),
		Many("contacts", contactsDB).AsMap("key", "value"),
	)
	asMap.conn = tx

	response, errAPI = asMap.Get(mockRequestID(nil, companyID))
	require.Nil(t, errAPI)
	values = response.(sql.Values)
	assert.Equal("Test Company", values["name"])

	contactsMap := values["contacts"].(map[string]interface{})
	require.Equal(t, 2, len(contactsMap))
	assert.Equal("whatever", contactsMap["faceagram"])

	// Detail only
	detailOnly := Resource(
		FromTable(companyDB),
		Many("contacts", contactsDB).DetailOnly(),
	)
	detailOnly.conn = tx

	// Detail should still work
	response, errAPI = detailOnly.Get(mockRequestID(nil, companyID))
	require.Nil(t, errAPI)
	contactsValues = response.(sql.Values)["contacts"].([]sql.Values)
	require.Equal(t, 2, len(contactsValues))

	// But not List
	response, errAPI = detailOnly.List(mockRequest(nil))
	require.Nil(t, errAPI)
	multiresults = response.(MultiResponse).Results.([]sql.Values)
	require.Equal(t, 1, len(multiresults))
	assert.Nil(multiresults[0]["contacts"])
}
