package argo

import (
	"testing"

	sql "github.com/aodin/aspect"
	"github.com/stretchr/testify/assert"
)

var counterDB = sql.Table("counter",
	sql.Column("id", sql.Integer{NotNull: true}),
)

func TestMakeOptional(t *testing.T) {
	assert := assert.New(t)

	typ := counterDB.C["id"].Type()
	assert.Equal(true, typ.IsRequired())

	opt := MakeOptional(typ)
	assert.Equal(false, opt.IsRequired())

	var i int64 = 1
	value, err := opt.Validate(i)
	assert.Nil(err)
	assert.Equal(i, value)

	_, err = opt.Validate("nope")
	assert.NotNil(err)
}
