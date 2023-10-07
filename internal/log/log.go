package log

import (
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

type logger interface {
	Println(args ...interface{})
	Error(args ...interface{})
	Debug(args ...interface{})
	Printf(format string, args ...interface{})
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
	//TODO: make this also configurable with lsp configs
	// Check the value of the environment variable
	if os.Getenv("LOG_LEVEL") == "debug" {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}
	return logger
}
