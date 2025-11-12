package suplog

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDefer(t *testing.T) {
	t.Run("deferred scalars", func(t *testing.T) {
		var recorder strings.Builder
		l := NewLogger(&recorder, new(TextFormatter))

		done := make(chan struct{})
		var (
			name    string
			age     int
			weight  float64
			hide    bool
			elapsed time.Duration
			plane   complex128
		)

		go func() {
			defer close(done)
			defer l.Defer("name", &name).
				Defer("age", &age).
				Defer("weight", &weight).
				Defer("hide", &hide).
				Defer("elapsed", &elapsed).
				Defer("plane", &plane).
				Infof("deferred values")

			name = "Alice"
			age = 30
			weight = 65.5
			hide = true
			elapsed = 1 * time.Second
			plane = complex(3, 2)
		}()
		<-done

		out := recorder.String()
		require.Contains(t, out, "name=Alice")
		require.Contains(t, out, "age=30")
		require.Contains(t, out, "weight=65.5")
		require.Contains(t, out, "hide=true")
		require.Contains(t, out, "elapsed=1s")
		require.Contains(t, out, "plane=\"(3+2i)\"")
	})

	t.Run("skips deferred nil values", func(t *testing.T) {
		var recorder strings.Builder
		l := NewLogger(&recorder, new(TextFormatter))

		done := make(chan struct{})
		var (
			name    *string
			age     *int
			weight  *float64
			hide    *bool
			elapsed *time.Duration
			plane   *complex128
		)

		go func() {
			defer close(done)
			defer l.Defer("name", name).
				Defer("age", age).
				Defer("weight", weight).
				Defer("hide", hide).
				Defer("elapsed", elapsed).
				Defer("plane", plane).
				Infof("deferred nil values")
		}()
		<-done

		out := recorder.String()
		require.NotContains(t, out, "name=")
		require.NotContains(t, out, "age=")
		require.NotContains(t, out, "weight=")
		require.NotContains(t, out, "hide=")
		require.NotContains(t, out, "elapsed=")
		require.NotContains(t, out, "plane=")
	})

	t.Run("deferred error", func(t *testing.T) {
		var recorder strings.Builder
		l := NewLogger(&recorder, new(TextFormatter))

		done := make(chan struct{})
		var err error

		go func() {
			defer close(done)
			defer l.DeferError(&err).Warningf("deferred error should be set")
			err = errors.New("oopsie")
		}()
		<-done

		require.Contains(t, recorder.String(), "error=oopsie")
	})

	t.Run("skip nil errors", func(t *testing.T) {
		var recorder strings.Builder
		l := NewLogger(&recorder, new(TextFormatter))

		done := make(chan struct{})
		var err error

		go func() {
			defer close(done)
			defer l.DeferError(&err).Warningf("deferred error should not appear")
			// no err, remains nil
		}()
		<-done

		require.NotContains(t, recorder.String(), "error=")
	})

	t.Run("invalid types are unsupported", func(t *testing.T) {
		var recorder strings.Builder
		l := NewLogger(&recorder, new(TextFormatter))

		done := make(chan struct{})
		var ch chan int

		go func() {
			defer close(done)
			defer l.Defer("wrong", &ch).Warningf("deferred unsupported type")
		}()
		<-done

		require.Contains(t, recorder.String(), "wrong=\"<unsupported *chan int>\"")
	})

	t.Run("skips nil interfaces", func(t *testing.T) {
		var recorder strings.Builder
		l := NewLogger(&recorder, new(TextFormatter))

		done := make(chan struct{})
		var i interface{}

		go func() {
			defer close(done)
			defer l.Defer("iface", i).Warningf("skip nil interface")
		}()
		<-done

		require.NotContains(t, recorder.String(), "unsupported")
		require.NotContains(t, recorder.String(), "iface")
	})

	t.Run("raise level based on deferred error", func(t *testing.T) {
		var recorder strings.Builder
		l := NewLogger(&recorder, new(TextFormatter))

		done := make(chan struct{})
		var err error

		go func() {
			defer close(done)
			defer l.DeferError(&err).ErrLevel(WarnLevel).Debugf("deferred log level")
			err = errors.New("crash")
		}()
		<-done

		out := recorder.String()
		require.Contains(t, out, "error=crash")
		require.Contains(t, out, "warning")
	})

	t.Run("keeps original level if no error", func(t *testing.T) {
		var recorder strings.Builder
		l := NewLogger(&recorder, new(TextFormatter))

		done := make(chan struct{})
		var err error

		go func() {
			defer close(done)
			defer l.DeferError(&err).ErrLevel(WarnLevel).Debugf("deferred log level")
			err = nil // no error
		}()
		<-done

		out := recorder.String()
		require.NotContains(t, out, "error=")
		require.Contains(t, out, "debug")
	})
}
