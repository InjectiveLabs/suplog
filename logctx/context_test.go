package logctx

import (
	"context"
	"errors"
	"strings"
	"testing"

	log "github.com/InjectiveLabs/suplog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type handler func(ctx context.Context) error

func fieldMiddleware(key string, value interface{}, next handler) handler {
	return func(ctx context.Context) error {
		ctx = WithField(ctx, key, value)
		return next(ctx)
	}
}

func fieldsMiddleware(key string, value interface{}, next handler) handler {
	return func(ctx context.Context) error {
		ctx = WithFields(ctx, log.Fields{key: value})
		return next(ctx)
	}
}

func TestContextLogging(t *testing.T) {
	var recorder strings.Builder
	entryPoint := func(next handler) handler {
		return func(ctx context.Context) (err error) {
			l := log.NewLogger(&recorder, new(log.TextFormatter))

			// set the logger in the context, must be done
			// before any other modifications
			ctx = WithLogger(ctx, l)

			// scoped fields apply to this context and contexts derived from it.
			ctx = WithField(ctx, "pre", "true")

			defer func() {
				ctx = WithField(ctx, "defer", "true")
				ctx = WithErr(ctx, err)

				// recover the logger for this scope.
				Logger(ctx).Debug("chain executed")
			}()

			err = next(ctx)

			ctx = WithField(ctx, "post", "true")

			return err
		}
	}

	// execute the chain
	err := entryPoint(
		fieldMiddleware("one", 1, // adds a field via middleware
			fieldsMiddleware("two", 2, // adds another field via middleware
				func(ctx context.Context) error { // this is the end handler

					// this adds a field into the context logger
					ctx = WithField(ctx, "name", "John")

					// this log will contain all prior fields but
					// will NOT bubble up the "buble-up" field
					Logger(ctx).
						WithField("bubble-up", false).
						Info("handler called")

					return errors.New("oops")
				},
			),
		),
	)(context.Background())

	require.ErrorContains(t, err, "oops", "error from handler")

	lines := strings.Split(recorder.String(), "\n")
	require.Len(t, lines, 3) // 2 logs + final newline

	// first log is from the handler
	assert.Contains(t, lines[0], "bubble-up=false")
	assert.Contains(t, lines[0], "pre=true")
	assert.Contains(t, lines[0], "one=1")
	assert.Contains(t, lines[0], "two=2")
	assert.Contains(t, lines[0], "name=John")
	assert.Contains(t, lines[0], "level=info")
	assert.NotContains(t, lines[0], "post=")
	assert.NotContains(t, lines[0], "defer=")
	assert.NotContains(t, lines[0], "error=")
	assert.Contains(t, lines[0], "msg=\"handler called\"")

	assert.Contains(t, lines[1], "error=oops")
	assert.Contains(t, lines[1], "pre=true")
	assert.Contains(t, lines[1], "post=true")
	assert.Contains(t, lines[1], "defer=true")
	assert.NotContains(t, lines[1], "one=")
	assert.NotContains(t, lines[1], "two=")
	assert.NotContains(t, lines[1], "name=")
	assert.Contains(t, lines[1], "level=debug")
	assert.Contains(t, lines[1], "msg=\"chain executed\"")
	assert.NotContains(t, lines[1], "bubble-up=false")
}

func TestPromotedFieldsSurviveReturnPath(t *testing.T) {
	var recorder strings.Builder
	logger := log.NewLogger(&recorder, new(log.TextFormatter))

	ctx := WithLogger(context.Background(), logger)
	ctx = WithField(ctx, "svc", "foo-api")

	grpcCtx := WithField(ctx, "method", "ListFoo")
	handlerCtx := PromoteField(grpcCtx, "handler", "ListFooHandler")

	dbCtx := WithField(handlerCtx, "db_service", "foo-db")
	Info(dbCtx, "db query")

	Logger(grpcCtx).Info("grpc finished")

	lines := strings.Split(strings.TrimSpace(recorder.String()), "\n")
	require.Len(t, lines, 2)

	assert.Contains(t, lines[0], "svc=foo-api")
	assert.Contains(t, lines[0], "method=ListFoo")
	assert.Contains(t, lines[0], "handler=ListFooHandler")
	assert.Contains(t, lines[0], "db_service=foo-db")
	assert.Contains(t, lines[0], `msg="db query"`)

	assert.Contains(t, lines[1], "svc=foo-api")
	assert.Contains(t, lines[1], "method=ListFoo")
	assert.Contains(t, lines[1], "handler=ListFooHandler")
	assert.NotContains(t, lines[1], "db_service=")
	assert.Contains(t, lines[1], `msg="grpc finished"`)
}

func TestScopedFieldsRequireReturnedContext(t *testing.T) {
	var recorder strings.Builder
	logger := log.NewLogger(&recorder, new(log.TextFormatter))
	ctx := WithLogger(context.Background(), logger)

	WithField(ctx, "ignored", true)
	Info(ctx, "without returned context")

	ctx = WithField(ctx, "kept", true)
	Info(ctx, "with returned context")

	lines := strings.Split(strings.TrimSpace(recorder.String()), "\n")
	require.Len(t, lines, 2)

	assert.NotContains(t, lines[0], "ignored=")
	assert.NotContains(t, lines[0], "kept=")
	assert.Contains(t, lines[1], "kept=true")
	assert.NotContains(t, lines[1], "ignored=")
}

func TestSiblingScopesAreIsolated(t *testing.T) {
	var recorder strings.Builder
	logger := log.NewLogger(&recorder, new(log.TextFormatter))
	ctx := WithLogger(context.Background(), logger)
	ctx = WithField(ctx, "request_id", "req-123")

	left := WithField(ctx, "branch", "left")
	right := WithField(ctx, "branch", "right")

	Info(left, "left branch")
	Info(right, "right branch")

	lines := strings.Split(strings.TrimSpace(recorder.String()), "\n")
	require.Len(t, lines, 2)

	assert.Contains(t, lines[0], "request_id=req-123")
	assert.Contains(t, lines[0], "branch=left")
	assert.NotContains(t, lines[0], "branch=right")

	assert.Contains(t, lines[1], "request_id=req-123")
	assert.Contains(t, lines[1], "branch=right")
	assert.NotContains(t, lines[1], "branch=left")
}

func TestPromotedFieldsAreVisibleToSiblings(t *testing.T) {
	var recorder strings.Builder
	logger := log.NewLogger(&recorder, new(log.TextFormatter))
	ctx := WithLogger(context.Background(), logger)

	left := WithField(ctx, "branch", "left")
	right := WithField(ctx, "branch", "right")

	PromoteFields(left, log.Fields{"handler": "ListFooHandler"})
	Info(right, "right branch")

	lines := strings.Split(strings.TrimSpace(recorder.String()), "\n")
	require.Len(t, lines, 1)

	assert.Contains(t, lines[0], "handler=ListFooHandler")
	assert.Contains(t, lines[0], "branch=right")
	assert.NotContains(t, lines[0], "branch=left")
}

func TestContextLogHelpers(t *testing.T) {
	t.Run("Log helper uses the contextual logger", func(t *testing.T) {
		var recorder strings.Builder
		logger := log.NewLogger(&recorder, new(log.TextFormatter))
		ctx := WithLogger(context.Background(), logger)
		ctx = WithField(ctx, "helper", "log")

		Info(ctx, "info helper message")

		lines := strings.Split(strings.TrimSpace(recorder.String()), "\n")
		require.Len(t, lines, 1)
		require.Contains(t, lines[0], "level=info")
		require.Contains(t, lines[0], "helper=log")
		require.Contains(t, lines[0], `msg="info helper message"`)
	})

	t.Run("Logf helper uses the contextual logger", func(t *testing.T) {
		var recorder strings.Builder
		logger := log.NewLogger(&recorder, new(log.TextFormatter))
		ctx := WithLogger(context.Background(), logger)
		ctx = WithField(ctx, "helper", "logf")

		Infof(ctx, "info helper %d", 1)

		lines := strings.Split(strings.TrimSpace(recorder.String()), "\n")
		require.Len(t, lines, 1)
		require.Contains(t, lines[0], "level=info")
		require.Contains(t, lines[0], "helper=logf")
		require.Contains(t, lines[0], `msg="info helper 1"`)
	})
}

func TestEnsureLogger(t *testing.T) {
	var recorder strings.Builder
	logger := log.NewLogger(&recorder, new(log.TextFormatter))
	ctx := context.Background()

	// Should set the logger since none exists
	ctxWithLogger := EnsureLogger(ctx, logger)
	require.NotNil(t, Logger(ctxWithLogger))
	require.Equal(t, logger, Logger(ctxWithLogger))

	// Should NOT overwrite existing logger
	anotherLogger := log.NewLogger(&strings.Builder{}, new(log.TextFormatter))
	ctxWithLogger2 := EnsureLogger(ctxWithLogger, anotherLogger)
	require.Equal(t, logger, Logger(ctxWithLogger2))
}
