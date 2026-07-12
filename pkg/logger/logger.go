package logger

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
)

type loggerInCtx string

const (
	loggerKey loggerInCtx = "logger"
)

// logFormatEnv selects the output encoder at logger construction time. "text"
// yields the human-friendly formatter for local dev; anything else (default)
// yields structured JSON for production log aggregation.
const logFormatEnv = "LOG_FORMAT"

// logLevelEnv sets the default level for every constructed logger so that
// prefixed loggers (workers, reconcilers) inherit the configured level instead
// of silently defaulting to Info. Callers may still override via SetLevel.
const logLevelEnv = "LOG_LEVEL"

// Structured field keys. Correlation values travel on the context and are
// attached to every context-aware log call via withFields.
type contextKey string

const (
	requestIDKey contextKey = "request_id"
	userIDKey    contextKey = "user_id"
	orgIDKey     contextKey = "org_id"
	clusterIDKey contextKey = "cluster_id"
)

// Field name constants for structured logging. Callers attach these via
// WithField/WithFields instead of interpolating IDs into message strings.
const (
	FieldRequestID = "request_id"
	FieldUserID    = "user_id"
	FieldOrgID     = "org_id"
	FieldClusterID = "cluster_id"
	FieldStackID   = "stack_id"
	FieldReleaseID = "release_id"
	FieldResource  = "resource"
	FieldComponent = "component"
)

func AddLoggerToContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func GetLoggerFromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(loggerKey).(Logger); ok {
		return logger
	}
	return NewLogger()
}

// WithRequestID stamps a request/correlation ID onto the context so every
// context-aware log call made downstream carries it as a structured field.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func WithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

func WithOrgID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, orgIDKey, id)
}

func WithClusterID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, clusterIDKey, id)
}

func RequestIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDKey).(string)
	return id, ok
}

//go:generate mockgen -source=logger.go -destination=../mocks/mock_logger.go -package=mocks
type Logger interface {
	SetLevel(level logrus.Level) Logger

	GetLevel() logrus.Level

	// WithField returns a child logger that attaches the given key/value to
	// every subsequent log entry. The parent is left unchanged.
	WithField(key string, value interface{}) Logger

	// WithFields returns a child logger that attaches all the given fields to
	// every subsequent log entry. The parent is left unchanged.
	WithFields(fields map[string]interface{}) Logger

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

	// Error sends to the log an error message formatted using the fmt.Sprintf function and
	// the given format and arguments.
	Error(ctx context.Context, format string, args ...interface{})

	Errorf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Warnf(format string, args ...interface{})

	// Fatal sends to the log an error message formatted using the fmt.Sprintf function and
	// the given format and arguments; and then executes an os.Exit(1)
	// Fatal level is always enabled
	Fatal(ctx context.Context, format string, args ...interface{})
}

var _ Logger = &appLogger{}

type appLogger struct {
	debug  bool
	prefix string
	fields logrus.Fields
	logger *logrus.Logger
}

// newFormatter picks the encoder based on the LOG_FORMAT env var. JSON is the
// default so production defaults to structured logs without extra config.
func newFormatter() logrus.Formatter {
	if os.Getenv(logFormatEnv) == "text" {
		return &logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			DisableQuote:    true,
		}
	}
	return &logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "ts",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "msg",
		},
	}
}

// newLevel resolves the default log level from LOG_LEVEL, falling back to Info
// when unset or unparseable.
func newLevel() logrus.Level {
	if lvl, err := logrus.ParseLevel(os.Getenv(logLevelEnv)); err == nil {
		return lvl
	}
	return logrus.InfoLevel
}

// NewLogger creates a new logger instance with standard configuration
func NewLogger() Logger {
	l := logrus.New()
	l.SetFormatter(newFormatter())
	l.SetLevel(newLevel())
	return &appLogger{
		prefix: "",
		logger: l,
	}
}

func NewLoggerWithPrefix(ctx context.Context, prefix string) Logger {
	l := logrus.New()
	l.SetFormatter(newFormatter())
	l.SetLevel(newLevel())
	return &appLogger{
		prefix: prefix,
		logger: l,
	}
}

func NewLoggerWithDebug(ctx context.Context) Logger {
	l := logrus.New()
	l.SetLevel(logrus.DebugLevel)
	l.SetFormatter(newFormatter())
	return &appLogger{
		debug:  true,
		prefix: "",
		logger: l,
	}
}

// baseFields returns the fields attached to every entry regardless of context:
// the component prefix plus any fields bound via WithField/WithFields.
func (l *appLogger) baseFields() logrus.Fields {
	fields := logrus.Fields{}
	if l.prefix != "" {
		fields[FieldComponent] = l.prefix
	}
	for k, v := range l.fields {
		fields[k] = v
	}
	return fields
}

// withFields merges the base fields with correlation values carried on the
// context (request/user/org/cluster IDs) for a single log entry.
func (l *appLogger) withFields(ctx context.Context) *logrus.Entry {
	fields := l.baseFields()
	if ctx != nil {
		if v, ok := ctx.Value(requestIDKey).(string); ok && v != "" {
			fields[FieldRequestID] = v
		}
		if v, ok := ctx.Value(userIDKey).(string); ok && v != "" {
			fields[FieldUserID] = v
		}
		if v, ok := ctx.Value(orgIDKey).(string); ok && v != "" {
			fields[FieldOrgID] = v
		}
		if v, ok := ctx.Value(clusterIDKey).(string); ok && v != "" {
			fields[FieldClusterID] = v
		}
	}
	return l.logger.WithFields(fields)
}

func (l *appLogger) WithField(key string, value interface{}) Logger {
	return l.WithFields(map[string]interface{}{key: value})
}

func (l *appLogger) WithFields(fields map[string]interface{}) Logger {
	merged := logrus.Fields{}
	for k, v := range l.fields {
		merged[k] = v
	}
	for k, v := range fields {
		merged[k] = v
	}
	return &appLogger{
		debug:  l.debug,
		prefix: l.prefix,
		fields: merged,
		logger: l.logger,
	}
}

func (l *appLogger) SetLevel(level logrus.Level) Logger {
	l.logger.SetLevel(level)
	return l
}

func (l *appLogger) GetLevel() logrus.Level {
	return l.logger.GetLevel()
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
	l.logger.WithFields(l.baseFields()).Infof(format, args...)
}

func (l *appLogger) Warnf(format string, args ...interface{}) {
	l.logger.WithFields(l.baseFields()).Warnf(format, args...)
}

func (l *appLogger) Errorf(format string, args ...interface{}) {
	l.logger.WithFields(l.baseFields()).Errorf(format, args...)
}

func (l *appLogger) Debugf(format string, args ...interface{}) {
	l.logger.WithFields(l.baseFields()).Debugf(format, args...)
}

func (l *appLogger) Fatalf(format string, args ...interface{}) {
	l.logger.WithFields(l.baseFields()).Fatalf(format, args...)
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
