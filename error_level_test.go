package suplog

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorLevel(t *testing.T) {
	t.Run("raise level based on errors", func(t *testing.T) {
		var recorder strings.Builder
		l := NewLogger(&recorder, new(TextFormatter))

		l.WithError(errors.New("fail")).
			ErrLevel(ErrorLevel).
			Debugf("deferred log level")

		out := recorder.String()
		require.Contains(t, out, "error=fail")
		require.Contains(t, out, "level=error")
	})

	t.Run("keeps level based if no error", func(t *testing.T) {
		var recorder strings.Builder
		l := NewLogger(&recorder, new(TextFormatter))

		var err error
		l.WithError(err).
			ErrLevel(ErrorLevel).
			Debugf("deferred log level")

		out := recorder.String()
		require.NotContains(t, out, "error=")
		require.Contains(t, out, "level=debug")
	})
}
