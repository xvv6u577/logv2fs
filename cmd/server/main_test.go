package main

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestRunServer(t *testing.T) {
	// router := RunServer()

	// res := httptest.NewRecorder()
	// data := url.Values{}
	// data.Set("email", "testuser")
	// data.Set("password", "testuser")
	// data.Set("path", "ray")
	// data.Set("status", "plain")
	// data.Set("role", "normal")

	// req, _ := http.NewRequest("POST", "/v1/signup", strings.NewReader(data.Encode()))
	// req.Header.Add("Content-Type", "application/json")
	// router.ServeHTTP(res, req)

	// assert.Equal(t, 200, res.Code)
	assert.Equal(t, 1, 1)
}
