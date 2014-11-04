package argo

import (
	"fmt"
	"net/http"
	"strings"
)

type API struct {
	// Create a single node for node
	resources map[string]Resource
	routes    *node
}

// Add will add the given resource at the given name
func (api *API) Add(name string, resource Resource, keys ...string) error {
	if _, exists := api.resources[name]; exists {
		return fmt.Errorf("argo: resource %s already exists", name)
	}
	api.resources[name] = resource

	// Build the routes from the primary key(s)
	pks := make([]string, len(keys))
	for i, key := range keys {
		pks[i] = fmt.Sprintf(":%s", key)
	}
	pk := strings.Join(pks, "/")

	// Also add the routes
	api.routes.addRoute(fmt.Sprintf("/%s", name), resource)
	api.routes.addRoute(fmt.Sprintf("/%s/", name), resource)
	api.routes.addRoute(fmt.Sprintf("/%s/%s", name, pk), resource)
	api.routes.addRoute(fmt.Sprintf("/%s/%s/", name, pk), resource)

	return nil
}

func (api *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Parse the API parameters and build the request object
	resource, params, _ := api.routes.getValue(r.URL.Path)
	if resource == nil {
		http.NotFound(w, r)
		return
	}

	request := &Request{Request: r, Params: params}

	var response Response
	var err *Error

	// If there are no parameters
	if len(params) == 0 {
		switch r.Method {
		case "GET":
			response, err = resource.List(request)
		case "POST":
			response, err = resource.Post(request)
		default:
			http.Error(w, fmt.Sprintf("unsupported method: %s", r.Method), 400)
			return
		}
	} else {
		switch r.Method {
		case "GET":
			response, err = resource.Get(request)
		case "PATCH":
			response, err = resource.Patch(request)
		case "DELETE":
			response, err = resource.Delete(request)
		default:
			http.Error(w, fmt.Sprintf("unsupported method: %s", r.Method), 400)
			return
		}
	}
	if err != nil {
		http.Error(w, err.Error(), err.Code())
		return
	}
	if response == nil {
		// Set 204 no content
		w.WriteHeader(204)
		return
	}
	// Always set the media type
	encoder := resource.Encoder()
	w.Header().Set("Content-Type", encoder.MediaType())
	w.Write(encoder.Encode(response))
}

func New() *API {
	api := &API{
		resources: make(map[string]Resource),
		routes:    &node{},
	}
	return api
}
