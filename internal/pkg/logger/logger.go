package logger

import (
	"io"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	instance *logrus.Logger
	once     sync.Once
)

// Get returns the singleton logger instance
func Get() *logrus.Logger {
	once.Do(func() {
		instance = logrus.New()
		instance.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
			PrettyPrint:     false,
		})
		instance.SetOutput(os.Stdout)
		instance.SetLevel(logrus.InfoLevel)
	})
	return instance
}

// Init initializes the logger with custom configuration
// This should be called once at application startup
func Init(level logrus.Level, output io.Writer) {
	logger := Get()
	logger.SetLevel(level)
	if output != nil {
		logger.SetOutput(output)
	}
}

// SetLevel sets the logging level
func SetLevel(level logrus.Level) {
	Get().SetLevel(level)
}

// Convenience methods that use the singleton instance

func Debug(args ...interface{}) {
	Get().Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	Get().Debugf(format, args...)
}

func Info(args ...interface{}) {
	Get().Info(args...)
}

func Infof(format string, args ...interface{}) {
	Get().Infof(format, args...)
}

func Warn(args ...interface{}) {
	Get().Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	Get().Warnf(format, args...)
}

func Error(args ...interface{}) {
	Get().Error(args...)
}

func Errorf(format string, args ...interface{}) {
	Get().Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	Get().Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	Get().Fatalf(format, args...)
}

func WithField(key string, value interface{}) *logrus.Entry {
	return Get().WithField(key, value)
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return Get().WithFields(fields)
}

func WithError(err error) *logrus.Entry {
	return Get().WithError(err)
}
