package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Debug(args ...interface{})
	Warn(args ...interface{})
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
}

type logger struct {
	*logrus.Logger
}

func New() Logger {
	l := logrus.New()
	
	// Set output to stdout
	l.SetOutput(os.Stdout)
	
	// Set log level
	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "debug":
		l.SetLevel(logrus.DebugLevel)
	case "info":
		l.SetLevel(logrus.InfoLevel)
	case "warn":
		l.SetLevel(logrus.WarnLevel)
	case "error":
		l.SetLevel(logrus.ErrorLevel)
	default:
		l.SetLevel(logrus.InfoLevel)
	}
	
	// Set formatter
	l.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})
	
	return &logger{l}
}

func (l *logger) WithField(key string, value interface{}) Logger {
	return &logger{l.Logger.WithField(key, value).Logger}
}

func (l *logger) WithFields(fields map[string]interface{}) Logger {
	return &logger{l.Logger.WithFields(logrus.Fields(fields)).Logger}
} 