package ctx

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
)

// This package encapsulate the context values used by this application
// It allows access to these values in a type-safe way

type contextKey int

var (
	// requestTimeKey contains the time when the request was received
	requestTimeKey contextKey = 1

	// routeNameKey contains the request route if defined
	routeNameKey contextKey = 2

	// loggerKey contains a contextualized logger
	loggerKey contextKey = 3

	// transactionIDKey contains a unique ID used to track a specific transaction
	transactionIDKey contextKey = 4
)

// WithRequestTime returns a new context containing the request time
func WithRequestTime(ctx context.Context, t time.Time) context.Context {
	return context.WithValue(ctx, requestTimeKey, t)
}

// RequestTime returns the request time stored in the context
func RequestTime(ctx context.Context) (t time.Time, ok bool) {
	t, ok = ctx.Value(requestTimeKey).(time.Time)
	return
}

// WithRouteName returns a new context containing the route name
func WithRouteName(ctx context.Context, route string) context.Context {
	return context.WithValue(ctx, routeNameKey, route)
}

// RouteName returns the route stored in the context
func RouteName(ctx context.Context) (route string, ok bool) {
	route, ok = ctx.Value(routeNameKey).(string)
	return
}

// WithLogger returns a new context containing a contextualized logger
func WithLogger(ctx context.Context, logger log.FieldLogger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// Logger returns the logger stored in the context
func Logger(ctx context.Context) (logger log.FieldLogger, ok bool) {
	logger, ok = ctx.Value(loggerKey).(log.FieldLogger)
	return
}

// WithTransactionID returns a new context containing a transaction ID
func WithTransactionID(ctx context.Context, transactionID string) context.Context {
	return context.WithValue(ctx, transactionIDKey, transactionID)
}

// TransactionID returns the transaction ID stored in the context
func TransactionID(ctx context.Context) (transactionID string, ok bool) {
	transactionID, ok = ctx.Value(transactionIDKey).(string)
	return
}
