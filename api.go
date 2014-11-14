package argo

import (
	"fmt"
	"net/http"
	"strings"

	sql "github.com/aodin/aspect"
)

type method string

const (
	GET     method = "GET"
	POST    method = "POST"
	PUT     method = "PUT"
	PATCH   method = "PATCH"
	DELETE  method = "DELETE"
	OPTIONS method = "OPTIONS"
)

type API struct {
	prefix    string
	resources map[string]Rest
	routes    *node
	conn      sql.Connection
}

func (api *API) Prefix() string {
	return api.prefix
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

// Add adds the SQL resource to the API using its name
func (api *API) Add(resource *ResourceSQL) error {
	// Set the connection
	resource.conn = api.conn
	// Build the routes from the primary key(s)
	return api.AddRest(resource.Name, resource, resource.table.PrimaryKey()...)
}

// AddRest adds the Rest-ful resource to the API
func (api *API) AddRest(name string, resource Rest, keys ...string) error {
	if _, exists := api.resources[name]; exists {
		return fmt.Errorf(
			"argo: a resource named '%s' already exists",
			name,
		)
	}
	api.resources[name] = resource

	// TODO The prefix should be left out of the routing - it adds overhead
	p := api.prefix
	api.routes.addRoute(fmt.Sprintf("%s%s", p, name), resource)
	api.routes.addRoute(fmt.Sprintf("%s%s/", p, name), resource)

	if len(keys) > 0 {
		pks := make([]string, len(keys))
		for i, key := range keys {
			pks[i] = fmt.Sprintf(":%s", key)
		}
		pk := strings.Join(pks, "/")
		api.routes.addRoute(fmt.Sprintf("%s%s/%s", p, name, pk), resource)
		api.routes.addRoute(fmt.Sprintf("%s%s/%s/", p, name, pk), resource)
	}
	return nil
}

// Handle makes an API implement a handler with an argo Request instance
func (api *API) Handle(w http.ResponseWriter, request *Request) {
	// Publish the list of resources at root
	if request.URL.Path == api.prefix {
		// TODO alphabetical?
		response := make(map[string]string)
		for name, _ := range api.resources {
			// TODO base url? link?
			response[name] = fmt.Sprintf("%s%s", api.prefix, name)
		}
		w.Header().Set("Content-Type", request.Encoding.MediaType())
		w.Write(request.Encoding.Encode(response))
		return
	}

	// Parse the API parameters and build the request object
	resource, params, _ := api.routes.getValue(request.URL.Path)
	if resource == nil {
		http.NotFound(w, request.Request)
		return
	}

	// Build the new argo request instance
	// GetEncoder and GetDecoder should live in the argo Request constructor
	request.Params = params

	var response Response
	var err *APIError

	// If there are no parameters
	method := method(request.Method)
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
		err.Write(w, request.Encoding)
		return
	}
	if response == nil {
		w.WriteHeader(http.StatusNoContent) // 204
		return
	}
	// Always set the media type
	w.Header().Set("Content-Type", request.Encoding.MediaType())
	w.Write(request.Encoding.Encode(response))
}

// ServeHTTP implements the http Handler interface for APIs
func (api *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO request constructor function?
	request := &Request{
		Request:  r,
		Encoding: GetEncoder(r),
		Decoding: GetDecoder(r),
	}
	api.Handle(w, request)
}

// New creates a new API at root '/'
func New() *API {
	return &API{
		prefix:    "/",
		resources: make(map[string]Rest),
		routes:    &node{},
	}
}

func NewSQL(conn sql.Connection) *API {
	api := New()
	api.conn = conn
	return api
}
