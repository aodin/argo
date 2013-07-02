package argonaut

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

/*
Methods
-------
### Elements
GET, PUT / POST, DELETE

### Collections
GET, POST
*/

var DuplicateCollection = errors.New("A collection with this name already exists")
var EmptyUrl = errors.New("API and Collections cannot have empty URLs")

type Endpoint struct {
	baseUrl   string
	registry  map[string]Collection
	resources map[string]string // map[Resource Name] URL
}

func (e *Endpoint) Route(w http.ResponseWriter, r *http.Request) {
	pathLength := len(e.baseUrl)

	// We don't need to match the start of the path to the baseUrl since the
	// multiplexer has already performed that operation.
	shortPath := r.URL.Path[pathLength:]
	path := strings.Split(shortPath, "/")

	// Path should always be at least length 1
	// Just be extra safe
	if len(path) == 0 || path[0] == "" {
		w.Header().Set("Content-Type", "application/json")
		// TODO Return a 500 error if nil?
		w.Write(e.AvailableResources())
		return
	}

	// Attempt to match the first path item to a Collection
	collection, exists := e.registry[path[0]]
	if !exists {
		// Do not write to response or it will return 200, not 404
		http.NotFound(w, r)
		return
	}

	// If an id has been given (path array index 1), use "detail" operations
	if len(path) > 1 && path[1] != "" {
		// TODO Use POST for full item update and PUT for partial?
		key := path[1]
		if r.Method == "POST" || r.Method == "PUT" {
			var body []byte
			body, readErr := ioutil.ReadAll(r.Body)
			// TODO Does it need to be closed?
			r.Body.Close()
			if readErr != nil {
				http.Error(w, readErr.Error(), 400)
				return
			}
			// TODO readErr should probably be a server error
			response, updateErr := collection.Update(key, body)
			if updateErr != nil {
				if updateErr == DoesNotExist {
					http.NotFound(w, r)
					return
				}
				http.Error(w, updateErr.Error(), 400)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(response)
			return
		}

		if r.Method == "DELETE" {
			// TODO Set the content-type?
			// Return status code 204: No Content
			if deleteErr := collection.Delete(key); deleteErr != nil {
				if deleteErr == DoesNotExist {
					http.NotFound(w, r)
					return
				}
				http.Error(w, deleteErr.Error(), 400)
				return
			}
			w.WriteHeader(204)
			return
		}

		// Assume GET method
		// TODO Check for other methods?
		response, readErr := collection.Read(key)
		if readErr != nil {
			if readErr == DoesNotExist {
				http.NotFound(w, r)
				return
			}
			http.Error(w, readErr.Error(), 400)
			return
		}

		// Set MIME-type
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
		return
	}

	// If no given was given in the URL, use "list" level operations
	// TODO Use POST for full item update and PUT for partial?
	if r.Method == "POST" || r.Method == "PUT" {
		var body []byte
		body, readErr := ioutil.ReadAll(r.Body)
		// TODO Does it need to be closed?
		r.Body.Close()
		if readErr != nil {
			http.Error(w, readErr.Error(), 400)
			return
		}
		// TODO readErr should probably be a server error
		// The created key is discard for now
		_, response, updateErr := collection.Create(body)
		if updateErr != nil {
			http.Error(w, updateErr.Error(), 400)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(collection.List())
}

func (e *Endpoint) AvailableResources() []byte {
	available, jsonErr := json.Marshal(e.resources)
	if jsonErr != nil {
		// TODO Log an error
		return nil
	}
	return available
}

// Meet the requirements of the http.Handler interface.
// Sole responsibility is to proxy to the approriate handler
func (e *Endpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.Route(w, r)
}

func (e *Endpoint) BaseURL() string {
	return e.baseUrl
}

func (e *Endpoint) Register(name string, c Collection) error {
	if _, exists := e.registry[name]; exists {
		return DuplicateCollection
	}
	e.registry[name] = c

	// TODO name should not end in a slash

	// Construct a URL for the new collection
	e.resources[name] = e.baseUrl + name + "/"

	return nil
}

// Create an endpoint and wire it to the given http context
// TODO Another way to handle requests, or not bind directly to http
func API(baseUrl string, collections ...Collection) (*Endpoint, error) {
	if len(baseUrl) == 0 {
		return nil, EmptyUrl
	}
	// If the base url must start and end with a slash "/"
	if baseUrl[len(baseUrl)-1] != '/' {
		baseUrl += "/"
	}
	if baseUrl[0] != '/' {
		baseUrl = "/" + baseUrl
	}

	e := &Endpoint{
		baseUrl:   baseUrl,
		registry:  make(map[string]Collection),
		resources: make(map[string]string),
	}

	return e, nil
}
