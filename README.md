# assert

Simplified error handling for http routes using assert with status code.

[![Build Status][drone]](https://ci.rkusa.st/rkgo/assert)
[![GoDoc][godoc]](https://godoc.org/github.com/rkgo/assert)

### Example

Middleware usage

```go
app := web.New()
app.Use(assert.Middleware())
```

Asserting

```go
assert.OK(username != "", 400, "No username given")
assert.Error(err)
assert.Success(err, 400, "something failed")
```

[drone]: http://ci.rkusa.st/api/badges/rkgo/assert/status.svg?style=flat-square
[godoc]: http://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square