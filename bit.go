// Copyright 2017 Igor Dolzhikov. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package bit

import "net/http"

// Control interface contains methods that control
// URL/POST/JSON query parameters, handle request/response
// and accelerate access to HTTP `Status Code`, `Body`.
type Control interface {
	// Request returns *http.Request
	Request() *http.Request

	// Query searches URL/Post query parameters by key.
	// If there are no values associated with the key, an empty string is returned.
	Query(key string) string

	// Param sets URL/Post key/value query parameters.
	Param(key, value string)

	// Code sets HTTP status code e.g. http.StatusOk
	Code(code int)

	// GetCode shows HTTP status code that set by Code()
	GetCode() int

	// Body writes prepared header, status code and body data into http output.
	// It is equal to using sequence of http.ResponseWriter methods:
	// WriteHeader(code int) and Write(b []byte) int, error
	Body(data interface{})

	// Embedded response writer
	http.ResponseWriter

	// TODO Add more control methods.
}

// Param is a single URL parameter, consisting of a key and a value.
type Param struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

// Router interface contains base http methods e.g. GET, PUT, POST
// and allows to assign user defined handlers in regular use cases
// like `Page not found`, `Method is not allowed`,
// Recovery from panic case, middleware, etc.
type Router interface {
	// Standard methods

	// GET registers a new request handle for HTTP GET method.
	GET(path string, f func(Control))
	// PUT registers a new request handle for HTTP PUT method.
	PUT(path string, f func(Control))
	// POST registers a new request handle for HTTP POST method.
	POST(path string, f func(Control))
	// DELETE registers a new request handle for HTTP DELETE method.
	DELETE(path string, f func(Control))
	// HEAD registers a new request handle for HTTP HEAD method.
	HEAD(path string, f func(Control))
	// OPTIONS registers a new request handle for HTTP OPTIONS method.
	OPTIONS(path string, f func(Control))
	// PATCH registers a new request handle for HTTP PATCH method.
	PATCH(path string, f func(Control))

	// Handler supports usage of the Router as a regular http Handler.
	http.Handler

	// User defined options and handlers

	// If enabled, the router automatically replies to OPTIONS requests.
	// Nevertheless OPTIONS handlers take priority over automatic replies.
	// By default this option is disabled
	UseOptionsReplies(bool)

	// SetupNotAllowedHandler defines own handler which is called when a request
	// cannot be routed.
	SetupNotAllowedHandler(func(Control))

	// SetupNotFoundHandler allows to define own handler for undefined URL path.
	// If it is not set, http.NotFound is used.
	SetupNotFoundHandler(func(Control))

	// SetupRecoveryHandler allows to define handler that called when panic happen.
	// The handler prevents your server from crashing and should be used to return
	// http status code http.StatusInternalServerError (500)
	SetupRecoveryHandler(func(Control))

	// SetupMiddleware defines handler that is allowed to take control
	// before it is called standard methods above e.g. GET, PUT.
	SetupMiddleware(func(func(Control)) func(Control))

	// Listen and serve on requested host and port e.g "0.0.0.0:8080"
	Listen(hostPort string) error

	// Lookup allows the manual lookup of a method + path combo.
	// This is e.g. useful to build a framework around this router. If the path was found, it
	// returns the handle function and the path parameter values.
	// Otherwise the third return value indicates whether a redirection to the same path
	// with an extra / without the trailing slash should be performed.
	Lookup(method, path string) (func(Control), []Param, bool)
}
