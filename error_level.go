package suplog

import (
	"github.com/sirupsen/logrus"
)

type errLvlCtxKey struct{}

type errLevelHook struct{}

func (h *errLevelHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *errLevelHook) Fire(e *logrus.Entry) error {
	if e == nil {
		return nil
	}
	lvl, ok := e.Context.Value(errLvlCtxKey{}).(Level)
	if !ok {
		return nil
	}
	v := e.Data[logrus.ErrorKey]
	if v == nil {
		return nil
	}
	err, ok := v.(error)
	if !ok {
		return nil
	}
	if err == nil {
		return nil
	}

	// there is an error, set level accordingly if
	// level is lower (panic = 0 ... debug = 5)
	//
	if lvl < e.Level {
		e.Level = lvl
	}
	return nil

}
