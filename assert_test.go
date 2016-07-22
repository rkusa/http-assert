package assert

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rkusa/web"
)

func TestOK(t *testing.T) {
	// condition evaluates to true
	func() {
		defer func() {
			err := recover()
			if err != nil {
				t.Errorf("expected err to be nil, got %v", err)
			}
		}()

		OK(true, http.StatusNotFound, "")
	}()

	// condition evaluates to false - empty message provided
	func() {
		defer func() {
			err := recover().(error)

			if err == nil || err.Error() != "Not Found" {
				t.Errorf(`expected "Not Found" error, got %s`, err.Error())
			}
		}()

		OK(false, http.StatusNotFound, "")
	}()

	// condition evaluates to false - with provided message
	func() {
		defer func() {
			err := recover().(error)

			if err == nil || err.Error() != "Invalid input" {
				t.Errorf(`expected "Invalid input" error, got %s`, err.Error())
			}
		}()

		OK(false, http.StatusBadRequest, "Invalid input")
	}()
}

func TestSuccess(t *testing.T) {
	// no error
	func() {
		defer func() {
			err := recover()
			if err != nil {
				t.Errorf("expected err to be nil, got %v", err)
			}
		}()

		Success(nil, http.StatusNotFound, "")
	}()

	// error - empty message provided
	func() {
		defer func() {
			err := recover().(error)

			if err == nil || err.Error() != "Not Found" {
				t.Errorf(`expected "Not Found" error, got %s`, err.Error())
			}
		}()

		Success(fmt.Errorf("Fail"), http.StatusNotFound, "")
	}()

	// error - with provided message
	func() {
		defer func() {
			err := recover().(error)

			if err == nil || err.Error() != "Invalid input" {
				t.Errorf(`expected "Invalid input" error, got %s`, err.Error())
			}
		}()

		Success(fmt.Errorf("Fail"), http.StatusBadRequest, "Invalid input")
	}()
}

func TestError(t *testing.T) {
	// no error
	func() {
		defer func() {
			err := recover()
			if err != nil {
				t.Errorf("expected err to be nil, got %v", err)
			}
		}()

		Error(nil)
	}()

	// error
	func() {
		defer func() {
			err := recover().(error)
			if err == nil || err.Error() != "Fail" {
				t.Errorf(`expected "Fail" error, got %s`, err.Error())
			}

		}()

		Error(fmt.Errorf("Fail"))
	}()
}

func TestThrow(t *testing.T) {
	func() {
		defer func() {
			err := recover().(error)
			if err == nil || err.Error() != "Invalid input" {
				t.Errorf(`expected "Invalid input" error, got %s`, err.Error())
			}
		}()

		Throw(http.StatusBadRequest, "Invalid input")
	}()
}

func TestOnError(t *testing.T) {
	called := false
	defer func() {
		err := recover().(error)
		if err == nil || err.Error() != "Bad Request" {
			t.Errorf(`expected "Bad Request" error, got %s`, err.Error())
		}

		if !called {
			t.Errorf("OnError was not called")
		}
	}()

	as := New()
	as.OnError(func() {
		called = true
	})

	as.OK(false, http.StatusBadRequest, "")
}

func TestMiddleware(t *testing.T) {
	app := web.New()
	app.Use(Middleware(nil))
	app.Use(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		Error(fmt.Errorf("Fail"))
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, (*http.Request)(nil))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected %s, got %s", http.StatusText(http.StatusInternalServerError), http.StatusText(rec.Code))
	}

	if rec.Body.String() != "Fail\n" {
		t.Errorf(`Expected body of "Fail\n", got %s`, rec.Body.String())
	}
}
