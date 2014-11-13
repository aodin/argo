package argo

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockResource struct{}

func (m mockResource) List(*Request) (Response, *APIError) {
	return nil, nil
}

func (m mockResource) Post(*Request) (Response, *APIError) {
	return nil, nil
}

func (m mockResource) Get(*Request) (Response, *APIError) {
	return nil, nil
}

func (m mockResource) Patch(*Request) (Response, *APIError) {
	return nil, nil
}

func (m mockResource) Delete(*Request) (Response, *APIError) {
	return nil, nil
}

func TestAPI(t *testing.T) {
	assert := assert.New(t)

	// Start a server with no resources
	api := New()
	ts := httptest.NewServer(api)
	defer ts.Close()

	// API root returns routes
	resp, err := http.Get(ts.URL)
	assert.Nil(err)
	assert.Equal(http.StatusOK, resp.StatusCode)

	// A resource can be added after the server is started
	api.AddRest("nothing", mockResource{})

	resp, err = http.Get(ts.URL + "/nothing")
	assert.Nil(err)
	assert.Equal(http.StatusNoContent, resp.StatusCode)

	// Since no parameters were given, detail requests will 404
	resp, err = http.Get(ts.URL + "/nothing/1")
	assert.Nil(err)
	assert.Equal(http.StatusNotFound, resp.StatusCode)

	// Add a resource with parameters
	api.AddRest("things", mockResource{}, "id")
	resp, err = http.Get(ts.URL + "/things")
	assert.Nil(err)
	assert.Equal(http.StatusNoContent, resp.StatusCode)

	resp, err = http.Get(ts.URL + "/things/1")
	assert.Nil(err)
	assert.Equal(http.StatusNoContent, resp.StatusCode)
}
