package argo

import (
	"encoding/json"
	"fmt"
	"github.com/aodin/volta/config"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

type API struct {
	baseURL   string
	resources Resources
	router    *httprouter.Router
	config    config.Config
}

// Resources lists the available resources
func (a *API) Resources(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.MarshalIndent(a.resources.URLs(a.baseURL), "", "    ")
	w.Write(b)
}

// Get will GET the resource with the requested name
func (a *API) Get(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Determine which resource was requested
	name := ps.ByName("resource")
	resource, exists := a.resources[name]
	if !exists {
		http.NotFound(w, r)
		return
	}

	// Send the query parameters to the resource
	response := resource.Get(r.URL.Query())

	// Set the content type
	if response.ContentType != "" {
		w.Header().Set("Content-Type", response.ContentType)
	}

	// TODO Set status code

	b, _ := json.MarshalIndent(response, "", "    ")
	w.Write(b)
}

// Add will add the given resource at the given name
func (a *API) Add(name string, resource Resource) error {
	if _, exists := a.resources[name]; exists {
		return fmt.Errorf("Resource %s already exists", name)
	}
	a.resources[name] = resource
	return nil
}

func New(c config.Config, r *httprouter.Router, b string) *API {
	return &API{
		baseURL:   b,
		resources: make(Resources),
		router:    r,
		config:    c,
	}
}
