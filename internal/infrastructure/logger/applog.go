package logger

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// LogType is the log entry category per standard-app-log v1.0.
type LogType string

const (
	LogTypeApp   LogType = "APP_LOG"
	LogTypeReq   LogType = "REQ_LOG"
	LogTypeRes   LogType = "RES_LOG"
	LogTypeReqEx LogType = "REQ_EX_LOG"
	LogTypeResEx LogType = "RES_EX_LOG"
	LogTypePII   LogType = "PII_LOG"
)

// LogLevel is the severity of a log entry per standard-app-log v1.0.
type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
)

// Field is a key-value pair appended to a log entry.
type Field struct {
	key   string
	value interface{}
}

// F creates a generic log field.
func F(key string, value interface{}) Field { return Field{key: key, value: value} }

// FError creates a field for an error value.
func FError(err error) Field { return Field{key: "error", value: err.Error()} }

// FDurationMs converts a duration to milliseconds and stores it as execution_time.
func FDurationMs(d time.Duration) Field {
	return Field{key: "execution_time", value: d.Milliseconds()}
}

// logEntry is the JSON payload emitted per standard-app-log v1.0.
type logEntry struct {
	EventDateTime  string   `json:"event_date_time"`
	LogType        LogType  `json:"log_type"`
	Level          LogLevel `json:"level"`
	AppID          string   `json:"app_id,omitempty"`
	AppVersion     string   `json:"app_version,omitempty"`
	ServiceID      string   `json:"service_id,omitempty"`
	ServicePodName string   `json:"service_pod_name,omitempty"`
	Message        string   `json:"message,omitempty"`
	CorrelationID  string   `json:"correlation_id,omitempty"`
	RequestID      string   `json:"request_id,omitempty"`
	TraceID        string   `json:"trace_id,omitempty"`
	SpanID         string   `json:"span_id,omitempty"`
	ExecutionTime  *int64   `json:"execution_time,omitempty"` // milliseconds
	CallerUser     string   `json:"caller_user,omitempty"`
	CallerAddress  string   `json:"caller_address,omitempty"`
	CodeLocation   string   `json:"code_location,omitempty"`
	GeoLocation    string   `json:"geo_location,omitempty"`
}

// StandardLogger emits one JSON line per call, conforming to standard-app-log v1.0.
// It is safe for concurrent use. Use With* methods to derive child loggers scoped
// to a correlation, request, or trace span without mutating the parent.
type StandardLogger struct {
	appID          string
	appVersion     string
	serviceID      string
	servicePodName string
	correlationID  string
	requestID      string
	traceID        string
	spanID         string
	output         io.Writer
	minLevel       LogLevel
	nop            bool
}

// StandardLoggerOption is a functional option for NewStandardLogger.
type StandardLoggerOption func(*StandardLogger)

func WithAppID(id string) StandardLoggerOption { return func(l *StandardLogger) { l.appID = id } }
func WithAppVersion(v string) StandardLoggerOption {
	return func(l *StandardLogger) { l.appVersion = v }
}
func WithServiceID(id string) StandardLoggerOption {
	return func(l *StandardLogger) { l.serviceID = id }
}
func WithPodName(n string) StandardLoggerOption {
	return func(l *StandardLogger) { l.servicePodName = n }
}
func WithOutput(w io.Writer) StandardLoggerOption { return func(l *StandardLogger) { l.output = w } }

// WithMinLevel sets the minimum log level. Calls below this level are silently dropped.
// Defaults to LogLevelInfo when not set.
func WithMinLevel(level LogLevel) StandardLoggerOption {
	return func(l *StandardLogger) { l.minLevel = level }
}

// NewStandardLogger creates a StandardLogger that writes JSON to stdout.
// The default minimum level is INFO; use WithMinLevel(LogLevelDebug) for verbose output.
//
// Example:
//
//	log := logger.NewStandardLogger(
//	    logger.WithAppID("go-nixcopy"),
//	    logger.WithAppVersion("1.2.0"),
//	    logger.WithServiceID("sftp-to-s3"),
//	)
func NewStandardLogger(opts ...StandardLoggerOption) *StandardLogger {
	l := &StandardLogger{output: os.Stdout, minLevel: LogLevelInfo}
	for _, o := range opts {
		o(l)
	}
	return l
}

// NewNopLogger returns a logger that discards all output (for use in tests).
func NewNopLogger() *StandardLogger { return &StandardLogger{nop: true} }

// GenerateID returns a random UUID v4 string using crypto/rand.
func GenerateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// WithCorrelation returns a child logger scoped to a correlation ID and request ID pair.
func (l *StandardLogger) WithCorrelation(correlationID, requestID string) *StandardLogger {
	c := *l
	c.correlationID = correlationID
	c.requestID = requestID
	return &c
}

