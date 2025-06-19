package suplog

import (
	"errors"
	"fmt"
	"math"
	"testing"
	"time"
)

func Benchmark_OnConditionComplexMessage(b *testing.B) {
	start := time.Now()
	dummyErr := errors.New("dummy error")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		OnCondition(false).
			WithError(dummyErr).
			WithField("simple", "hi").
			WithField("some time", time.Since(start).String()).
			WithField("some print", fmt.Sprintf("%d", i)).
			Warningln("This is a warning message for benchmark testing")
	}
}

func Benchmark_OnConditionSimpleMessage(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		OnCondition(false).
			WithField("simple", "hi").
			Warningln("This is a warning message for benchmark testing")
	}
}

func Benchmark_OnConditionDo(b *testing.B) {
	start := time.Now()
	dummyErr := errors.New("dummy error")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		OnCondition(false).Do(func(l Logger) {
			b.Errorf("should not be called, condition is false")
			l.WithError(dummyErr).
				WithField("simple", "hi").
				WithField("some math", math.Trunc(float64(i)*0.5*1000)/1000).
				WithField("some time", time.Since(start).String()).
				WithField("some print", fmt.Sprintf("%d", i)).
				Warningln("This is a warning message for benchmark testing")
		})
	}
}
