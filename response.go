package argo

import ()

type Response struct {
	ContentType string      `json:"-"`
	StatusCode  int         `json:"-"`
	Meta        Meta        `json:"meta"`
	Results     interface{} `json:"results"`
}

type Meta struct {
	Limit  int64  `json:"limit"`
	Offset int64  `json:"offset"`
	URL    string `json:"url"`
}
