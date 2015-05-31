// Simplified error handling for http routes using assert with status code.
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
	"log"
	"net/http"
	"runtime"
	"strings"

	"github.com/rkgo/web"
)

type assertError struct {
	statusCode int
	message    string
}

func (err assertError) Error() string {
	return err.message
}

func (err assertError) stack() string {
	buf := make([]byte, 32)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			break
		}
		buf = make([]byte, len(buf)*2)
	}

	stack := string(buf)
	lines := strings.Split(stack, "\n")
	return strings.Join(append(lines[0:1], lines[5:]...), "\n")
}

func ok(condition bool, statusCode int, message string, args ...interface{}) error {
	if !condition {
		if len(message) == 0 {
			message = http.StatusText(statusCode)
		}

		return assertError{statusCode, fmt.Sprintf(message, args...)}
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
	if e := ok(err == nil, statusCode, message, args...); e != nil {
		panic(e)
	}
}

// Error throws and responds with an 500 Internal Server Error if the provided
// error exists.
func Error(err error) {
	if err != nil {
		panic(ok(false, http.StatusInternalServerError, err.Error()))
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
	if e := ok(err == nil, statusCode, message, args...); e != nil {
		a.throw(e)
	}
}

// Error throws and responds with an 500 Internal Server Error if the provided
// error exists.
func (a *assertEncapsulation) Error(err error) {
	if err != nil {
		a.throw(ok(false, http.StatusInternalServerError, err.Error()))
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

// This Middleware is required to properly handle the errors thrown using
// this assert package. It must be called before the asserts are used.
func Middleware(logger *log.Logger) web.Middleware {
	return func(ctx web.Context, next web.Next) {
		defer func() {
			err := recover()
			if err == nil {
				return
			}

			switch assert := err.(type) {
			case assertError:
				if assert.statusCode == http.StatusInternalServerError && logger != nil {
					logger.Printf("PANIC: %s\n%s", assert.Error(), assert.stack())
				}

				http.Error(ctx, assert.Error(), assert.statusCode)
			default:
				panic(err)
			}
		}()

		next(ctx)
	}
}
