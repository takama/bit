# Bit Router

[![Build Status](https://travis-ci.org/takama/bit.svg?branch=master)](https://travis-ci.org/takama/bit)
[![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/takama/bit/issues)
[![Go Report Card](https://goreportcard.com/badge/github.com/takama/bit)](https://goreportcard.com/report/github.com/takama/bit)
[![codecov](https://codecov.io/gh/takama/bit/branch/master/graph/badge.svg)](https://codecov.io/gh/takama/bit)

A simplest HTTP router contains Router interface that compatible with other routers. As well additional Control interface embeds standard http.ResponseWriter and has methods that accelerate access to `Status Code`, `Body`, `URL/Post/JSON` parameters. This router is useful to prepare a RESTful API. Also it is able to prepare JSON output, which bind automatically for relevant types of data.

## Router interface

Router interface contains base http methods e.g. GET, PUT, POST and allows to assign user defined handlers in regular use cases like `Page not found`, `Method is not allowed`, Recovery from panic case, middleware, etc.

```go
type Router interface {
    GET(path string, f func(Control))
    PUT(path string, f func(Control))
    POST(path string, f func(Control))
    DELETE(path string, f func(Control))
    HEAD(path string, f func(Control))
    OPTIONS(path string, f func(Control))
    PATCH(path string, f func(Control))

    http.Handler

    UseOptionsReplies(bool)
    SetupNotAllowedHandler(func(Control))
    SetupNotFoundHandler(func(Control))
    SetupRecoveryHandler(func(Control))
    SetupMiddleware(func(func(Control)) func(Control))
    Listen(hostPort string) error
}
```

## Control interface

Control interface contains methods that control URL/POST/JSON query parameters, handle request/response and accelerate access to HTTP `Status Code`, `Body`.

```go
type Control interface {
    Request() *http.Request
    Query(key string) string
    Param(key, value string)
    Code(code int)
    GetCode() int
    Body(data interface{})

    http.ResponseWriter
}
```

### Examples

- Simplest example (listen static route):

```go
package main

import (
    "github.com/takama/bit"
)

func Hello(c bit.Control) {
    c.Body("Hello world")
}

func main() {
    r := bit.NewRouter()
    r.GET("/hello", Hello)

    // Listen and serve on 0.0.0.0:8080
    r.Listen(":8080")
}
```

Check it:

```sh
curl -i http://localhost:8080/hello/

HTTP/1.1 200 OK
Content-Type: text/plain; charset=utf-8
Date: Sun, 01 Oct 2017 17:33:50 GMT
Content-Length: 11

Hello world
```

- Listen dynamic route with parameter:

```go
package main

import (
    "github.com/takama/bit"
)

func main() {
    r := bit.NewRouter()
    r.GET("/hello/:name", func(c bit.Control) {
        c.Body("Hello " + c.Query(":name"))
    })

    // Listen and serve on 0.0.0.0:8080
    r.Listen(":8080")
}
```

Check it:

```sh
curl -i http://localhost:8080/hello/John

HTTP/1.1 200 OK
Content-Type: text/plain; charset=utf-8
Date: Sun, 01 Oct 2017 17:33:55 GMT
Content-Length: 10

Hello John
```

- Apply JSON `Content-Type` for all non-string types:

```go
package main

import (
    "github.com/takama/bit"
)

// Data is helper to construct JSON
type Data map[string]interface{}

func main() {
    r := bit.NewRouter()
    r.PUT("/api/v1/db/settings/name/:name/host/:host/port/:port", func(c bit.Control) {
        // Get parameters
        name := c.Query(":name")
        host := c.Query(":host")
        port := c.Query(":port")

        // Verify and save name, host, port
        // ...

        // Show new settings
        data := Data{
            "Database settings": Data{
                "name": name,
                "host": host,
                "port": port,
            },
        }
        c.Code(http.StatusAccepted)
        c.Body(data)
    })
    // Listen and serve on 0.0.0.0:8080
    r.Listen(":8080")
}
```

Check it:

```sh
curl -i -XPUT http://localhost:8080/api/v1/db/settings/name/test/host/localhost/port/3306

HTTP/1.1 201 OK
Content-Type: application/json
Date: Sun, 01 Oct 2017 17:33:56 GMT
Content-Length: 96

{
  "Database settings": {
    "name": "test",
    "host": "localhost",
    "port": "3306"
  }
}
```

## Contributing to the project

See the [contribution guidelines](docs/CONTRIBUTING.md) for information on how to
participate in the Bit router project by submitting pull requests or issues.

## License

[MIT Public License](https://github.com/takama/bit/blob/master/LICENSE)
