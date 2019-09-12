// Package assert simplifies error handling for http routes using assert with
// status code. It exposes a middleware that works well (but not exclusively)
// with [rkusa/web](https://github.com/rkusa/web).
//
// Middleware usage
//
//  app := web.New()
//  app.Use(assert.Middleware())
//
// Asserting
//
//  assert.OK(username != "", 400, "No username given")
//  assert.Error(err)
//  assert.Success(err, 400, "something failed")
//
package assert

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
)

type AssertError struct {
	statusCode int
	message    string
	Err        error
}

func (err AssertError) Error() string {
	return err.message
}

func (err AssertError) stack() string {
	stack := string(debug.Stack())
	lines := strings.Split(stack, "\n")
	return strings.Join(lines[9:], "\n")
}

// NewAssertError creates a new assertion error.
func NewAssertError(statusCode int, message string, args ...interface{}) error {
	if len(message) == 0 {
		message = http.StatusText(statusCode)
	}

	return AssertError{statusCode, fmt.Sprintf(message, args...), nil}
}

func ok(condition bool, statusCode int, message string, args ...interface{}) error {
	if !condition {
		return NewAssertError(statusCode, message, args...)
	}

	return nil
}

func ok_err(err error, statusCode int, message string, args ...interface{}) error {
	if err != nil {
		if len(message) == 0 {
			message = http.StatusText(statusCode)
		}

		return AssertError{statusCode, fmt.Sprintf(message, args...), err}
	}

	return nil
}

// Success throws with the given statusCode and message if the provided
// condition evaluates to false. If message is an empty string, the default
// status description is used.
func OK(condition bool, statusCode int, message string, args ...interface{}) {
	if err := ok(condition, statusCode, message, args...); err != nil {
		panic(err)
	}
}

// Success throws with the given statusCode and message if the provided error
// exists. If message is an empty string, the default status description is used.
func Success(err error, statusCode int, message string, args ...interface{}) {
	if e := ok_err(err, statusCode, message, args...); e != nil {
		panic(e)
	}
}

// Error throws and responds with an 500 Internal Server Error if the provided
// error exists.
func Error(err error) {
	if err != nil {
		if _, ok := err.(AssertError); ok {
			panic(err)
		} else {
			panic(ok_err(err, http.StatusInternalServerError, err.Error()))
		}
	}
}

// Build and directly throw an error using the provided status code and message.
func Throw(statusCode int, message string, args ...interface{}) {
	OK(false, statusCode, message, args...)
}

// Assert represents an encapsulation for the assertions to provide an OnError
// hook.
type Assert interface {
	OnError(func())
	OK(bool, int, string, ...interface{})
	Success(error, int, string, ...interface{})
	Throw(int, string, ...interface{})
	Error(error)
}

type assertEncapsulation struct {
	onError func()
}

func (a *assertEncapsulation) throw(err error) {
	if a.onError != nil {
		a.onError()
	}

	panic(err)
}

// Register a callback that will be called once a assertion throws.
func (a *assertEncapsulation) OnError(fn func()) {
	a.onError = fn
}

// Success throws with the given statusCode and message if the provided
// condition evaluates to false. If message is an empty string, the default
// status description is used.
func (a *assertEncapsulation) OK(condition bool, statusCode int, message string, args ...interface{}) {
	if err := ok(condition, statusCode, message, args...); err != nil {
		a.throw(err)
	}
}

// Success throws with the given statusCode and message if the provided error
// exists. If message is an empty string, the default status description is used.
func (a *assertEncapsulation) Success(err error, statusCode int, message string, args ...interface{}) {
	if e := ok_err(err, statusCode, message, args...); e != nil {
		a.throw(e)
	}
}

// Error throws and responds with an 500 Internal Server Error if the provided
// error exists.
func (a *assertEncapsulation) Error(err error) {
	if err != nil {
		a.throw(ok_err(err, http.StatusInternalServerError, err.Error()))
	}
}

// Build and directly throw an error using the provided status code and message.
func (a *assertEncapsulation) Throw(statusCode int, message string, args ...interface{}) {
	a.OK(false, statusCode, message, args...)
}

// Create a new assertion encapsulation.
func New() Assert {
	return &assertEncapsulation{nil}
}

type Logger interface {
	Printf(format string, v ...interface{})
}

// This Middleware is required to properly handle the errors thrown using
// this assert package. It must be called before the asserts are used.
func Middleware(l Logger) func(http.ResponseWriter, *http.Request, http.HandlerFunc) {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		defer func() {
			err := recover()
			if err == nil {
				return
			}

			switch assert := err.(type) {
			case AssertError:
				if assert.statusCode == http.StatusInternalServerError && l != nil {
					l.Printf("PANIC: %s\n%s", assert.Error(), assert.stack())
					http.Error(rw, http.StatusText(assert.statusCode), assert.statusCode)
				} else {
					http.Error(rw, assert.Error(), assert.statusCode)
				}
			default:
				panic(err)
			}
		}()

		next(rw, r)
	}
}
