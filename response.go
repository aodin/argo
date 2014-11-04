package argo

import ()

type Response interface{}

type Meta struct {
	Limit  int64 `json:"limit"`
	Offset int64 `json:"offset"`
}

type MultiResponse struct {
	Meta    Meta        `json:"meta"`
	Results interface{} `json:"results"`
}
