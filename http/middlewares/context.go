package middlewares

import (
	"net/http"
	"time"

	"github.com/fchoquet/golibs/http/ctx"
)

// RouteGetterFunc is a function that extracts the route from an http requests
// It returns false as a second argument if no route is defined
type RouteGetterFunc func(r *http.Request) (string, bool)

// Timestamp injects the request time in the context
func Timestamp(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r.WithContext(ctx.WithRequestTime(r.Context(), time.Now())))
	})
}

// RouteName extracts the route name and injects it in the context
func RouteName(getRoute RouteGetterFunc) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if route, ok := getRoute(r); ok {
				r = r.WithContext(ctx.WithRouteName(r.Context(), route))
			}
			h.ServeHTTP(w, r)
		})
	}
}

// TransactionID extracts the transaction id from the URL and injects it in the context if it exists
// It does not return any error if it does not exist
func TransactionID(queryParam string) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ids := r.URL.Query()[queryParam]
			if ids != nil && len(ids) != 0 && ids[0] != "" {
				r = r.WithContext(ctx.WithTransactionID(r.Context(), ids[0]))
			}
			h.ServeHTTP(w, r)
		})
	}
}
