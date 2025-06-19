package suplog

import (
	"time"
)

type ConditionLogger interface {
	Logger
	Do(func(Logger))
}

// OnCondition returns a logger if the condition is true, otherwise returns NoOp logger.
// Only the first logger in the variadic argument is used, if provided.
func OnCondition(cond bool, logger ...Logger) ConditionLogger {
	if cond {
		return getLogger(logger...)
	}
	return NoOp
}

// OnErr returns a logger if the error is not nil, otherwise returns NoOp logger.
// Only the first logger in the variadic argument is used, if provided.
func OnErr(err error, logger ...Logger) ConditionLogger {
	if err == nil {
		return NoOp
	}
	return &DoerLogger{Logger: getLogger(logger...).WithError(err)}
}

// OnTime returns a logger if the tick channel has received a tick, otherwise returns NoOp logger.
// It checks the channel without blocking, so it will not wait for a tick.
func OnTime(tick <-chan time.Time, logger ...Logger) ConditionLogger {
	var triggered bool
	select {
	case <-tick:
		triggered = true
	default:
	}
	return OnCondition(triggered, logger...)
}

func getLogger(logger ...Logger) ConditionLogger {
	if len(logger) == 0 || logger[0] == nil {
		return DefaultDoerLogger
	}
	if l, ok := logger[0].(ConditionLogger); ok {
		return l
	}
	return &DoerLogger{Logger: logger[0]}
}

var DefaultDoerLogger = &DoerLogger{
	Logger: DefaultLogger,
}

type DoerLogger struct {
	Logger
}

func (d *DoerLogger) Do(fn func(Logger)) {
	logger := d.Logger
	if logger == nil {
		logger = DefaultLogger
	}
	fn(logger)
}
