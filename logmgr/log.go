package logmgr

import (
	"io"
	"sync"

	"github.com/Sirupsen/logrus"
)

var (
	loggers     []*logrus.Logger
	loggersLock sync.Mutex
)

func NewLogger(pkg string) *logrus.Entry {
	logger := logrus.New()

	loggersLock.Lock()
	loggers = append(loggers, logger)
	loggersLock.Unlock()

	return logger.WithField("package", pkg)
}

func SetLevel(level logrus.Level) {
	loggersLock.Lock()
	for _, logger := range loggers {
		logger.Level = level
	}
	loggersLock.Unlock()
}

func SetOutput(out io.Writer) {
	loggersLock.Lock()
	for _, logger := range loggers {
		logger.Out = out
	}
	loggersLock.Unlock()
}
