package suplog

import (
	"context"
	"time"
)

var (
	_ ConditionLogger = (*NoOpLogger)(nil)
	_ Logger          = (*NoOpLogger)(nil)
)

var NoOp = NoOpLogger{}

type NoOpLogger struct{}

func (n NoOpLogger) Do(fn func(Logger)) {
	// do... nothing!
}

func (n NoOpLogger) Success(format string, args ...interface{}) {

}

func (n NoOpLogger) Warning(format string, args ...interface{}) {

}

func (n NoOpLogger) Error(format string, args ...interface{}) {

}

func (n NoOpLogger) Debug(format string, args ...interface{}) {

}

func (n NoOpLogger) WithField(key string, value interface{}) Logger {
	return n
}

func (n NoOpLogger) WithFields(fields Fields) Logger {
	return n
}

func (n NoOpLogger) WithError(err error) Logger {
	return n
}

func (n NoOpLogger) WithContext(ctx context.Context) Logger {
	return n
}

func (n NoOpLogger) WithTime(t time.Time) Logger {
	return n
}

func (n NoOpLogger) Logf(level Level, format string, args ...interface{}) {

}

func (n NoOpLogger) Tracef(format string, args ...interface{}) {

}

func (n NoOpLogger) Debugf(format string, args ...interface{}) {

}

func (n NoOpLogger) Infof(format string, args ...interface{}) {

}

func (n NoOpLogger) Printf(format string, args ...interface{}) {

}

func (n NoOpLogger) Warningf(format string, args ...interface{}) {

}

func (n NoOpLogger) Errorf(format string, args ...interface{}) {

}

func (n NoOpLogger) Fatalf(format string, args ...interface{}) {

}

func (n NoOpLogger) Panicf(format string, args ...interface{}) {

}

func (n NoOpLogger) Log(level Level, args ...interface{}) {

}

func (n NoOpLogger) Trace(args ...interface{}) {

}

func (n NoOpLogger) Info(args ...interface{}) {

}

func (n NoOpLogger) Print(args ...interface{}) {

}

func (n NoOpLogger) Fatal(args ...interface{}) {

}

func (n NoOpLogger) Panic(args ...interface{}) {

}

func (n NoOpLogger) Logln(level Level, args ...interface{}) {

}

func (n NoOpLogger) Traceln(args ...interface{}) {

}

func (n NoOpLogger) Debugln(args ...interface{}) {

}

func (n NoOpLogger) Infoln(args ...interface{}) {

}

func (n NoOpLogger) Println(args ...interface{}) {

}

func (n NoOpLogger) Warningln(args ...interface{}) {

}

func (n NoOpLogger) Errorln(args ...interface{}) {

}

func (n NoOpLogger) Fatalln(args ...interface{}) {

}

func (n NoOpLogger) Panicln(args ...interface{}) {

}
