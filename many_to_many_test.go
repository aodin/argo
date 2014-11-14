package argo

import (
	"encoding/json"
	"testing"

	sql "github.com/aodin/aspect"
	"github.com/aodin/aspect/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type company struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name"`
}

var companyDB = sql.Table("companies",
	sql.Column("id", postgres.Serial{NotNull: true}),
	sql.Column("name", sql.String{NotNull: true}),
	sql.PrimaryKey("id"),
	sql.Unique("name"),
)

type campus struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name"`
}

var campusDB = sql.Table("campuses",
	sql.Column("id", postgres.Serial{NotNull: true}),
	sql.Column("name", sql.String{NotNull: true}),
	sql.PrimaryKey("id"),
	sql.Unique("name"),
)

type companyCampus struct {
	ID        int64 `json:"id,omitempty"`
	CompanyID int64 `json:"company_id"`
	CampusID  int64 `json:"campus_id"`
	IsActive  bool  `json:"is_active"`
}

var companyCampusesDB = sql.Table("company_campuses",
	sql.Column("id", postgres.Serial{NotNull: true}),
	sql.ForeignKey("campus_id", campusDB.C["id"], sql.Integer{NotNull: true}),
	sql.ForeignKey("company_id", companyDB.C["id"], sql.Integer{NotNull: true}),
	sql.Column("is_active", sql.Boolean{Default: sql.True}),
	sql.PrimaryKey("id"),
	sql.Unique("campus_id", "company_id"),
)

func TestManyToMany(t *testing.T) {
	assert := assert.New(t)

	// Create a resource with a many to many table with through fields
	conn, tx := initSchemas(t, companyDB, campusDB, companyCampusesDB)
	defer tx.Rollback()
	defer conn.Close()

	companies := Resource(tx, FromTable(companyDB))
	campuses := Resource(
		tx,
		FromTable(campusDB),
		ManyToMany("companies", companyDB, companyCampusesDB),
	)
	companyCampuses := Resource(tx, FromTable(companyCampusesDB))

	var b []byte
	var err error
	var response interface{}
	var errAPI *APIError
	var values sql.Values

	// Add a company, campus, and companyCampus
	// Get the created id from the company and campus
	b, err = json.Marshal(company{Name: "Test Company"})
	require.Nil(t, err)
	response, errAPI = companies.Post(mockRequest(b))
	require.Nil(t, errAPI)
	values = response.(sql.Values)
	companyID := values["id"].(int64)
	assert.Equal(true, companyID > 0)

	b, err = json.Marshal(campus{Name: "Test Campus"})
	require.Nil(t, err)
	response, errAPI = campuses.Post(mockRequest(b))
	require.Nil(t, errAPI)
	values = response.(sql.Values)
	campusID := values["id"].(int64)
	assert.Equal(true, campusID > 0)

	b, err = json.Marshal(companyCampus{
		CampusID:  campusID,
		CompanyID: companyID,
		IsActive:  true,
	})
	require.Nil(t, err)
	response, errAPI = companyCampuses.Post(mockRequest(b))
	require.Nil(t, errAPI)
	values = response.(sql.Values)
	locationID := values["id"].(int64)
	assert.Equal(true, locationID > 0)

	// Write a new resource with included and excluded information
	activity := Resource(
		tx,
		FromTable(campusDB),
		ManyToMany("companies", companyDB, companyCampusesDB).Exclude("name").IncludeThrough("is_active"),
	)

	response, errAPI = activity.Get(mockRequestID(nil, locationID))
	require.Nil(t, errAPI)
	values = response.(sql.Values)
	assert.Equal("Test Campus", values["name"])

	companiesValues := values["companies"].([]sql.Values)
	require.Equal(t, 1, len(companiesValues))
	assert.Equal(1, companiesValues[0]["id"])
	assert.Equal(true, companiesValues[0]["is_active"])
	assert.Nil(companiesValues[0]["name"])
}
