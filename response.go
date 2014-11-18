package argo

import (
	sql "github.com/aodin/aspect"
)

type Response interface{}

type Meta struct {
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
	order  []sql.Orderable `json:"-"`
}

type MultiResponse struct {
	Meta    Meta        `json:"meta"`
	Results interface{} `json:"results"`
}
