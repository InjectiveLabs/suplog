package logctx

import (
	"context"
	"sync"

	"github.com/InjectiveLabs/suplog"
)

type ctxLogKey struct{}

type loggerCtx struct {
	logger suplog.Logger
	mx     sync.Mutex
}

// WithLogger adds the logger to the context, wrapped in our thread-safe struct.
func WithLogger(ctx context.Context, logger suplog.Logger) context.Context {
	return context.WithValue(ctx, ctxLogKey{}, &loggerCtx{logger: logger})
}

// WithErr adds an error field to the logger in the context (thread-safe).
func WithErr(ctx context.Context, err error) context.Context {
	l, ok := fromContext(ctx)
	if !ok {
		return ctx
	}

	l.mx.Lock()
	defer l.mx.Unlock()

	l.logger = l.logger.WithError(err)

	return ctx
}

// WithField adds a single field to the logger in the context (thread-safe).
func WithField(ctx context.Context, field string, value interface{}) context.Context {
	l, ok := fromContext(ctx)
	if !ok {
		return ctx
	}

	l.mx.Lock()
	defer l.mx.Unlock()

	l.logger = l.logger.WithField(field, value)

	return ctx
}

// WithFields adds multiple fields to the logger in the context (thread-safe).
func WithFields(ctx context.Context, fields suplog.Fields) context.Context {
	l, ok := fromContext(ctx)
	if !ok {
		return ctx
	}

	l.mx.Lock()
	defer l.mx.Unlock()

	l.logger = l.logger.WithFields(fields)

	return ctx
}

// Logger retrieves the *current* logger from the context.
func Logger(ctx context.Context) suplog.Logger {
	l, ok := fromContext(ctx)
	if !ok {
		return suplog.DefaultLogger
	}

	// We must also lock when reading to prevent a race with a concurrent write
	l.mx.Lock()
	defer l.mx.Unlock()

	return l.logger
}

func fromContext(ctx context.Context) (*loggerCtx, bool) {
	l, ok := ctx.Value(ctxLogKey{}).(*loggerCtx)
	return l, ok && l != nil
}
