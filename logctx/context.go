package logctx

import (
	"context"
	"sync"

	"github.com/InjectiveLabs/suplog"
)

type ctxLogKey struct{}

type loggerCtx struct {
	shared    *sharedLogger
	modifiers []loggerModifier
}

type sharedLogger struct {
	logger suplog.Logger
	mx     sync.Mutex
}

type loggerModifier struct {
	fields suplog.Fields
	err    error
}

// WithLogger adds the logger to the context, wrapped in our thread-safe struct.
func WithLogger(ctx context.Context, logger suplog.Logger) context.Context {
	return context.WithValue(ctx, ctxLogKey{}, &loggerCtx{
		shared: &sharedLogger{
			logger: logger,
		},
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

// WithErr adds an error field to this context and contexts derived from it.
func WithErr(ctx context.Context, err error) context.Context {
	if err == nil {
		return ctx
	}

	return withModifier(ctx, loggerModifier{err: err})
}

// WithField adds a single field to this context and contexts derived from it.
func WithField(ctx context.Context, field string, value interface{}) context.Context {
	return WithFields(ctx, suplog.Fields{field: value})
}

// WithFields adds multiple fields to this context and contexts derived from it.
func WithFields(ctx context.Context, fields suplog.Fields) context.Context {
	return withModifier(ctx, loggerModifier{fields: copyFields(fields)})
}

// PromoteErr adds an error field to the shared request logger, making it visible
// to this context, its parents, siblings, and contexts derived from them.
func PromoteErr(ctx context.Context, err error) context.Context {
	if err == nil {
		return ctx
	}

	l, ok := fromContext(ctx)
	if !ok {
		return ctx
	}

	l.shared.mx.Lock()
	defer l.shared.mx.Unlock()

	l.shared.logger = l.shared.logger.WithError(err)

	return ctx
}

// PromoteField adds a single field to the shared request logger, making it
// visible to this context, its parents, siblings, and contexts derived from them.
func PromoteField(ctx context.Context, field string, value interface{}) context.Context {
	return PromoteFields(ctx, suplog.Fields{field: value})
}

// PromoteFields adds multiple fields to the shared request logger, making them
// visible to this context, its parents, siblings, and contexts derived from them.
func PromoteFields(ctx context.Context, fields suplog.Fields) context.Context {
	l, ok := fromContext(ctx)
	if !ok {
		return ctx
	}

	l.shared.mx.Lock()
	defer l.shared.mx.Unlock()

	l.shared.logger = l.shared.logger.WithFields(copyFields(fields))

	return ctx
}

// Logger retrieves the current logger for this context.
func Logger(ctx context.Context) suplog.Logger {
	return loggerFromContext(ctx)
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
	loggerFromContext(ctx).Log(level, msg)
}

// Logf logs a formatted message using the logger from the context at the specified level.
func Logf(ctx context.Context, level suplog.Level, format string, args ...interface{}) {
	loggerFromContext(ctx).Logf(level, format, args...)
}

func fromContext(ctx context.Context) (*loggerCtx, bool) {
	l, ok := ctx.Value(ctxLogKey{}).(*loggerCtx)
	return l, ok && l != nil && l.shared != nil
}

func withModifier(ctx context.Context, modifier loggerModifier) context.Context {
	l, ok := fromContext(ctx)
	if !ok {
		return ctx
	}

	next := l.clone()
	next.modifiers = append(next.modifiers, modifier)

	return context.WithValue(ctx, ctxLogKey{}, next)
}

func loggerFromContext(ctx context.Context) suplog.Logger {
	l, ok := fromContext(ctx)
	if !ok {
		return suplog.DefaultLogger
	}

	l.shared.mx.Lock()
	logger := l.shared.logger
	l.shared.mx.Unlock()

	for _, modifier := range l.modifiers {
		logger = modifier.apply(logger)
	}

	return logger
}

func (l *loggerCtx) clone() *loggerCtx {
	next := &loggerCtx{
		shared: l.shared,
	}
	if len(l.modifiers) > 0 {
		next.modifiers = append([]loggerModifier(nil), l.modifiers...)
	}

	return next
}

func (m loggerModifier) apply(logger suplog.Logger) suplog.Logger {
	if len(m.fields) > 0 {
		logger = logger.WithFields(m.fields)
	}
	if m.err != nil {
		logger = logger.WithError(m.err)
	}

	return logger
}

func copyFields(fields suplog.Fields) suplog.Fields {
	if len(fields) == 0 {
		return nil
	}

	out := make(suplog.Fields, len(fields))
	for key, value := range fields {
		out[key] = value
	}

	return out
}
