package main

import (
	"net/http"
	"testing"
	"time"
)

func TestSimpleAPIServer(t *testing.T) {
	const host = "http://localhost:9000"
	go main()
	time.Sleep(time.Second * 3) // Giving some time to start the server

	testcases := []struct {
		path       string
		statusCode int
		body       string
	}{
		{"/", 404, ""},
		{"/hello", 200, "Hello World!"},
		{"/hello?name=gofr", 200, "Hello gofr!"},
		{"/error", 500, ""},
		{"/redis", 200, ""},
		{"/mysql", 200, ""},
		{"/trace", 200, ""},
		{"/.well-known/health", 200, ""}, // Health check should be added by the framework.
		{"/favicon.ico", 200, ""}, //Favicon should be added by the framework.
	}

	for _, tc := range testcases {
		req, _ := http.NewRequest("GET", host+tc.path, nil)
		c := http.Client{}
		resp, err := c.Do(req)
		if err != nil {
			t.Error("Could not get response", err)
		}

		if resp != nil && resp.StatusCode != tc.statusCode {
			t.Errorf("Failed. \t Expected %v\t Got %v", tc.statusCode, resp.StatusCode)
		}
	}
}
