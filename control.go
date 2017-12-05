// Copyright 2017 Igor Dolzhikov. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package bit

import (
	"compress/gzip"
	"encoding/json"
	"net/http"
	"strings"
)

type control struct {
	req    *http.Request
	w      http.ResponseWriter
	code   int
	params *Params
}

// NewControl returns new control that implement Control interface.
func NewControl(w http.ResponseWriter, req *http.Request) Control {
	params := make(Params, 0)
	return &control{
		req:    req,
		w:      w,
		params: &params,
	}
}

// Request returns *http.Request
func (c *control) Request() *http.Request {
	return c.req
}

// Response writer implementation
// Header represents http.ResponseWriter header, the key-value pairs in an HTTP header.
func (c *control) Header() http.Header {
	return c.w.Header()
}

// Write writes the data to the connection as part of an HTTP reply.
func (c *control) Write(b []byte) (int, error) {
	return c.w.Write(b)
}

// WriteHeader sends an HTTP response header with status code.
func (c *control) WriteHeader(code int) {
	c.w.WriteHeader(code)
}

// Params get embedded key/value data that contains URL/Post query parameters
func (c *control) Params() *Params {
	return c.params
}

// Query searches URL/Post value by key.
// If there are no values associated with the key, an empty string is returned.
func (c *control) Query(key string) string {
	if param, ok := c.params.Get(key); ok {
		return param
	}

	return c.req.URL.Query().Get(key)
}

// Code sets HTTP status code e.g. http.StatusOk
func (c *control) Code(code int) {
	if code >= 100 && code < 600 {
		c.code = code
	}
}

// GetCode shows HTTP status code that set by Code()
func (c *control) GetCode() int {
	return c.code
}

// Body writes prepared header, status code and body data into http output.
// It is equal to using sequence of http.ResponseWriter methods:
// WriteHeader(code int) and Write(b []byte) int, error
func (c *control) Body(data interface{}) {
	var content []byte

	if str, ok := data.(string); ok {
		content = []byte(str)
	} else {
		var err error
		content, err = json.Marshal(data)
		if err != nil {
			c.w.WriteHeader(http.StatusInternalServerError)
			c.w.Write([]byte(err.Error()))
			return
		}
		if c.w.Header().Get("Content-type") == "" {
			c.w.Header().Add("Content-type", "application/json")
		}
	}
	if strings.Contains(c.req.Header.Get("Accept-Encoding"), "gzip") {
		// Detect content type before encoding if it isn't defined
		if c.w.Header().Get("Content-type") == "" {
			c.w.Header().Set("Content-type", http.DetectContentType(content))
		}
		c.w.Header().Add("Content-Encoding", "gzip")
		if c.code > 0 {
			c.w.WriteHeader(c.code)
		}
		gz := gzip.NewWriter(c.w)
		gz.Write(content)
		gz.Close()
	} else {
		if c.code > 0 {
			c.w.WriteHeader(c.code)
		}
		c.w.Write(content)
	}
}
