package relational

import (
	"github.com/aodin/aspect"
	"github.com/aodin/aspect/postgis"
	"testing"
)

var licenses = aspect.Table("licenses",
	aspect.Column("id", aspect.String{PrimaryKey: true}),
	aspect.Column("name", aspect.String{}),
	aspect.Column("description", aspect.String{}),
	aspect.Column("issued", aspect.Timestamp{WithTimezone: true}),
	aspect.Column("is_expired", aspect.Boolean{}),
	aspect.Column("price", aspect.Real{}),
	aspect.Column("longitude", aspect.Real{}),
	aspect.Column("latitude", aspect.Real{}),
	aspect.Column("location", postgis.Geometry{postgis.Point{}, 4326}),
)

func TestTableResource(t *testing.T) {
	ResourceFromTable(licenses)
}
