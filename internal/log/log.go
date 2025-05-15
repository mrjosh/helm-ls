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

var (
	l    logger
	once sync.Once
)

// start a new logger
func GetLogger() logger {
	once.Do(func() {
		l = createLogger()
	})
	return l
}

func createLogger() logger {
	logrus.SetReportCaller(true)
	logger := logrus.New()

	formatter := &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "@timestamp",
			logrus.FieldKeyLevel: "@level",
			logrus.FieldKeyMsg:   "@message",
			logrus.FieldKeyFunc:  "@caller",
		},
	}
	logger.SetFormatter(formatter)
	return logger
}
