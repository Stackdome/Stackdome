package logger

import (
	"context"
	"fmt"

	"github.com/golang/glog"
)

type Logger interface {
	// DebugEnabled returns true if the debug level is enabled.
	DebugEnabled() bool

	// InfoEnabled returns true if the information level is enabled.
	InfoEnabled() bool

	// WarnEnabled returns true if the warning level is enabled.
	WarnEnabled() bool

	// ErrorEnabled returns true if the error level is enabled.
	ErrorEnabled() bool

	// Debug sends to the log a debug message formatted using the fmt.Sprintf function and the
	// given format and arguments.
	Debug(ctx context.Context, format string, args ...interface{})

	// Info sends to the log an information message formatted using the fmt.Sprintf function and
	// the given format and arguments.
	Infof(format string, args ...interface{})

	Info(ctx context.Context, format string, args ...interface{})

	// Warn sends to the log a warning message formatted using the fmt.Sprintf function and the
	// given format and arguments.
	Warn(ctx context.Context, format string, args ...interface{})

	// Error sends to the log an error message formatted using the fmt.Sprintf function and the
	// given format and arguments.
	Error(ctx context.Context, format string, args ...interface{})

	Errorf(format string, args ...interface{})

	// Fatal sends to the log an error message formatted using the fmt.Sprintf function and the
	// given format and arguments; and then executes an os.Exit(1)
	// Fatal level is always enabled
	Fatal(ctx context.Context, format string, args ...interface{})
}

var _ Logger = &logger{}

type logger struct {
	debug   bool
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

func NewLoggerWithDebug(ctx context.Context) Logger {
	logger := &logger{
		context: ctx,
		debug:   true,
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

func (l *logger) Info(ctx context.Context, message string, args ...any) {
	glog.Infof(l.WithPrefix(message), args...)
}

func (l *logger) Warning(message string) {
	glog.Warningf(l.WithPrefix(message))
}

func (l *logger) Warnf(message string, args ...any) {
	glog.Warningf(l.WithPrefix(message), args...)
}

func (l *logger) Error(ctx context.Context, message string, args ...any) {
	glog.Errorf(l.WithPrefix(message), args...)
}

func (l *logger) Errorf(message string, args ...any) {
	glog.Errorf(l.WithPrefix(message), args...)
}

func (l *logger) Fatal(ctx context.Context, message string, args ...any) {
	glog.Fatalf(l.WithPrefix(message), args...)
}

func (l *logger) DebugEnabled() bool {
	return l.debug
}

func (l *logger) InfoEnabled() bool {
	return true
}

func (l *logger) WarnEnabled() bool {
	return true
}

func (l *logger) ErrorEnabled() bool {
	return true
}

func (l *logger) Debug(ctx context.Context, format string, args ...interface{}) {
	if l.debug {
		glog.Infof(l.WithPrefix(format), args...)
	}
}

func (l *logger) Warn(ctx context.Context, format string, args ...interface{}) {
	glog.Warningf(l.WithPrefix(format), args...)
}
