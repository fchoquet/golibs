package response

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Default response function. Feel free to override with your own implementation
var (
	Error   ErrorFunc   = jsonError
	Errors  ErrorsFunc  = jsonErrors
	Success SuccessFunc = jsonSuccess
)

// SuccessFunc generates a successs response
type SuccessFunc func(ctx context.Context, w http.ResponseWriter, resp interface{}, code int) error

// An ErrorFunc generates an error response
type ErrorFunc func(w http.ResponseWriter, msg string, code int)

// An ErrorsFunc generates an error response composed of several error messages
// This is useful for validation results
type ErrorsFunc func(w http.ResponseWriter, msgs []string, code int)

// errorResponse is the default error response
type errorResponse struct {
	Errors []errorLine `json:"errors"`
}

type errorLine struct {
	Type    string `json:"type,omitempty"`
	Message string `json:"message"`
	Code    int    `json:"code,omitempty"`
}

// jsonError generates a formatted error
// It is similar to http.Error except that it returns json
// You should return and stop the middleware chain after calling this function
// since nothing prevents from stacking json structures
func jsonError(w http.ResponseWriter, msg string, code int) {
	jsonErrors(w, []string{msg}, code)
}

// jsonErrors generates a formatted error with multiple messages
// You should return and stop the middleware chain after calling this function
// since nothing prevents from stacking json structures
func jsonErrors(w http.ResponseWriter, msgs []string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)

	resp := errorResponse{}

	for _, msg := range msgs {
		resp.Errors = append(resp.Errors, errorLine{
			Message: msg,
			Code:    code,
		})
	}

	j, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		// It should never happen, but not sure what we can do if it's the case.
		// An empty body seems acceptable
		return
	}

	fmt.Fprintln(w, string(j))
}

// jsonSuccess writes a JSON formatted output according to the passed context
func jsonSuccess(ctx context.Context, w http.ResponseWriter, resp interface{}, code int) error {
	j, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintln(w, string(j))
	return nil
}

// StatusAwareResponseWriter is a custom response writer that keeps track of status code
// This is useful for logging
type StatusAwareResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

// WriteHeader overrides the standard WriteHeader method to keep track of the status code
func (rw *StatusAwareResponseWriter) WriteHeader(code int) {
	rw.StatusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// WrapWriter upgrades the response writer to create a StatusAwareResponseWriter
func WrapWriter(rw http.ResponseWriter) *StatusAwareResponseWriter {
	return &StatusAwareResponseWriter{
		ResponseWriter: rw,
	}
}
