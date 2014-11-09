package argo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlugify(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("ca-cb", slugify("Ca CB"))
	assert.Equal("co-in", slugify("CO_IN"))
}
