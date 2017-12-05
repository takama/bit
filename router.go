// Copyright 2017 Igor Dolzhikov. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package bit

import (
	"net/http"
	"strings"
)

type router struct {
	// List of handlers that associated with known http methods (GET, POST ...)
	handlers map[string]*parser

	// If enabled, the router automatically replies to OPTIONS requests.
	// Nevertheless OPTIONS handlers take priority over automatic replies.
	optionsRepliesEnabled bool

	// Configurable handler which is called when a request cannot be routed.
	notAllowed func(Control)

	// Configurable handler which is called when panic happen.
	recoveryHandler func(Control)

	// Configurable middleware which is allowed to take control
	// before registration of new handlers via GET, PUT, etc..
	registerMiddleware func(string, string, func(Control)) (string, string, func(Control))

	// Configurable handler which is allowed to take control
	// before it is called standard methods e.g. GET, PUT.
	middlewareHandler func(func(Control)) func(Control)

	// Configurable http.Handler which is called when URL path has not defined method.
	// If it is not set, http.NotFound is used.
	notFound func(Control)
}

// NewRouter returns new router that implement Router interface.
func NewRouter() Router {
	return &router{
		handlers: make(map[string]*parser),
	}
}

// GET registers a new request handle for HTTP GET method.
func (r *router) GET(path string, f func(Control)) {
	r.register("GET", path, f)
}

// PUT registers a new request handle for HTTP PUT method.
func (r *router) PUT(path string, f func(Control)) {
	r.register("PUT", path, f)
}

// POST registers a new request handle for HTTP POST method.
func (r *router) POST(path string, f func(Control)) {
	r.register("POST", path, f)
}

// DELETE registers a new request handle for HTTP DELETE method.
func (r *router) DELETE(path string, f func(Control)) {
	r.register("DELETE", path, f)
}

// HEAD registers a new request handle for HTTP HEAD method.
func (r *router) HEAD(path string, f func(Control)) {
	r.register("HEAD", path, f)
}

// OPTIONS registers a new request handle for HTTP OPTIONS method.
func (r *router) OPTIONS(path string, f func(Control)) {
	r.register("OPTIONS", path, f)
}

// PATCH registers a new request handle for HTTP PATCH method.
func (r *router) PATCH(path string, f func(Control)) {
	r.register("PATCH", path, f)
}

// If enabled, the router automatically replies to OPTIONS requests.
// Nevertheless OPTIONS handlers take priority over automatic replies.
// By default this option is disabled
func (r *router) UseOptionsReplies(enabled bool) {
	r.optionsRepliesEnabled = enabled
}

// SetupNotAllowedHandler defines own handler which is called when a request
// cannot be routed.
func (r *router) SetupNotAllowedHandler(f func(Control)) {
	r.notAllowed = f
}

// SetupNotFoundHandler allows to define own handler for undefined URL path.
// If it is not set, http.NotFound is used.
func (r *router) SetupNotFoundHandler(f func(Control)) {
	r.notFound = f
}

// SetupRecoveryHandler allows to define handler that called when panic happen.
// The handler prevents your server from crashing and should be used to return
// http status code http.StatusInternalServerError (500)
func (r *router) SetupRecoveryHandler(f func(Control)) {
	r.recoveryHandler = f
}

// SetupRegisterMiddleware allows to define a middleware that take place
// during registration of new handlers in the Router via methods GET, POST, etc..
//
// The middleware is inteded to be used for integration of the routing information
// to thirdparty systems.
func (r *router) SetupRegisterMiddleware(f func(string, string, func(Control)) (string, string, func(Control))) {
	r.registerMiddleware = f
}

// SetupMiddleware defines handler is allowed to take control
// before it is called standard methods e.g. GET, PUT.
func (r *router) SetupMiddleware(f func(func(Control)) func(Control)) {
	r.middlewareHandler = f
}

// Listen and serve on requested host and port
func (r *router) Listen(hostPort string) error {
	return http.ListenAndServe(hostPort, r)
}

// registers a new handler with the given path and method.
func (r *router) register(method, path string, f func(Control)) {
	if r.registerMiddleware != nil {
		method, path, f = r.registerMiddleware(method, path, f)
	}
	if r.handlers[method] == nil {
		r.handlers[method] = newParser()
	}
	r.handlers[method].register(path, f)
}

func (r *router) recovery(w http.ResponseWriter, req *http.Request) {
	if recv := recover(); recv != nil {
		c := NewControl(w, req)
		r.recoveryHandler(c)
	}
}

// AllowedMethods returns list of allowed methods
func (r *router) allowedMethods(path string) []string {
	var allowed []string
	for method, parser := range r.handlers {
		if _, _, ok := parser.get(path); ok {
			allowed = append(allowed, method)
		}
	}

	return allowed
}

// ServeHTTP implements http.Handler interface.
func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if r.recoveryHandler != nil {
		defer r.recovery(w, req)
	}
	if _, ok := r.handlers[req.Method]; ok {
		if handle, params, ok := r.handlers[req.Method].get(req.URL.Path); ok {
			c := NewControl(w, req)
			if len(params) > 0 {
				for _, item := range params {
					c.Param(item.Key, item.Value)
				}
			}
			if r.middlewareHandler != nil {
				r.middlewareHandler(handle)(c)
			} else {
				handle(c)
			}
			return
		}
	}
	allowed := r.allowedMethods(req.URL.Path)

	if len(allowed) == 0 {
		if r.notFound != nil {
			c := NewControl(w, req)
			r.notFound(c)
		} else {
			http.NotFound(w, req)
		}
		return
	}

	w.Header().Set("Allow", strings.Join(allowed, ", "))
	if req.Method == "OPTIONS" && r.optionsRepliesEnabled {
		return
	}
	if r.notAllowed != nil {
		c := NewControl(w, req)
		r.notAllowed(c)
	} else {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

// Lookup allows the manual lookup of a method + path combo.
func (r *router) Lookup(method, path string) (func(Control), []Param, bool) {
	if root := r.handlers[method]; root != nil {
		return root.get(path)
	}
	return nil, nil, false
}
