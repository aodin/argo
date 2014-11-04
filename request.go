package argo

import (
	"net/http"
)

type Request struct {
	*http.Request
	Params Params
}
