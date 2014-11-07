package argo

import (
	"fmt"
	"net/http"
	"strings"
)

type method string

const (
	GET    method = "GET"
	POST   method = "POST"
	PUT    method = "PUT"
	PATCH  method = "PATCH"
	DELETE method = "DELETE"
)

type API struct {
	prefix    string
	resources map[string]Rest
	routes    *node
}

func (api *API) SetPrefix(prefix string) *API {
	// If empty, then set to slash
	if prefix == "" {
		prefix = "/"
	} else if prefix[0] != '/' {
		prefix = "/" + prefix
	}
	api.prefix = prefix
	return api
}

// Add adds the resource to the API using its name
func (api *API) Add(resource *ResourceSQL) error {
	name := resource.Name
	if _, exists := api.resources[name]; exists {
		return fmt.Errorf(
			"argo: resource %s already exists",
			name,
		)
	}
	api.resources[name] = resource

	// Build the routes from the primary key(s)
	keys := resource.table.PrimaryKey()
	pks := make([]string, len(keys))
	for i, key := range keys {
		pks[i] = fmt.Sprintf(":%s", key)
	}
	pk := strings.Join(pks, "/")

	// Also add the routes
	p := api.prefix
	api.routes.addRoute(fmt.Sprintf("%s%s", p, name), resource)
	api.routes.addRoute(fmt.Sprintf("%s%s/", p, name), resource)
	api.routes.addRoute(fmt.Sprintf("%s%s/%s", p, name, pk), resource)
	api.routes.addRoute(fmt.Sprintf("%s%s/%s/", p, name, pk), resource)
	return nil
}

func (api *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Parse the API parameters and build the request object
	resource, params, _ := api.routes.getValue(r.URL.Path)
	if resource == nil {
		http.NotFound(w, r)
		return
	}

	// Determine the encoder type
	// TODO this could be done with routes / headers / auth
	encoder := resource.Encoder()

	request := &Request{Request: r, Params: params}

	var response Response
	var err *APIError

	// If there are no parameters
	method := method(r.Method)
	if len(params) == 0 {
		switch method {
		case GET:
			response, err = resource.List(request)
		case POST:
			response, err = resource.Post(request)
		default:
			err = MetaError(
				400,
				"unsupported collection method: %s",
				method,
			)
		}
	} else {
		switch method {
		case GET:
			response, err = resource.Get(request)
		case PATCH:
			response, err = resource.Patch(request)
		case DELETE:
			response, err = resource.Delete(request)
		default:
			err = MetaError(
				400,
				"unsupported item method: %s",
				method,
			)
		}
	}
	if err != nil {
		err.Write(w, encoder)
		return
	}
	if response == nil {
		// Set 204 no content
		w.WriteHeader(204)
		return
	}
	// Always set the media type
	w.Header().Set("Content-Type", encoder.MediaType())
	w.Write(encoder.Encode(response))
}

func New() *API {
	api := &API{
		resources: make(map[string]Rest),
		routes:    &node{},
	}
	return api
}
