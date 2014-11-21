package argo

import (
	"strings"

	sql "github.com/aodin/aspect"
)

type Filter interface {
	Filter(string) sql.Clause
}

type StringFilter struct {
	column sql.ColumnElem
}

func (s StringFilter) Filter(v string) sql.Clause {
	return s.column.ILike(`%` + strings.TrimSpace(v) + `%`)
}

type EqualsFilter struct {
	column sql.ColumnElem
}

func (s EqualsFilter) Filter(v string) sql.Clause {
	return s.column.Equals(v)
}
