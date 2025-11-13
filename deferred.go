package suplog

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const deferredFieldKey = "::deferred::"

type deferredHook struct{}

func (h *deferredHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *deferredHook) Fire(e *logrus.Entry) error {
	if e == nil {
		return nil
	}
	for k, v := range e.Data {
		if !strings.HasPrefix(k, deferredFieldKey) {
			continue
		}

		// remove deferred field entry
		delete(e.Data, k)

		key := strings.TrimPrefix(k, deferredFieldKey)
		var value interface{}
		switch x := v.(type) {
		case nil:
			// nil interface, skip
			continue
		case *error:
			if x == nil || *x == nil {
				// do not write nil errors
				continue
			}
			value = *x
		case *string:
			if x == nil {
				continue
			}
			value = *x
		case *bool:
			if x == nil {
				continue
			}
			value = *x
		case *int:
			if x == nil {
				continue
			}
			value = *x
		case *int8:
			if x == nil {
				continue
			}
			value = *x
		case *int16:
			if x == nil {
				continue
			}
			value = *x
		case *int32:
			if x == nil {
				continue
			}
			value = *x
		case *int64:
			if x == nil {
				continue
			}
			value = *x
		case *uint:
			if x == nil {
				continue
			}
			value = *x
		case *uint8:
			if x == nil {
				continue
			}
			value = *x
		case *uint16:
			if x == nil {
				continue
			}
			value = *x
		case *uint32:
			if x == nil {
				continue
			}
			value = *x
		case *uint64:
			if x == nil {
				continue
			}
			value = *x
		case *float32:
			if x == nil {
				continue
			}
			value = *x
		case *float64:
			if x == nil {
				continue
			}
			value = *x
		case *time.Duration:
			if x == nil {
				continue
			}
			value = *x
		case *complex64:
			if x == nil {
				continue
			}
			value = *x
		case *complex128:
			if x == nil {
				continue
			}
			value = *x
		default:
			value = fmt.Sprintf("<unsupported %T>", v)
		}
		// overwrite with dereferenced value
		e.Data[key] = value
	}
	return nil
}
