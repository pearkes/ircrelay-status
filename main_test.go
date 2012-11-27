package main

import (
	"fmt"
	"net/http"
	"testing"
)

// Help developers know why the tests might be failing.
func checks() {
	fmt.Println("WARNING: The server must be running for the tests to pass.")
}

func TestIndex(t *testing.T) {

	var route = "/"
	var url = "http://localhost:4242" + route

	resp, err := http.Get(url)
	if err != nil {
		checks()
		t.Errorf("Fail => %v", err, resp)
	}

}
