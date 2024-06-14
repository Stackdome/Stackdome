package logger

import (
	"context"
	"fmt"

	"github.com/golang/glog"
)

type Logger interface {
	Infof(format string, args ...interface{})
	Info(message string)
	Warning(message string)
	Error(message string)
	Errorf(message string, args ...any)
	Fatal(message string)
}

var _ Logger = &logger{}

type logger struct {
	prefix  string
	context context.Context
}

// NewOCMLogger creates a new logger instance with a default verbosity of 1
func NewLogger(ctx context.Context) Logger {
	logger := &logger{
		context: ctx,
		prefix:  "",
	}
	return logger
}

func NewLoggerWithPrefix(ctx context.Context, prefix string) Logger {
	logger := &logger{
		context: ctx,
		prefix:  prefix,
	}
	return logger
}

func (l *logger) WithPrefix(input string) string {
	if len(l.prefix) > 0 {
		return fmt.Sprintf("%s: %s", l.prefix, input)
	}
	return input
}

func (l *logger) Infof(message string, args ...interface{}) {
	glog.Infof(l.WithPrefix(message), args...)
}

func (l *logger) Info(message string) {
	glog.Infof(l.WithPrefix(message))
}

func (l *logger) Warning(message string) {
	glog.Warningf(l.WithPrefix(message))
}

func (l *logger) Error(message string) {
	glog.Errorf(l.WithPrefix(message))
}
func (l *logger) Errorf(message string, args ...any) {
	glog.Errorf(l.WithPrefix(message), args...)
}

func (l *logger) Fatal(message string) {
	glog.Fatalf(l.WithPrefix(message))
}
