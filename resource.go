package argo

import (
	"fmt"
	"net/url"
)

type Resource interface {
	Get(url.Values) Response
}

type Resources map[string]Resource

func (r Resources) URLs(root string) map[string]string {
	urls := make(map[string]string)
	for key, _ := range r {
		urls[key] = fmt.Sprintf("%s/%s/", root, key)
	}
	return urls
}
