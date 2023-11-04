package log

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type logger interface {
	Println(args ...interface{})
	Error(args ...interface{})
	Debug(args ...interface{})
	Printf(format string, args ...interface{})
	SetLevel(level logrus.Level)
}

var l logger
var once sync.Once

// start a new logger
func GetLogger() logger {
	once.Do(func() {
		l = createLogger()
	})
	return l
}

func createLogger() logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	return logger
}
