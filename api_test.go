package argo

import (
	"github.com/aodin/volta/config"
	"github.com/julienschmidt/httprouter"
	"testing"
)

func TestAPI(t *testing.T) {
	c := config.Config{
		Port: 8008,
	}
	router := httprouter.New()
	api := New(c, router, "/v1/")
	err := api.Add("example", exampleResource{"hello", "goodbye"})
	if err != nil {
		t.Fatalf("Failed to add resource to API: %s", err)
	}
}
