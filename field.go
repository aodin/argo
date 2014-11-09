package argo

import (
	"fmt"

	sql "github.com/aodin/aspect"
)

type Field interface {
	IsRequired() bool
	Validate(interface{}) (interface{}, error)
}

func validateFieldName(name string) error {
	if name == "" {
		return fmt.Errorf("argo: invalid field name '%s'", name)
	}
	return nil
}

type ColumnField struct {
	Name string
	c    sql.ColumnElem
}
