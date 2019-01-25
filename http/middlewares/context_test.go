package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fchoquet/golibs/http/ctx"
	"github.com/gorilla/mux"
)

func TestTimestamp(t *testing.T) {
	req, _ := http.NewRequest("GET", "whatever", nil)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := ctx.RequestTime(r.Context())

		if !ok {
			t.Error("Request time not found in context")
			return
		}
	})

	Timestamp(testHandler).ServeHTTP(httptest.NewRecorder(), req)
}

func TestRouteName(t *testing.T) {
	req, _ := http.NewRequest("GET", "/test", nil)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		routeName, ok := ctx.RouteName(r.Context())

		if !ok {
			t.Error("route name not found in context")
			return
		}

		if routeName != "test_route" {
			t.Errorf("expected \"test_route\" - got %q", routeName)
		}

	})

	// For this test we need a real router
	r := mux.NewRouter()
	getterFunc := func(r *http.Request) (string, bool) {
		if route := mux.CurrentRoute(r); route != nil {
			return route.GetName(), true
		}
		return "", false
	}

	r.Handle("/test", RouteName(getterFunc)(testHandler)).Methods("GET").Name("test_route")

	r.ServeHTTP(httptest.NewRecorder(), req)
}

func TestTransactionID(t *testing.T) {

	t.Run("when transaction_id is not provided", func(t *testing.T) {
		queries := []string{
			"whatever",
			"whatever?transaction_id",
			"whatever?transaction_id=",
		}

		for _, query := range queries {
			req, _ := http.NewRequest("GET", query, nil)

			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if _, ok := ctx.TransactionID(r.Context()); ok {
					t.Error("No transaction id should be provided")
				}
			})

			recorder := httptest.NewRecorder()
			TransactionID("transation_id")(testHandler).ServeHTTP(recorder, req)

			if recorder.Code != 200 {
				t.Errorf("expected 200 - got %d", recorder.Code)
			}
		}
	})

	t.Run("when transaction_id is provided", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "whatever?transaction_id=123-ABC-456", nil)

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			transactionID, ok := ctx.TransactionID(r.Context())
			if !ok {
				t.Error("Transaction ID not in request context")
				return
			}

			if transactionID != "123-ABC-456" {
				t.Errorf("expected %s - got %s", "foo.com", "123-ABC-456")
			}
		})

		recorder := httptest.NewRecorder()
		TransactionID("transaction_id")(testHandler).ServeHTTP(recorder, req)

		if recorder.Code != 200 {
			t.Errorf("expected %d - got %d", 200, recorder.Code)
		}
	})
}
