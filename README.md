# assert

Simplified error handling for http routes using assert with status code that works well (but not exclusively) with [rkusa/web](https://github.com/rkusa/web).

[![Build Status][travis]](https://travis-ci.org/rkusa/http-assert.svg)
[![GoDoc][godoc]](https://godoc.org/github.com/rkusa/http-assert)

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

[travis]: https://img.shields.io/travis/rkusa/http-assert.svg?style=flat-square
[godoc]: http://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square
