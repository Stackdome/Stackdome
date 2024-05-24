package logger

import (
	"context"

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
	context context.Context
}

// NewOCMLogger creates a new logger instance with a default verbosity of 1
func NewLogger(ctx context.Context) Logger {
	logger := &logger{
		context: ctx,
	}
	return logger
}

func (l *logger) Infof(format string, args ...interface{}) {
	glog.Infof(format, args...)
}

func (l *logger) Info(message string) {
	glog.Infof(message)
}

func (l *logger) Warning(message string) {
	glog.Warningf(message)
}

func (l *logger) Error(message string) {
	glog.Errorf(message)
}
func (l *logger) Errorf(message string, args ...any) {
	glog.Errorf(message, args...)
}

func (l *logger) Fatal(message string) {
	glog.Fatalf(message)
}
