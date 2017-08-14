package utils

import (
	"fmt"

	"github.com/Sirupsen/logrus"
)

const moduleLogField = "module"
const moduleSeparator = ":"

type Logger interface {
	New(module string) Logger
	Module() string

	SetLevel(level logrus.Level)
	SetFormatter(formatter logrus.Formatter)

	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Panic(args ...interface{})
	Panicf(format string, args ...interface{})
}

type StdLogger struct {
	innerLogger *logrus.Logger
	logFields   map[string]interface{}
}

var _ Logger = &StdLogger{}

func NewStdLogger(module string) *StdLogger {
	var l = &StdLogger{
		innerLogger: logrus.New(),
	}

	l.logFields = make(map[string]interface{})
	l.logFields[moduleLogField] = module + moduleSeparator
	return l
}

func (l *StdLogger) New(module string) Logger {
	newLogger := NewStdLogger(l.Module() + module)
	newLogger.innerLogger.Level = l.innerLogger.Level
	newLogger.innerLogger.Formatter = l.innerLogger.Formatter
	return newLogger
}

func (l *StdLogger) Module() string {
	return l.logFields[moduleLogField].(string)
}

func (l *StdLogger) SetLevel(level logrus.Level) {
	l.innerLogger.Level = level
}

func (l *StdLogger) SetFormatter(formatter logrus.Formatter) {
	l.innerLogger.Formatter = formatter
}

func (l *StdLogger) AddField(key string, value interface{}) {
	if key == moduleLogField {
		panic(fmt.Sprintf("Cannot add field with name %q", key))
	}
	l.logFields[key] = value
}

func (l *StdLogger) AddFields(fields map[string]interface{}) {
	for key, val := range fields {
		l.AddField(key, val)
	}
}

func (l *StdLogger) GetField(key string) (value interface{}, ok bool) {
	value, ok = l.logFields[key]
	return value, ok
}

func (l *StdLogger) Debug(args ...interface{}) {
	l.innerLogger.WithFields(l.logFields).Debug(args...)
}
func (l *StdLogger) Debugf(format string, args ...interface{}) {
	l.innerLogger.WithFields(l.logFields).Debugf(format, args...)
}
func (l *StdLogger) Info(args ...interface{}) {
	l.innerLogger.WithFields(l.logFields).Info(args...)
}
func (l *StdLogger) Infof(format string, args ...interface{}) {
	l.innerLogger.WithFields(l.logFields).Infof(format, args...)
}
func (l *StdLogger) Warn(args ...interface{}) {
	l.innerLogger.WithFields(l.logFields).Warn(args...)
}
func (l *StdLogger) Warnf(format string, args ...interface{}) {
	l.innerLogger.WithFields(l.logFields).Warnf(format, args...)
}
func (l *StdLogger) Error(args ...interface{}) {
	l.innerLogger.WithFields(l.logFields).Error(args...)
}
func (l *StdLogger) Errorf(format string, args ...interface{}) {
	l.innerLogger.WithFields(l.logFields).Errorf(format, args...)
}
func (l *StdLogger) Panic(args ...interface{}) {
	l.innerLogger.WithFields(l.logFields).Panic(args...)
}
func (l *StdLogger) Panicf(format string, args ...interface{}) {
	l.innerLogger.WithFields(l.logFields).Panicf(format, args...)
}

type NoLogger struct{}

var _ Logger = &NoLogger{}

func NewNoLogger() *NoLogger {
	return &NoLogger{}
}

func (l *NoLogger) New(module string) Logger {
	return l
}
func (l *NoLogger) Module() string {
	return moduleSeparator
}

func (l *NoLogger) SetLevel(level logrus.Level)             {}
func (l *NoLogger) SetFormatter(formatter logrus.Formatter) {}

func (l *NoLogger) Debug(args ...interface{})                 {}
func (l *NoLogger) Debugf(format string, args ...interface{}) {}
func (l *NoLogger) Info(args ...interface{})                  {}
func (l *NoLogger) Infof(format string, args ...interface{})  {}
func (l *NoLogger) Warn(args ...interface{})                  {}
func (l *NoLogger) Warnf(format string, args ...interface{})  {}
func (l *NoLogger) Error(args ...interface{})                 {}
func (l *NoLogger) Errorf(format string, args ...interface{}) {}
func (l *NoLogger) Panic(args ...interface{}) {
	panic(args)
}
func (l *NoLogger) Panicf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args))
}
