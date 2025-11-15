package logctx

import (
	"context"
	"sync"

	"github.com/InjectiveLabs/suplog"
)

type ctxLogKey struct{}

type loggerCtx struct {
	logger suplog.Logger
	level  suplog.Level
	mx     sync.Mutex
}

// WithLogger adds the logger to the context, wrapped in our thread-safe struct.
func WithLogger(ctx context.Context, logger suplog.Logger) context.Context {
	return context.WithValue(ctx, ctxLogKey{}, &loggerCtx{
		logger: logger,
		level:  suplog.TraceLevel,
	})
}

// EnsureLogger sets the logger in the context only if it doesn't already exist.
// It's recommended to use this in middleware to avoid overwriting existing loggers.
func EnsureLogger(ctx context.Context, logger suplog.Logger) context.Context {
	if _, ok := fromContext(ctx); ok {
		return ctx
	}
	return WithLogger(ctx, logger)
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

// WithLevel sets the logging level for the logger in the context (thread-safe).
func WithLevel(ctx context.Context, level suplog.Level) context.Context {
	l, ok := fromContext(ctx)
	if !ok {
		return ctx
	}

	l.mx.Lock()
	defer l.mx.Unlock()

	l.level = level

	return ctx
}

// Logger retrieves the *current* logger from the context.
func Logger(ctx context.Context) suplog.Logger {
	l, ok := fromContext(ctx)
	if !ok {
		return suplog.DefaultLogger
	}

	l.mx.Lock()
	defer l.mx.Unlock()

	return l.logger
}

func Debug(ctx context.Context, msg string) {
	Log(ctx, suplog.DebugLevel, msg)
}

func Info(ctx context.Context, msg string) {
	Log(ctx, suplog.InfoLevel, msg)
}

func Warn(ctx context.Context, msg string) {
	Log(ctx, suplog.WarnLevel, msg)
}

func Error(ctx context.Context, msg string) {
	Log(ctx, suplog.ErrorLevel, msg)
}

func Fatal(ctx context.Context, msg string) {
	Log(ctx, suplog.FatalLevel, msg)
}

func Panic(ctx context.Context, msg string) {
	Log(ctx, suplog.PanicLevel, msg)
}

func Debugf(ctx context.Context, format string, args ...interface{}) {
	Logf(ctx, suplog.DebugLevel, format, args...)
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	Logf(ctx, suplog.InfoLevel, format, args...)
}

func Warnf(ctx context.Context, format string, args ...interface{}) {
	Logf(ctx, suplog.WarnLevel, format, args...)
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	Logf(ctx, suplog.ErrorLevel, format, args...)
}

func Fatalf(ctx context.Context, format string, args ...interface{}) {
	Logf(ctx, suplog.FatalLevel, format, args...)
}

// Log logs a message using the logger from the context at the specified level.
func Log(ctx context.Context, level suplog.Level, msg string) {
	l, ok := fromContext(ctx)
	if !ok {
		suplog.DefaultLogger.Log(level, msg)
		return
	}

	l.mx.Lock()
	defer l.mx.Unlock()

	// Enforce the context's log level if set and with higher severity
	if level > l.level {
		level = l.level
	}

	l.logger.Log(level, msg)
}

// Logf logs a formatted message using the logger from the context at the specified level.
func Logf(ctx context.Context, level suplog.Level, format string, args ...interface{}) {
	l, ok := fromContext(ctx)
	if !ok {
		suplog.DefaultLogger.Logf(level, format, args...)
		return
	}

	l.mx.Lock()
	defer l.mx.Unlock()

	// Enforce the context's log level if set and with higher severity
	if level > l.level {
		level = l.level
	}

	l.logger.Logf(level, format, args...)
}

func fromContext(ctx context.Context) (*loggerCtx, bool) {
	l, ok := ctx.Value(ctxLogKey{}).(*loggerCtx)
	return l, ok && l != nil
}