// WithRequest returns a child logger with a new request_id, preserving the correlation_id.
func (l *StandardLogger) WithRequest(requestID string) *StandardLogger {
	c := *l
	c.requestID = requestID
	return &c
}

// WithTrace returns a child logger with a trace_id and span_id.
func (l *StandardLogger) WithTrace(traceID, spanID string) *StandardLogger {
	c := *l
	c.traceID = traceID
	c.spanID = spanID
	return &c
}

// Info emits APP_LOG INFO.
func (l *StandardLogger) Info(msg string, fields ...Field) {
	l.emit(LogTypeApp, LogLevelInfo, msg, fields)
}

// Warn emits APP_LOG WARN.
func (l *StandardLogger) Warn(msg string, fields ...Field) {
	l.emit(LogTypeApp, LogLevelWarn, msg, fields)
}

// Error emits APP_LOG ERROR.
func (l *StandardLogger) Error(msg string, fields ...Field) {
	l.emit(LogTypeApp, LogLevelError, msg, fields)
}

// Debug emits APP_LOG DEBUG.
func (l *StandardLogger) Debug(msg string, fields ...Field) {
	l.emit(LogTypeApp, LogLevelDebug, msg, fields)
}

// InfoReqEx emits REQ_EX_LOG INFO — use when initiating a call to an external service.
func (l *StandardLogger) InfoReqEx(msg string, fields ...Field) {
	l.emit(LogTypeReqEx, LogLevelInfo, msg, fields)
}

// InfoResEx emits RES_EX_LOG INFO — use when a response is received from an external service.
func (l *StandardLogger) InfoResEx(msg string, fields ...Field) {
	l.emit(LogTypeResEx, LogLevelInfo, msg, fields)
}

// WarnResEx emits RES_EX_LOG WARN.
func (l *StandardLogger) WarnResEx(msg string, fields ...Field) {
	l.emit(LogTypeResEx, LogLevelWarn, msg, fields)
}

// ErrorReqEx emits REQ_EX_LOG ERROR.
func (l *StandardLogger) ErrorReqEx(msg string, fields ...Field) {
	l.emit(LogTypeReqEx, LogLevelError, msg, fields)
}

// ErrorResEx emits RES_EX_LOG ERROR.
func (l *StandardLogger) ErrorResEx(msg string, fields ...Field) {
	l.emit(LogTypeResEx, LogLevelError, msg, fields)
}

// levelOrder maps a LogLevel to a numeric rank for comparison (lower = less severe).
func levelOrder(level LogLevel) int {
	switch level {
	case LogLevelDebug:
		return 0
	case LogLevelInfo:
		return 1
	case LogLevelWarn:
		return 2
	case LogLevelError:
		return 3
	default:
		return 1
	}
}

func (l *StandardLogger) emit(logType LogType, level LogLevel, msg string, fields []Field) {
	if l.nop {
		return
	}
	if levelOrder(level) < levelOrder(l.minLevel) {
		return
	}

	entry := logEntry{
		EventDateTime:  time.Now().UTC().Format(time.RFC3339Nano),
		LogType:        logType,
		Level:          level,
		AppID:          l.appID,
		AppVersion:     l.appVersion,
		ServiceID:      l.serviceID,
		ServicePodName: l.servicePodName,
		Message:        msg,
		CorrelationID:  l.correlationID,
		RequestID:      l.requestID,
		TraceID:        l.traceID,
		SpanID:         l.spanID,
	}

	// Separate standard optional fields from app-specific extra fields.
	extra := make(map[string]interface{})
	for _, f := range fields {
		switch f.key {
		case "execution_time":
			if v, ok := f.value.(int64); ok {
				entry.ExecutionTime = &v
			}
		case "caller_user":
			entry.CallerUser = fmt.Sprintf("%v", f.value)
		case "caller_address":
			entry.CallerAddress = fmt.Sprintf("%v", f.value)
		case "code_location":
			entry.CodeLocation = fmt.Sprintf("%v", f.value)
		case "geo_location":
			entry.GeoLocation = fmt.Sprintf("%v", f.value)
		default:
			extra[f.key] = f.value
		}
	}

	data, err := marshalWithExtra(entry, extra)
	if err != nil {
		return
	}
	_, _ = fmt.Fprintln(l.output, string(data))
}

// marshalWithExtra serialises entry then merges app-specific fields at the top level.
func marshalWithExtra(entry logEntry, extra map[string]interface{}) ([]byte, error) {
	if len(extra) == 0 {
		return json.Marshal(entry)
	}
	base, err := json.Marshal(entry)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(base, &m); err != nil {
		return nil, err
	}
	for k, v := range extra {
		m[k] = v
	}
	return json.Marshal(m)
}
