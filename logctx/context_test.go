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
		return next(WithField(ctx, key, value))
	}
}

func fieldsMiddleware(key string, value interface{}, next handler) handler {
	return func(ctx context.Context) error {
		return next(WithFields(ctx, log.Fields{key: value}))
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

			// to add fields we MUST use the context helpers
			WithField(ctx, "pre", "true")

			defer func() {
				WithField(ctx, "defer", "true")

				// recover the modified logger from context
				Logger(ctx).
					WithError(err).
					Debug("chain executed")
			}()

			err = next(ctx)

			WithField(ctx, "post", "true")

			return err
		}
	}

	// execute the chain
	err := entryPoint(
		fieldMiddleware("one", 1, // adds a field via middleware
			fieldsMiddleware("two", 2, // adds another field via middleware
				func(ctx context.Context) error { // this is the end handler

					// this adds a field into the context logger
					WithField(ctx, "name", "John")

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
	assert.Contains(t, lines[1], "one=1")
	assert.Contains(t, lines[1], "two=2")
	assert.Contains(t, lines[1], "name=John")
	assert.Contains(t, lines[1], "level=debug")
	assert.Contains(t, lines[1], "msg=\"chain executed\"")
	assert.NotContains(t, lines[1], "bubble-up=false")
}
