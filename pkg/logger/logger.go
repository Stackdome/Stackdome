package logger

import (
	"context"

	"github.com/sirupsen/logrus"
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
	Debugf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Warnf(format string, args ...interface{})

	// Fatal sends to the log an error message formatted using the fmt.Sprintf function and the
	// given format and arguments; and then executes an os.Exit(1)
	// Fatal level is always enabled
	Fatal(ctx context.Context, format string, args ...interface{})
}

var _ Logger = &appLogger{}

type appLogger struct {
	debug  bool
	prefix string
	logger *logrus.Logger
}

// NewLogger creates a new logger instance with standard configuration
func NewLogger() Logger {
	l := logrus.New()
	// Set default formatter with timestamps and full log level names
	l.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableQuote:    true,
	})
	return &appLogger{
		prefix: "",
		logger: l,
	}
}

func NewLoggerWithPrefix(ctx context.Context, prefix string) Logger {
	l := logrus.New()
	l.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableQuote:    true,
	})
	return &appLogger{
		prefix: prefix,
		logger: l,
	}
}

func NewLoggerWithDebug(ctx context.Context) Logger {
	l := logrus.New()
	l.SetLevel(logrus.DebugLevel)
	l.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableQuote:    true,
	})
	return &appLogger{
		debug:  true,
		prefix: "",
		logger: l,
	}
}

// withFields adds consistent fields to all log entries
func (l *appLogger) withFields(ctx context.Context) *logrus.Entry {
	fields := logrus.Fields{}
	if l.prefix != "" {
		fields["component"] = l.prefix
	}
	// You can add more context-based fields here, like:
	// - Request ID from context
	// - Correlation ID
	// - User info
	return l.logger.WithFields(fields)
}

func (l *appLogger) DebugEnabled() bool {
	return l.logger.IsLevelEnabled(logrus.DebugLevel)
}

func (l *appLogger) InfoEnabled() bool {
	return l.logger.IsLevelEnabled(logrus.InfoLevel)
}

func (l *appLogger) WarnEnabled() bool {
	return l.logger.IsLevelEnabled(logrus.WarnLevel)
}

func (l *appLogger) ErrorEnabled() bool {
	return l.logger.IsLevelEnabled(logrus.ErrorLevel)
}

func (l *appLogger) Debug(ctx context.Context, format string, args ...interface{}) {
	if l.DebugEnabled() {
		l.withFields(ctx).Debugf(format, args...)
	}
}

func (l *appLogger) Infof(format string, args ...interface{}) {
	// For backwards compatibility with methods that don't have context
	l.logger.WithFields(logrus.Fields{
		"component": l.prefix,
	}).Infof(format, args...)
}

func (l *appLogger) Warnf(format string, args ...interface{}) {
	// For backwards compatibility with methods that don't have context
	l.logger.WithFields(logrus.Fields{
		"component": l.prefix,
	}).Warnf(format, args...)
}

func (l *appLogger) Errorf(format string, args ...interface{}) {
	// For backwards compatibility with methods that don't have context
	l.logger.WithFields(logrus.Fields{
		"component": l.prefix,
	}).Errorf(format, args...)
}

func (l *appLogger) Debugf(format string, args ...interface{}) {
	// For backwards compatibility with methods that don't have context
	l.logger.WithFields(logrus.Fields{
		"component": l.prefix,
	}).Debugf(format, args...)
}

func (l *appLogger) Fatalf(format string, args ...interface{}) {
	l.logger.WithFields(logrus.Fields{
		"component": l.prefix,
	}).Fatalf(format, args...)
}

func (l *appLogger) Info(ctx context.Context, format string, args ...interface{}) {
	l.withFields(ctx).Infof(format, args...)
}

func (l *appLogger) Warn(ctx context.Context, format string, args ...interface{}) {
	l.withFields(ctx).Warnf(format, args...)
}

func (l *appLogger) Error(ctx context.Context, format string, args ...interface{}) {
	l.withFields(ctx).Errorf(format, args...)
}

func (l *appLogger) Fatal(ctx context.Context, format string, args ...interface{}) {
	l.withFields(ctx).Fatalf(format, args...)
}
