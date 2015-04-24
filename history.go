package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

var lastId int

func freezeRequest(r *http.Request) error {
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Body); err != nil {
		return fmt.Errorf("unable to clone request: error reading original request body: %s", err)
	}

	if err := r.Body.Close(); err != nil {
		return fmt.Errorf("unable to clone request: cannot close original request body: %s", err)
	}
	r.Body = ioutil.NopCloser(&buf)
	return nil
}

var requestHistory []http.Request
